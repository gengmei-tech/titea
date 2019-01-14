package store

import (
	"bytes"
	"github.com/gengmei-tech/titea/server/types"
	"time"
)

// DecodeGcKey return uuid
func DecodeGcKey(key []byte) []byte {
	return key[len(key)-16:]
}

//AddToGc value = key|type
func AddToGc(environ *Environ, store *Store, id []byte) error {
	var gcKey = make([]byte, environ.HeaderLen+17)
	copy(gcKey[0:], environ.Header)
	gcKey[environ.HeaderLen] = types.GC
	copy(gcKey[environ.HeaderLen+1:], id)
	environ.AddGc()
	return store.Set(gcKey, []byte{environ.Type})
}

func startGc(environ *Environ, store *Store) {
	go func(environ *Environ, store *Store) {
		store = store.Read()
		for {
			runGc(environ, store)
			time.Sleep(time.Duration(SLEEPTIME))
			store.SetRead()
		}
	}(environ, store)
}

// gc specified namespace
func runGc(environ *Environ, store *Store) {
	var encodeGcDataPrefix = func(id []byte) []byte {
		buffer := make([]byte, environ.HeaderLen+17)
		copy(buffer[0:], environ.Header)
		buffer[environ.HeaderLen] = types.DATA
		copy(buffer[environ.HeaderLen+1:], id)
		return buffer
	}
	var gcPrefix = func() []byte {
		var buffer = make([]byte, environ.HeaderLen+1)
		copy(buffer[0:], environ.Header)
		buffer[environ.HeaderLen] = types.GC
		return buffer
	}
	prefix := gcPrefix()
	iter, err := store.Seek(prefix)
	if err != nil {
		return
	}
	defer iter.Close()

	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
		key := iter.Key()
		dataPrefix := encodeGcDataPrefix(DecodeGcKey(key))
		environ.SetType(iter.Value()[0])
		iter.Next()

		store.WriteReset()
		if _, err := store.DeletePrefix(dataPrefix); err != nil {
			continue
		}
		if err = store.Delete(key); err != nil {
			continue
		}
		if err := store.Commit(); err != nil {
			environ.FailedTxn("runGc")
			continue
		}
		environ.ExecGc()
	}
}
