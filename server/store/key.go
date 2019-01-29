package store

import (
	"bytes"
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	"time"
)

// KEY implement key command
type KEY struct {
	store   *Store
	environ *Environ
	// key
	key []byte

	// metaKey
	metaKey []byte

	// value saved in meta extra
	meta *Meta
}

// InitKey init
func InitKey(environ *Environ, store *Store) *KEY {
	pk := KEY{
		store:   store,
		environ: environ,
	}
	return &pk
}

func (pk *KEY) existsForRead(key []byte) error {
	mkey := EncodeMetaKey(pk.environ.Header, key)
	value, err := pk.store.Get(mkey)
	if err != nil {
		return err
	}
	if value == nil {
		return terror.ErrKeyNotValid
	}
	meta := DecodeMetaValue(value)
	pk.environ.SetType(meta.Type)
	if meta.CheckIfExpire() {
		pk.store.WriteReset()
		// remove from expire set
		if err = RemoveExpire(pk.environ, pk.store, key, meta.ExpireAt); err != nil {
			return err
		}
		if err = pk.store.Delete(mkey); err != nil {
			return err
		}
		if meta.Type != types.STRING {
			if err = AddToGc(pk.environ, pk.store, meta.ID); err != nil {
				return err
			}
		}
		if err = pk.store.Commit(); err != nil {
			pk.environ.FailedTxn("existsForRead")
			return err
		}
		return terror.ErrKeyNotValid
	}
	pk.meta = meta
	pk.metaKey = mkey
	pk.key = key
	return nil
}

// Prefix header|m
func (pk *KEY) Prefix() []byte {
	buffer := make([]byte, pk.environ.HeaderLen+1)
	copy(buffer[0:], pk.environ.Header)
	buffer[pk.environ.HeaderLen] = types.META
	return buffer
}

// EncodeDataKey header|m|key
func (pk *KEY) EncodeDataKey(key []byte) []byte {
	return EncodeMetaKey(pk.environ.Header, key)
}

// Del keys
func (pk *KEY) Del(keys ...[]byte) (uint64, error) {
	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = EncodeMetaKey(pk.environ.Header, key)
	}
	result, err := pk.store.BatchGet(mkeys)
	if err != nil || result == nil {
		return 0, nil
	}
	var delCnt uint64
	for key, value := range result {
		meta := DecodeMetaValue(value)
		pk.environ.SetType(meta.Type)
		if err = pk.store.Delete([]byte(key)); err != nil {
			return 0, err
		}
		if meta.Type != types.STRING {
			// add to gc
			if err = AddToGc(pk.environ, pk.store, meta.ID); err != nil {
				return 0, err
			}
		}
		if meta.ExpireAt > 0 {
			if err := RemoveExpire(pk.environ, pk.store, []byte(key), meta.ExpireAt); err != nil {
				return 0, err
			}
		}
		delCnt++
	}
	if err := pk.store.Commit(); err != nil {
		pk.environ.FailedTxn("Del")
		return 0, nil
	}
	return delCnt, nil
}

// Exists keys
func (pk *KEY) Exists(keys ...[]byte) (uint64, error) {
	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = EncodeMetaKey(pk.environ.Header, key)
	}
	result, err := pk.store.BatchGet(mkeys)
	if err != nil || result == nil {
		return 0, nil
	}
	var total uint64
	for _, val := range result {
		if val != nil {
			meta := DecodeMetaValue(val)
			if meta.CheckIfExpire() {
				continue
			}
			total++
		}
	}
	return total, nil
}

// Expire keys
func (pk *KEY) Expire(key []byte, expireAt int64) (uint8, error) {
	if err := pk.existsForRead(key); err != nil || pk.meta == nil {
		return 0, nil
	}
	if expireAt < 0 || expireAt < time.Now().Unix() {
		// del direct
		if err := pk.store.Delete(pk.metaKey); err != nil {
			return 0, err
		}
		if pk.meta.Type != types.STRING {
			if err := AddToGc(pk.environ, pk.store, pk.meta.ID); err != nil {
				return 0, err
			}
		}
		if pk.meta.ExpireAt > 0 {
			if err := RemoveExpire(pk.environ, pk.store, key, pk.meta.ExpireAt); err != nil {
				return 0, err
			}
		}
	} else {
		var oldExpireAt uint64
		if pk.meta.ExpireAt > 0 {
			oldExpireAt = pk.meta.ExpireAt
		}
		pk.meta.ExpireAt = uint64(expireAt)
		if err := pk.store.Set(pk.metaKey, pk.meta.EncodeMetaValue()); err != nil {
			return 0, err
		}
		if err := AddExpire(pk.environ, pk.store, key, pk.meta.ID, pk.meta.ExpireAt, oldExpireAt); err != nil {
			return 0, err
		}
	}
	if err := pk.store.Commit(); err != nil {
		pk.environ.FailedTxn("Expire")
		return 0, err
	}
	return 1, nil
}

// TTL keys
func (pk *KEY) TTL(key []byte) (int64, error) {
	if err := pk.existsForRead(key); err != nil || pk.meta == nil {
		return -2, nil
	}
	if pk.meta.ExpireAt == 0 {
		return -1, nil
	}
	ttl := int64(pk.meta.ExpireAt) - time.Now().UnixNano()/1000/1000
	return ttl, nil
}

// Type key
func (pk *KEY) Type(key []byte) ([]byte, error) {
	if err := pk.existsForRead(key); err != nil || pk.meta == nil {
		return nil, nil
	}
	return []byte(pk.meta.StringType()), nil
}

// Keys return keys
func (pk *KEY) Keys(key []byte, start uint64, limit uint64) ([][]byte, error) {
	var prefix []byte
	if string(key) == "*" {
		prefix = pk.Prefix()
	} else {
		prefix = EncodeMetaKey(pk.environ.Header, bytes.TrimRight(key, "*"))
	}
	if limit == 0 || limit > 5000 {
		limit = 5000
	}
	result, err := pk.store.ScanKeys(prefix, nil, start, limit)
	if err != nil || result == nil {
		return nil, nil
	}
	keys := make([][]byte, len(result))
	for i, key := range result {
		keys[i] = DecodeMetaKey(pk.environ.Header, key)
	}
	return keys, nil
}
