package store

import (
	"github.com/gengmei-tech/titea/server/metrics"
	"github.com/gengmei-tech/titea/server/types"
)

// Header the completed header = prefix|namespace
type Header struct {
	header []byte
	hlen   int
}

// Environ contain context about a client
type Environ struct {

	// system  prefix, default kv
	Prefix []byte

	// Namespace default.default
	Namespace *types.Namespace

	// redis data type
	Type byte

	// opeation (add delete gc expire)
	Operation byte

	// prefix|namespace
	Header []byte

	// 头长度
	HeaderLen int

	// 操作命令数量
	OpCount uint64
}

var headers = make(map[string]Header)

// CreateEnviron to new environ
func CreateEnviron(namespace *types.Namespace) *Environ {
	environ := Environ{
		Prefix:    []byte(types.SystemPrefix),
		Namespace: namespace,
		OpCount:   0,
	}
	environ.loadHeader()
	return &environ
}

// Update when change namespace
func (environ *Environ) Update(namespace *types.Namespace) {
	environ.Namespace = namespace
	environ.loadHeader()
}

func (environ *Environ) loadHeader() {
	if header, ok := headers[environ.Namespace.Name]; ok {
		environ.Header = header.header
		environ.HeaderLen = header.hlen
	} else {
		buffer := make([]byte, 2+len(environ.Namespace.Name))
		copy(buffer[0:], environ.Prefix)
		copy(buffer[2:], []byte(environ.Namespace.Name))
		header := Header{
			header: buffer,
			hlen:   len(buffer),
		}
		headers[environ.Namespace.Name] = header
		environ.Header = header.header
		environ.HeaderLen = header.hlen
	}
}

// SetType when change redis command type
func (environ *Environ) SetType(tp byte) {
	environ.Type = tp
}

// FailedTxn record metric of txn failed
func (environ *Environ) FailedTxn(event string) {
	metrics.TxnFailedCounter.WithLabelValues(environ.Namespace.Group, environ.Namespace.Service, types.StringType(environ.Type), event).Inc()
}

// AddGc record metric of add gc
func (environ *Environ) AddGc() {
	metrics.GcAddCounter.WithLabelValues(environ.Namespace.Group, environ.Namespace.Service, types.StringType(environ.Type)).Inc()
}

// ExecGc record metric of exec gc
func (environ *Environ) ExecGc() {
	metrics.GcExecCounter.WithLabelValues(environ.Namespace.Group, environ.Namespace.Service, types.StringType(environ.Type)).Inc()
}

// ExecExpire recode metric of exec expire
func (environ *Environ) ExecExpire() {
	metrics.ExpireExecCounter.WithLabelValues(environ.Namespace.Group, environ.Namespace.Service, types.StringType(environ.Type)).Inc()
}
