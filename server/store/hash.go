package store

import (
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	log "github.com/sirupsen/logrus"
	"strconv"
)

// Hash implement redis hash
type Hash struct {
	// db
	store *Store

	// context
	environ *Environ

	// hash key
	key []byte

	// metaKey
	metaKey []byte

	// meta
	meta *Meta
}

// InitHash new a hash object
func InitHash(environ *Environ, store *Store, key []byte) (*Hash, error) {
	mkey := EncodeMetaKey(environ.Header, key)
	value, err := store.Get(mkey)
	if err != nil {
		return nil, err
	}
	hash := Hash{
		store:   store,
		environ: environ,
		key:     key,
		metaKey: mkey,
	}
	if value != nil {
		meta := DecodeMetaValue(value)
		if !meta.CheckType(types.HASH) {
			return nil, terror.ErrTypeNotMatch
		}
		hash.meta = meta
	}
	return &hash, nil
}

// ExistsForWrite exists fro write or not
func (hash *Hash) ExistsForWrite() error {
	if hash.meta == nil {
		hash.meta = NewMeta(types.HASH)
	} else {
		if hash.meta.CheckIfExpire() {
			// add to gc and remove from expire set
			if err := hash.destroy(); err != nil {
				return err
			}
			hash.meta.Reset()
		}
	}
	return nil
}

// ExistsForRead exists for read or not exist
func (hash *Hash) ExistsForRead() error {
	if hash.meta == nil {
		return terror.ErrKeyNotExist
	}
	if hash.meta.CheckIfExpire() {
		hash.store.WriteReset()
		if err := hash.store.Delete(hash.metaKey); err != nil {
			return err
		}
		// add to gc and remove from expire set
		if err := hash.destroy(); err != nil {
			return err
		}
		if err := hash.store.Commit(); err != nil {
			// 日志
			hash.environ.FailedTxn("ExistsForRead")
			log.Infof("key:%s, expireAt:%d\n", hash.key, hash.meta.ExpireAt)
			return err
		}
		return terror.ErrKeyNotExist
	}
	return nil
}

// EncodeDataKey encode hash field key
// header|d(1)|uuid(16)|field
func (hash *Hash) EncodeDataKey(field []byte) []byte {
	buffer := make([]byte, hash.environ.HeaderLen+17+len(field))
	copy(buffer[0:], hash.environ.Header)
	buffer[hash.environ.HeaderLen] = types.DATA
	copy(buffer[hash.environ.HeaderLen+1:], hash.meta.ID)
	copy(buffer[hash.environ.HeaderLen+17:], field)
	return buffer
}

// DecodeDataKey decode hash field key
// header|d(1)|uuid(16)|field
func (hash *Hash) DecodeDataKey(key []byte) ([]byte, []byte) {
	index := hash.environ.HeaderLen + 1
	return key[index : index+16], key[index+16:]
}

// Prefix hash field key prefix
// header|d|uuid(16)
func (hash *Hash) Prefix() []byte {
	buffer := make([]byte, hash.environ.HeaderLen+17)
	copy(buffer[0:], hash.environ.Header)
	buffer[hash.environ.HeaderLen] = types.DATA
	copy(buffer[hash.environ.HeaderLen+1:], hash.meta.ID)
	return buffer
}

// Set a field return 0 if exists else 1 if not
func (hash *Hash) Set(field []byte, value []byte) (uint8, error) {
	dataKey := hash.EncodeDataKey(field)
	val, err := hash.store.Get(dataKey)
	var status uint8
	if err != nil {
		return status, err
	}
	if err = hash.store.Set(dataKey, value); err != nil {
		return status, err
	}
	if val == nil {
		// new add
		hash.meta.Count++
		if err = hash.store.Set(hash.metaKey, hash.meta.EncodeMetaValue()); err != nil {
			return status, err
		}
		if err = hash.store.Commit(); err != nil {
			hash.environ.FailedTxn("Set")
			return status, err
		}
		status = 1
	}
	return status, nil
}

// MSet set multil hash fields
func (hash *Hash) MSet(items map[string][]byte) (uint64, error) {
	dataKeys := make([][]byte, len(items))
	keyValues := make(map[string][]byte)
	for field, value := range items {
		dataKey := hash.EncodeDataKey([]byte(field))
		dataKeys = append(dataKeys, dataKey)
		keyValues[string(dataKey)] = value
	}
	result, err := hash.store.BatchGet(dataKeys)
	if err != nil {
		return 0, err
	}
	var addCnt uint64
	for key, value := range keyValues {
		if err = hash.store.Set([]byte(key), value); err != nil {
			return 0, err
		}
		// 已经存在
		if val, ok := result[key]; ok && val != nil {
			continue
		}
		addCnt++
	}
	if addCnt > 0 {
		hash.meta.Count += addCnt
		if err = hash.store.Set(hash.metaKey, hash.meta.EncodeMetaValue()); err != nil {
			return 0, err
		}
	}
	if err := hash.store.Commit(); err != nil {
		hash.environ.FailedTxn("MSet")
		return 0, err
	}
	return addCnt, nil
}

// Field get a field
func (hash *Hash) Field(field []byte) ([]byte, error) {
	dataKey := hash.EncodeDataKey(field)
	return hash.store.Get(dataKey)
}

// Fields get mulit field
func (hash *Hash) Fields(fields [][]byte) ([][]byte, error) {
	dataKeys := make([][]byte, len(fields))
	for i, field := range fields {
		dataKeys[i] = hash.EncodeDataKey(field)
	}
	result, err := hash.store.BatchGet(dataKeys)
	if err != nil {
		return nil, err
	}
	values := make([][]byte, len(fields))
	i := 0
	for _, dataKey := range dataKeys {
		if val, ok := result[string(dataKey)]; ok {
			values[i] = val
		} else {
			values[i] = nil
		}
		i++
	}
	return values, nil
}

//Count hash fields
func (hash *Hash) Count() uint64 {
	return hash.meta.Count
}

// Keys hash all keys
func (hash *Hash) Keys() ([][]byte, error) {
	result, err := hash.store.ScanKeys(hash.Prefix(), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	keys := make([][]byte, len(result))
	for i, key := range result {
		_, keys[i] = hash.DecodeDataKey(key)
	}
	return keys, err
}

// Values hash all values
func (hash *Hash) Values() ([][]byte, error) {
	values, err := hash.store.ScanValues(hash.Prefix(), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return values, err
}

// All keys and values
func (hash *Hash) All() ([][]byte, error) {
	result, err := hash.store.ScanAll(hash.Prefix(), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	total := len(result)
	keyValues := make([][]byte, total)
	for i := 0; i < total; i += 2 {
		_, keyValues[i] = hash.DecodeDataKey(result[i])
		keyValues[i+1] = result[i+1]
	}
	return keyValues, err
}

// Incrby a field
func (hash *Hash) Incrby(field []byte, step int64) (int64, error) {
	dataKey := hash.EncodeDataKey(field)
	value, err := hash.store.Get(dataKey)
	if err != nil {
		return 0, err
	}
	var num int64
	if value == nil {
		num = step
		hash.meta.Count++
		if err = hash.store.Set(hash.metaKey, hash.meta.EncodeMetaValue()); err != nil {
			return 0, err
		}
	} else {
		num, err = strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			num = step
		} else {
			num += step
		}
	}
	if err = hash.store.Set(dataKey, []byte(strconv.FormatInt(num, 10))); err != nil {
		return 0, err
	}
	if err := hash.store.Commit(); err != nil {
		hash.environ.FailedTxn("Incrby")
		return 0, err
	}
	return num, nil
}

// Remove one field
func (hash *Hash) Remove(fields ...[]byte) (uint64, error) {
	dataKeys := make([][]byte, len(fields))
	for i, field := range fields {
		dataKeys[i] = hash.EncodeDataKey(field)
	}
	result, err := hash.store.BatchGet(dataKeys)
	if err != nil {
		return 0, err
	}
	var delCnt uint64
	for _, key := range dataKeys {
		if val, ok := result[string(key)]; ok && val != nil {
			// 存在的key
			if err = hash.store.Delete(key); err != nil {
				return 0, err
			}
			delCnt++
		}
	}
	if delCnt > 0 {
		hash.meta.Count -= delCnt
		if hash.meta.Count <= 0 {
			if err = hash.store.Delete(hash.metaKey); err != nil {
				return 0, err
			}
		} else {
			if err = hash.store.Set(hash.metaKey, hash.meta.EncodeMetaValue()); err != nil {
				return 0, err
			}
		}
	}
	if err := hash.store.Commit(); err != nil {
		hash.environ.FailedTxn("Remove")
		return 0, err
	}
	return delCnt, nil
}

// add to gc and remove from expire set
func (hash *Hash) destroy() error {
	// add to gc
	if err := AddToGc(hash.environ, hash.store, hash.meta.ID); err != nil {
		return err
	}
	// 从过期集合里删除
	if err := RemoveExpire(hash.environ, hash.store, hash.key, hash.meta.ExpireAt); err != nil {
		return err
	}
	return nil
}
