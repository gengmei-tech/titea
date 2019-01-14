package store

import "fmt"

// Server implement server command
type Server struct {
	store   *Store
	environ *Environ
}

// InitServer init
func InitServer(store *Store) *Server {
	svr := Server{
		store: store,
	}
	return &svr
}

// FlushDB flush current namespace data
func (svr *Server) FlushDB(environ *Environ) error {
	return svr.flush(environ.Header)
}

// FlushPrefix flush begin with prefix
func (svr *Server) FlushPrefix(prefix []byte) error {
	return svr.flush(prefix)
}

// DBSize in a namespace
func (svr *Server) DBSize(environ *Environ) (uint64, error) {
	return svr.count(EncodeMetaPrefix(environ.Header))
}

// Count members
func (svr *Server) Count(prefix []byte) (uint64, error) {
	return svr.count(prefix)
}

// flush begin with prefix
func (svr *Server) flush(prefix []byte) error {
	// one txn max 5000
	var limit uint64 = 2560
	for {
		keys, err := svr.store.ScanKeys(prefix, nil, 0, limit)
		if err != nil {
			return err
		}
		if keys == nil {
			break
		}
		for _, key := range keys {
			if err := svr.store.Delete(key); err != nil {
				continue
			}
		}
		if err := svr.store.Commit(); err != nil {
			fmt.Println("flush txn error:", err.Error())
		}
		if uint64(len(keys)) < limit {
			break
		}
		svr.store.WriteReset()
	}
	return nil
}

// ScanPrefix scan with prefix
func (svr *Server) ScanPrefix(prefix []byte, offset uint64) ([][]byte, error) {
	return svr.scan(prefix, offset, 1280)
}

func (svr *Server) scan(prefix []byte, offset uint64, limit uint64) ([][]byte, error) {
	keys, err := svr.store.ScanKeys(prefix, nil, offset, limit)
	if err != nil {
		return nil, err
	}
	if keys == nil {
		return nil, err
	}
	return keys, nil
}

// count with prefix
func (svr *Server) count(prefix []byte) (uint64, error) {
	return svr.store.CountPrefix(prefix)
}
