package store

import (
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
)

// Set implement redis set command
type Set struct {
	store *Store

	// environ context
	environ *Environ

	// hash key
	key []byte

	// metaKey
	metaKey []byte

	// Meta
	meta *Meta
}

// InitSet init
func InitSet(environ *Environ, store *Store, key []byte) (*Set, error) {
	mkey := EncodeMetaKey(environ.Header, key)
	value, err := store.Get(mkey)
	if err != nil {
		return nil, err
	}
	set := Set{
		store:   store,
		environ: environ,
		key:     key,
		metaKey: mkey,
	}
	if value != nil {
		meta := DecodeMetaValue(value)
		if !meta.CheckType(types.SET) {
			return nil, terror.ErrTypeNotMatch
		}
		set.meta = meta
	}
	return &set, nil
}

// ExistsForWrite exist for write or not exist
func (set *Set) ExistsForWrite() error {
	if set.meta == nil {
		set.meta = NewMeta(types.SET)
	} else {
		if set.meta.CheckIfExpire() {
			if err := set.destroy(); err != nil {
				return err
			}
			// reset meta
			set.meta.Reset()
		}
	}
	return nil
}

// ExistsForRead exist for read or not exist
func (set *Set) ExistsForRead() error {
	if set.meta == nil {
		return terror.ErrKeyNotExist
	}
	if set.meta.CheckIfExpire() {
		set.store.WriteReset()
		// delete meta
		if err := set.store.Delete(set.metaKey); err != nil {
			return err
		}
		if err := set.destroy(); err != nil {
			return err
		}
		if err := set.store.Commit(); err != nil {
			set.environ.FailedTxn("ExistsForRead")
			return err
		}
		return terror.ErrKeyNotExist
	}
	return nil
}

// EncodeDataKey encode hash field key
// header|type|uuid|member
func (set *Set) EncodeDataKey(member []byte) []byte {
	buffer := make([]byte, set.environ.HeaderLen+17+len(member))
	copy(buffer[0:], set.environ.Header)
	buffer[set.environ.HeaderLen] = types.DATA
	copy(buffer[set.environ.HeaderLen+1:], set.meta.ID)
	copy(buffer[set.environ.HeaderLen+17:], member)
	return buffer
}

// DecodeDataKey header|type|uuid|member
func (set *Set) DecodeDataKey(key []byte) ([]byte, []byte) {
	index := set.environ.HeaderLen + 1
	return key[index : index+16], key[index+16:]
}

// Prefix prefix
func (set *Set) Prefix() []byte {
	buffer := make([]byte, set.environ.HeaderLen+17)
	copy(buffer[0:], set.environ.Header)
	buffer[set.environ.HeaderLen] = types.DATA
	copy(buffer[set.environ.HeaderLen+1:], set.meta.ID)
	return buffer
}

// Add member to set
func (set *Set) Add(members ...[]byte) (uint64, error) {
	keyValues := make(map[string][]byte)
	dataKeys := make([][]byte, len(members))
	for i, member := range members {
		dataKey := set.EncodeDataKey(member)
		keyValues[string(dataKey)] = member
		dataKeys[i] = dataKey
	}
	result, err := set.store.BatchGet(dataKeys)
	if err != nil {
		return 0, err
	}
	var addCnt uint64
	for key, value := range keyValues {
		if _, ok := result[key]; ok {
			// exists
			continue
		}
		if err = set.store.Set([]byte(key), value); err != nil {
			return 0, err
		}
		addCnt++
	}
	if addCnt > 0 {
		set.meta.Count += addCnt
		if err = set.store.Set(set.metaKey, set.meta.EncodeMetaValue()); err != nil {
			return 0, err
		}
	}
	if err := set.store.Commit(); err != nil {
		set.environ.FailedTxn("Add")
		return 0, err
	}
	return addCnt, nil
}

// IsMember check whether a member in a set
func (set *Set) IsMember(member []byte) (uint8, error) {
	dataKey := set.EncodeDataKey(member)
	value, err := set.store.Get(dataKey)
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, nil
	}
	return 1, nil
}

// Members get all member
func (set *Set) Members() ([][]byte, error) {
	prefix := set.Prefix()
	return set.store.ScanValues(prefix, nil, 0, 0)
}

// Remove members from set
func (set *Set) Remove(members ...[]byte) (uint64, error) {
	dataKeys := make([][]byte, len(members))
	for i, member := range members {
		dataKey := set.EncodeDataKey(member)
		dataKeys[i] = dataKey
	}
	result, err := set.store.BatchGet(dataKeys)
	if err != nil {
		return 0, err
	}
	var delCnt uint64
	for _, dataKey := range dataKeys {
		if val, ok := result[string(dataKey)]; ok && val != nil {
			// exist
			if err = set.store.Delete(dataKey); err != nil {
				return 0, err
			}
			delCnt++
		}
	}
	if delCnt > 0 {
		set.meta.Count -= delCnt
		if set.meta.Count <= 0 {
			if err = set.store.Delete(set.metaKey); err != nil {
				return 0, err
			}
		} else {
			if err = set.store.Set(set.metaKey, set.meta.EncodeMetaValue()); err != nil {
				return 0, err
			}
		}
	}
	if err := set.store.Commit(); err != nil {
		set.environ.FailedTxn("Remove")
		return 0, nil
	}
	return delCnt, nil
}

// Card member count
func (set *Set) Card() uint64 {
	return set.meta.Count
}

// add to gc and remove from expire set
func (set *Set) destroy() error {
	// add to gc
	if err := AddToGc(set.environ, set.store, set.meta.ID); err != nil {
		return err
	}
	// remove from expire set
	return RemoveExpire(set.environ, set.store, set.key, set.meta.ExpireAt)
}
