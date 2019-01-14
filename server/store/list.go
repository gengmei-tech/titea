package store

import (
	"github.com/gengmei-tech/titea/pkg/util/number"
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
)

const (
	// POSITIVE represent positive, ascii 76
	POSITIVE = byte('>')

	// NEGATIVE represent negative, ascii 74
	NEGATIVE = byte('<')
)

// List implement redis list operation
type List struct {
	store *Store

	// 上下文环境
	environ *Environ

	// hash key
	key []byte

	// metaKey 不需要再次encode
	metaKey []byte

	// 元信息
	meta *Meta

	// index
	head int64
	tail int64
}

// InitList init
func InitList(environ *Environ, store *Store, key []byte) (*List, error) {
	mkey := EncodeMetaKey(environ.Header, key)
	value, err := store.Get(mkey)
	if err != nil {
		return nil, err
	}
	list := List{
		store:   store,
		environ: environ,
		key:     key,
		metaKey: mkey,
	}
	if value != nil {
		meta := DecodeMetaValue(value)
		if !meta.CheckType(types.LIST) {
			return nil, terror.ErrTypeNotMatch
		}
		list.head = number.BytesToInt64(meta.Extra[0:8])
		list.tail = number.BytesToInt64(meta.Extra[8:16])
		list.meta = meta
	}
	return &list, nil
}

// ExistsForWrite exist for write or not exist
func (list *List) ExistsForWrite() error {
	if list.meta == nil {
		list.meta = NewMeta(types.LIST)
		list.meta.Extra = make([]byte, 16)
	} else {
		if list.meta.CheckIfExpire() {
			if err := list.destroy(); err != nil {
				return err
			}
			// reset meta
			list.meta.Reset()
		}
	}
	return nil
}

// ExistsForRead exist for read or not exist
func (list *List) ExistsForRead() error {
	if list.meta == nil {
		return terror.ErrKeyNotExist
	}
	if list.meta.CheckIfExpire() {
		list.store.WriteReset()
		// 删除meta
		if err := list.store.Delete(list.metaKey); err != nil {
			return err
		}
		if err := list.destroy(); err != nil {
			return err
		}
		if err := list.store.Commit(); err != nil {
			list.environ.FailedTxn("ExistsForRead")
			return err
		}
		return terror.ErrKeyNotExist
	}
	return nil
}

// EncodeDataKey encode list field key
// header|d(1)|uuid(16)|pos or neg|order(8)
func (list *List) EncodeDataKey(order int64) []byte {
	buffer := make([]byte, list.environ.HeaderLen+26)
	copy(buffer[0:], list.environ.Header)
	buffer[list.environ.HeaderLen] = types.DATA
	index := list.environ.HeaderLen + 1
	copy(buffer[index:], list.meta.ID)
	index = index + 16
	if order < 0 {
		buffer[index] = NEGATIVE
	} else {
		buffer[index] = POSITIVE
	}
	index++
	copy(buffer[index:], number.Int64ToBytes(order))
	return buffer
}

// Prefix list data key prefix
// header|d(1)|uuid(16)|
func (list *List) Prefix() []byte {
	buffer := make([]byte, list.environ.HeaderLen+17)
	copy(buffer[0:], list.environ.Header)
	buffer[list.environ.HeaderLen] = types.DATA
	copy(buffer[list.environ.HeaderLen+1:], list.meta.ID)
	return buffer
}

// LPush push elements from left
func (list *List) LPush(values ...[]byte) (uint64, error) {
	if list.Len() == 0 {
		list.head = 0
		list.tail = -1
	}
	for _, value := range values {
		list.head--
		dataKey := list.EncodeDataKey(list.head)
		if err := list.store.Set(dataKey, value); err != nil {
			return 0, err
		}
		list.meta.Count++
	}
	copy(list.meta.Extra[0:], number.Int64ToBytes(list.head))
	copy(list.meta.Extra[8:], number.Int64ToBytes(list.tail))
	if err := list.store.Set(list.metaKey, list.meta.EncodeMetaValue()); err != nil {
		return 0, err
	}
	if err := list.store.Commit(); err != nil {
		list.environ.FailedTxn("LPush")
		return 0, err
	}
	return list.meta.Count, nil
}

// RPush push elements from right
func (list *List) RPush(values ...[]byte) (uint64, error) {
	if list.Len() == 0 {
		list.head = 0
		list.tail = -1
	}
	for _, value := range values {
		list.tail++
		dataKey := list.EncodeDataKey(list.tail)
		if err := list.store.Set(dataKey, value); err != nil {
			return 0, err
		}
		list.meta.Count++
	}
	copy(list.meta.Extra[0:8], number.Int64ToBytes(list.head))
	copy(list.meta.Extra[8:16], number.Int64ToBytes(list.tail))
	if err := list.store.Set(list.metaKey, list.meta.EncodeMetaValue()); err != nil {
		return 0, err
	}
	if err := list.store.Commit(); err != nil {
		list.environ.FailedTxn("RPush")
		return 0, err
	}
	return list.meta.Count, nil
}

// LPop pop from left
func (list *List) LPop() ([]byte, error) {
	dataKey := list.EncodeDataKey(list.head)
	value, err := list.store.Get(dataKey)
	if err != nil {
		return nil, err
	}
	if err = list.store.Delete(dataKey); err != nil {
		return nil, err
	}
	list.meta.Count--
	list.head++
	if list.meta.Count <= 0 {
		if err = list.store.Delete(list.metaKey); err != nil {
			return nil, err
		}
	} else {
		copy(list.meta.Extra[0:8], number.Int64ToBytes(list.head))
		if err = list.store.Set(list.metaKey, list.meta.EncodeMetaValue()); err != nil {
			return nil, err
		}
	}
	if err := list.store.Commit(); err != nil {
		list.environ.FailedTxn("LPop")
		return nil, err
	}
	return value, nil
}

// RPop pop from right
func (list *List) RPop() ([]byte, error) {
	dataKey := list.EncodeDataKey(list.tail)
	value, err := list.store.Get(dataKey)
	if err != nil {
		return nil, err
	}
	if err = list.store.Delete(dataKey); err != nil {
		return nil, err
	}
	list.meta.Count--
	list.tail--
	if list.meta.Count <= 0 {
		if err = list.store.Delete(list.metaKey); err != nil {
			return nil, err
		}
	} else {
		copy(list.meta.Extra[8:16], number.Int64ToBytes(list.tail))
		if err = list.store.Set(list.metaKey, list.meta.EncodeMetaValue()); err != nil {
			return nil, err
		}
	}
	if err := list.store.Commit(); err != nil {
		list.environ.FailedTxn("RPop")
		return nil, err
	}
	return value, nil
}

//Len list len
func (list *List) Len() uint64 {
	return list.meta.Count
}

// Range return elements between index range
func (list *List) Range(offset uint64, limit uint64) ([][]byte, error) {
	dataKey := list.EncodeDataKey(list.head + int64(offset))
	return list.store.ScanValues(list.Prefix(), dataKey, 0, limit)
}

// Index return element in specified index
func (list *List) Index(index int64) ([]byte, error) {
	index = list.head + index
	dataKey := list.EncodeDataKey(index)
	return list.store.Get(dataKey)
}

// Set elements in a specified index
func (list *List) Set(index int64, value []byte) error {
	index = list.head + index
	if index > list.tail {
		return terror.ErrOutOfIndex
	}
	dataKey := list.EncodeDataKey(index)
	if err := list.store.Set(dataKey, value); err != nil {
		return err
	}
	if err := list.store.Commit(); err != nil {
		list.environ.FailedTxn("Set")
		return err
	}
	return nil
}

// add to gc and remove from expire set
func (list *List) destroy() error {
	// add to gc
	if err := AddToGc(list.environ, list.store, list.meta.ID); err != nil {
		return err
	}
	// remove from expire set
	return RemoveExpire(list.environ, list.store, list.key, list.meta.ExpireAt)
}
