package store

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	"unsafe"
)

// Store implement communicate with TiKv
type Store struct {
	// 读写
	DB kv.Storage

	// 快照 用来read
	Snapshot kv.Snapshot

	// 事物 用来读写操作
	Txn kv.Transaction

	// 是否只读
	OnlyRead bool
}

// Open connection with TiKv
func Open(pdAddr string) *Store {
	driver := tikv.Driver{}
	storage, err := driver.Open(fmt.Sprintf("tikv://%s/pd?cluster=1", pdAddr))
	if err != nil {
		panic(err)
	}
	store := Store{
		DB: storage,
	}
	return &store
}

// Read operation
func (store *Store) Read() *Store {
	snapshot, err := store.DB.GetSnapshot(kv.MaxVersion)
	if err != nil {
		panic(err)
	}
	newStore := Store{
		DB:       store.DB,
		Snapshot: snapshot,
		OnlyRead: true,
	}
	return &newStore
}

// ReadReset reset read
func (store *Store) ReadReset() *Store {
	snapshot, err := store.DB.GetSnapshot(kv.MaxVersion)
	if err != nil {
		panic(err)
	}
	store.Snapshot = snapshot
	store.OnlyRead = true
	return store
}

// SetRead for read
func (store *Store) SetRead() {
	store.OnlyRead = true
}

// SetWrite for write
func (store *Store) SetWrite() {
	store.OnlyRead = false
}

// Write operation
func (store *Store) Write() *Store {
	txn, err := store.DB.Begin()
	if err != nil {
		panic(err)
	}
	newStore := Store{
		DB:       store.DB,
		Txn:      txn,
		OnlyRead: false,
	}
	return &newStore
}

// WriteReset reset write
func (store *Store) WriteReset() *Store {
	txn, err := store.DB.Begin()
	if err != nil {
		panic(err)
	}
	store.Txn = txn
	store.OnlyRead = false
	return store
}

// Commit txn
func (store *Store) Commit() error {
	if err := store.Txn.Commit(context.Background()); err != nil {
		store.Txn.Rollback()
		return err
	}
	return nil
}

// Close connection
func (store *Store) Close() error {
	return store.DB.Close()
}

// Set operation
func (store *Store) Set(key []byte, value []byte) error {
	return store.Txn.Set(key, value)
}

// Get operation
func (store *Store) Get(key []byte) ([]byte, error) {
	var (
		value []byte
		err   error
	)
	if store.OnlyRead {
		value, err = store.Snapshot.Get(key)
	} else {
		value, err = store.Txn.Get(key)
	}
	if err != nil && kv.IsErrNotFound(err) {
		return nil, nil
	}
	return value, err
}

// BatchGet operation
func (store *Store) BatchGet(keys [][]byte) (map[string][]byte, error) {
	k := *(*[]kv.Key)(unsafe.Pointer(&keys))
	if store.OnlyRead {
		return store.Snapshot.BatchGet(k)
	}
	return store.Txn.GetSnapshot().BatchGet(k)
}

// Seek operation
func (store *Store) Seek(key []byte) (kv.Iterator, error) {
	if store.OnlyRead {
		return store.Snapshot.Seek(key)
	}
	return store.Txn.GetSnapshot().Seek(key)
}

// Delete operation
func (store *Store) Delete(key []byte) error {
	return store.Txn.Delete(key)
}

// Index position from prefix
func (store *Store) Index(prefix []byte, key []byte) (uint64, error) {
	iter, err := store.Seek(prefix)
	if err != nil {
		return 0, err
	}
	defer iter.Close()
	var index uint64
	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
		if bytes.Compare(iter.Key(), key) == 0 {
			break
		}
		index++
		iter.Next()
	}
	return index, nil
}

// ScanKeys scan with prefix and return keys
func (store *Store) ScanKeys(prefix []byte, start []byte, offset uint64, limit uint64) ([][]byte, error) {
	return store.Scan(prefix, start, true, false, offset, limit)
}

// ScanValues scan with prefix and return values
func (store *Store) ScanValues(prefix []byte, start []byte, offset uint64, limit uint64) ([][]byte, error) {
	return store.Scan(prefix, start, false, true, offset, limit)
}

// ScanAll scan with prefix and return keys and values
func (store *Store) ScanAll(prefix []byte, start []byte, offset uint64, limit uint64) ([][]byte, error) {
	return store.Scan(prefix, start, true, true, offset, limit)
}

// Scan operation
func (store *Store) Scan(prefix []byte, start []byte, withKey bool, withValue bool, offset, limit uint64) ([][]byte, error) {
	var iter kv.Iterator
	var err error
	if start == nil {
		iter, err = store.Seek(prefix)
	} else {
		iter, err = store.Seek(start)
	}
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var result [][]byte
	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
		if offset > 0 {
			offset--
			iter.Next()
			continue
		}
		if withKey {
			result = append(result, iter.Key())
		}
		if withValue {
			result = append(result, iter.Value())
		}
		iter.Next()
		if limit > 0 {
			limit--
			if limit == 0 {
				break
			}
		}
	}
	return result, nil
}

// DeletePrefix delete with prefix
func (store *Store) DeletePrefix(prefix []byte) (uint64, error) {
	iter, err := store.Seek(prefix)
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	var delCnt uint64
	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
		if err := store.Delete(iter.Key()); err != nil {
			continue
		}
		iter.Next()
		delCnt++
	}
	return delCnt, nil
}

// CountPrefix count keys with prefix
func (store *Store) CountPrefix(prefix []byte) (uint64, error) {
	iter, err := store.Seek(prefix)
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	var total uint64
	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
		total++
		iter.Next()
	}
	return total, nil
}
