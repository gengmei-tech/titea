package store

import (
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	"strconv"
)

// STRING implement redis string operation
type STRING struct {
	store *Store

	// environ
	environ *Environ

	// key
	key []byte

	// value
	value []byte

	// metaKey
	metaKey []byte

	// meta
	meta *Meta
}

// InitString init
func InitString(environ *Environ, store *Store) *STRING {
	st := STRING{
		store:   store,
		environ: environ,
	}
	return &st
}

// internal recall
func (st *STRING) existsForRead(key []byte) error {
	mkey := EncodeMetaKey(st.environ.Header, key)
	value, err := st.store.Get(mkey)
	if err != nil {
		return err
	}
	st.metaKey = mkey
	if value == nil {
		return terror.ErrKeyNotValid
	}
	meta := DecodeMetaValue(value)
	if !meta.CheckType(types.STRING) {
		return terror.ErrTypeNotMatch
	}
	if meta.CheckIfExpire() {
		st.store.WriteReset()
		// remove from expire set
		if err = RemoveExpire(st.environ, st.store, key, meta.ExpireAt); err != nil {
			return err
		}
		if err = st.store.Delete(mkey); err != nil {
			return err
		}
		if err = st.store.Commit(); err != nil {
			st.environ.FailedTxn("existsForRead")
			return err
		}
		return terror.ErrKeyNotValid
	}
	st.value = meta.Extra
	return nil
}

// internal recall
func (st *STRING) existsForWrite(key []byte) error {
	mkey := EncodeMetaKey(st.environ.Header, key)
	value, err := st.store.Get(mkey)
	if err != nil {
		return err
	}
	st.metaKey = mkey
	if value == nil {
		meta := NewMeta(types.STRING)
		st.key = key
		st.meta = meta
		st.value = nil
	} else {
		meta := DecodeMetaValue(value)
		if !meta.CheckType(types.STRING) {
			return terror.ErrTypeNotMatch
		}
		if meta.CheckIfExpire() {
			// remove from expire set
			if err = RemoveExpire(st.environ, st.store, key, meta.ExpireAt); err != nil {
				return err
			}
			meta.Reset()
			st.meta = meta
			st.value = meta.Extra
		} else {
			st.meta = meta
			st.value = meta.Extra
		}
	}
	return nil
}

// Set operation
// isNX set value when key not exists
// isXX set value when key exists
func (st *STRING) Set(key []byte, value []byte, expireAt uint64, isNX bool, isXX bool) error {
	if err := st.existsForWrite(key); err != nil {
		return err
	}
	if isNX && st.value != nil {
		return terror.ErrKeyExist
	}
	if isXX && st.value == nil {
		return terror.ErrKeyNotExist
	}

	// set expire before
	if st.meta.ExpireAt > 0 {
		if expireAt > 0 {
			// set expire before
			if err := AddExpire(st.environ, st.store, key, st.meta.ID, expireAt, st.meta.ExpireAt); err != nil {
				return err
			}
		} else {
			if err := RemoveExpire(st.environ, st.store, key, st.meta.ExpireAt); err != nil {
				return err
			}
		}
	} else if expireAt > 0 {
		if err := AddExpire(st.environ, st.store, key, st.meta.ID, expireAt, 0); err != nil {
			return err
		}
	}
	st.meta.Extra = value
	st.meta.ExpireAt = expireAt
	if err := st.store.Set(st.metaKey, st.meta.EncodeMetaValue()); err != nil {
		return err
	}
	if err := st.store.Commit(); err != nil {
		st.environ.FailedTxn("Set")
		return err
	}
	return nil
}

// GetSet operation
func (st *STRING) GetSet(key []byte, value []byte) ([]byte, error) {
	if err := st.existsForWrite(key); err != nil {
		return nil, err
	}
	st.meta.Extra = value
	if err := st.store.Set(st.metaKey, st.meta.EncodeMetaValue()); err != nil {
		return nil, err
	}
	if err := st.store.Commit(); err != nil {
		st.environ.FailedTxn("GetSet")
		return nil, err
	}
	return st.value, nil
}

// Get value
func (st *STRING) Get(key []byte) ([]byte, error) {
	if err := st.existsForRead(key); err != nil {
		return nil, nil
	}
	return st.value, nil
}

// MGet @tip if expire, no sync expire execute
func (st *STRING) MGet(keys ...[]byte) ([][]byte, error) {
	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = EncodeMetaKey(st.environ.Header, key)
	}
	result, err := st.store.BatchGet(mkeys)
	if err != nil {
		return nil, err
	}
	var values = make([][]byte, len(keys))
	var i = 0
	for _, mkey := range mkeys {
		if val, ok := result[string(mkey)]; ok {
			meta := DecodeMetaValue(val)
			if !meta.CheckType(types.STRING) {
				values[i] = nil
				i++
				continue
			}
			if meta.CheckIfExpire() {
				values[i] = nil
				i++
				continue
			}
			values[i] = meta.Extra
		} else {
			values[i] = nil
		}
		i++
	}
	return values, nil
}

// MSet operation
func (st *STRING) MSet(items map[string][]byte) error {
	mkeys := make([][]byte, len(items))
	keyValues := make(map[string][]byte)
	for key, value := range items {
		mkey := EncodeMetaKey(st.environ.Header, []byte(key))
		mkeys = append(mkeys, mkey)
		keyValues[string(mkey)] = value
	}
	result, err := st.store.BatchGet(mkeys)
	if err != nil {
		return err
	}
	var meta *Meta
	for key, value := range keyValues {
		if val, ok := result[key]; ok && val != nil {
			// exist
			meta = DecodeMetaValue(val)
			rawKey := DecodeMetaKey(st.environ.Header, []byte(key))
			if meta.ExpireAt > 0 {
				if err = RemoveExpire(st.environ, st.store, rawKey, meta.ExpireAt); err != nil {
					return err
				}
			}
		} else {
			meta = NewMeta(types.STRING)
		}
		meta.Extra = value
		meta.ExpireAt = 0
		if err = st.store.Set([]byte(key), meta.EncodeMetaValue()); err != nil {
			return err
		}
	}
	if err := st.store.Commit(); err != nil {
		st.environ.FailedTxn("MSet")
		return err
	}
	return nil
}

// Incr operation
func (st *STRING) Incr(key []byte, step int64) (int64, error) {
	if err := st.existsForWrite(key); err != nil {
		return 0, err
	}
	var num int64
	var err error
	if st.value == nil {
		num = step
	} else {
		num, err = strconv.ParseInt(string(st.value), 10, 64)
		if err != nil {
			return 0, terror.ErrTypeTrans
		}
		num = num + step
	}
	st.meta.Extra = []byte(strconv.FormatInt(num, 10))
	if err := st.store.Set(st.metaKey, st.meta.EncodeMetaValue()); err != nil {
		return 0, err
	}
	if err := st.store.Commit(); err != nil {
		st.environ.FailedTxn("Incr")
		return 0, err
	}
	return num, nil
}

// Strlen operation
func (st *STRING) Strlen(key []byte) (uint, error) {
	if err := st.existsForRead(key); err != nil {
		return 0, nil
	}
	if st.value == nil {
		return 0, nil
	}
	return uint(len(st.value)), nil
}
