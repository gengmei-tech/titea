package store

import (
	"bytes"
	"github.com/gengmei-tech/titea/pkg/util/number"
	"github.com/gengmei-tech/titea/server/types"
	"time"
)

// SLEEPTIME interval before next loop
const SLEEPTIME = 5 * time.Second

// StartExpireGc start expire and gc
func StartExpireGc(store *Store, namespaces []*types.Namespace) {
	for _, namespace := range namespaces {
		environ := CreateEnviron(namespace)
		startExpire(environ, store)
		startGc(environ, store)
	}
}

// SyncExpireGc when register a new namespace, it need syncExpireGc
func SyncExpireGc(store *Store, namespace *types.Namespace) {
	environ := CreateEnviron(namespace)
	startExpire(environ, store)
	startGc(environ, store)
}

func startExpire(environ *Environ, store *Store) {
	go func(environ *Environ, store *Store) {
		store = store.Read()
		for {
			runExpire(environ, store)
			time.Sleep(time.Duration(SLEEPTIME))
			store.SetRead()
		}
	}(environ, store)
}

func runExpire(environ *Environ, store *Store) {
	prefix := expirePrefix(environ.Header)
	iter, err := store.Seek(prefix)
	if err != nil {
		return
	}
	defer iter.Close()

	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
		key := iter.Key()

		value := iter.Value()
		rawKey, expireAt := decodeExpireKey(environ.Header, key)
		// expire delay 12h, for decrease txn conflict
		if expireAt > uint64((time.Now().Unix()-43200)*1000) {
			break
		}
		// check meta
		mkey := EncodeMetaKey(environ.Header, rawKey)
		mvalue, err := store.Get(mkey)
		if err != nil {
			continue
		}
		var exec = false
		var change = false
		if mvalue != nil {
			meta := DecodeMetaValue(mvalue)
			// have change from one object to another object
			if meta.ExpireAt == 0 || bytes.Compare(meta.ID, value[0:16]) != 0 || meta.Type != value[16] {
				exec = true
				change = true
			}
			if meta.CheckIfExpire() {
				exec = true
			}
		} else {
			exec = true
			change = true
		}

		iter.Next()
		environ.Type = value[16]

		if exec {
			store.WriteReset()
			if environ.Type != types.STRING {
				if err = AddToGc(environ, store, value[0:16]); err != nil {
					continue
				}
			}

			// delete meta key
			if !change {
				if err = store.Delete(mkey); err != nil {
					continue
				}
			}

			// delete expire key
			if err = store.Delete(key); err != nil {
				continue
			}
			if err := store.Commit(); err != nil {
				environ.FailedTxn("runExpire")
				continue
			}
			environ.ExecExpire()
		}
	}
}

// header|type(1)|expireAt|key
func encodeExpireKey(header []byte, key []byte, expireAt uint64) []byte {
	var buffer []byte
	buffer = append(buffer, header...)
	buffer = append(buffer, types.EXPIRE)
	buffer = append(buffer, number.Uint64ToBytes(expireAt)...)
	buffer = append(buffer, key...)
	return buffer
}

func decodeExpireKey(header []byte, key []byte) ([]byte, uint64) {
	key = bytes.TrimPrefix(key, header)
	expireAt := number.BytesToUint64(key[1:9])
	return key[9:], expireAt
}

// header|type(1)
func expirePrefix(header []byte) []byte {
	var buffer []byte
	buffer = append(buffer, header...)
	buffer = append(buffer, types.EXPIRE)
	return buffer
}

// AddExpire add key to expire
func AddExpire(environ *Environ, store *Store, key []byte, id []byte, expireAt uint64, oldExpireAt uint64) error {
	if oldExpireAt > 0 {
		oldKey := encodeExpireKey(environ.Header, key, oldExpireAt)
		if err := store.Delete(oldKey); err != nil {
			return err
		}
	}
	if expireAt > 0 {
		newKey := encodeExpireKey(environ.Header, key, expireAt)
		// value = ID|type, type for metrics
		value := make([]byte, 17)
		copy(value[0:], id)
		value[16] = environ.Type
		if err := store.Set(newKey, value); err != nil {
			return err
		}
	}
	return nil
}

// RemoveExpire remove key from expire
func RemoveExpire(environ *Environ, store *Store, key []byte, expireAt uint64) error {
	expireKey := encodeExpireKey(environ.Header, key, expireAt)
	return store.Delete(expireKey)
}
