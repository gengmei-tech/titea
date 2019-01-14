package server

import (
	"fmt"
	"github.com/gengmei-tech/titea/server/config"
	"github.com/gengmei-tech/titea/server/metrics"
	ns "github.com/gengmei-tech/titea/server/namespace"
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/types"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

var (
	baseConnID uint32
)

// Server TiTea
type Server struct {
	cfg      *config.Config
	listener net.Listener
	store    *store.Store
	auth     string
	rwlock   *sync.RWMutex
	clients  map[uint32]*Client
}

// NewServer initialize an TiTea server
func NewServer(cfg *config.Config) *Server {
	server := Server{
		cfg:     cfg,
		auth:    cfg.Auth,
		rwlock:  &sync.RWMutex{},
		clients: make(map[uint32]*Client),
		store:   store.Open(cfg.Backend.PdAddr),
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	server.listener = listener
	log.Infof("Server is running at [%s]", addr)
	return &server
}

// deal some thing before run
func (s *Server) beforeRun() {
	// Load namespace from tikv
	s.store = s.store.Read()
	ns.LoadNamespace(s.store)

	// start expire and gc
	if s.cfg.RunExpGc {
		allNs := ns.GetAllNamespace()
		var namespaces []*types.Namespace
		for _, namespace := range allNs {
			if namespace.Type == ns.TYPESYSTEM {
				continue
			}
			namespaces = append(namespaces, namespace)
		}
		store.StartExpireGc(s.store, namespaces)
	}
}

// Run Server
func (s *Server) Run() error {
	s.beforeRun()
	defer s.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Err.Error() == "use of closed network connection" {
					return nil
				}
			}
			log.Errorf("accept error %s", err.Error())
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	c := newClient(s)
	c.setConn(conn)
	startTime := time.Now()
	log.Infof("[server] con:%d new connection %s", c.connectionID, conn.RemoteAddr().String())
	defer func() {
		log.Infof("[server] con:%d close connection %s, OpCount:%d, ConsumeTime:%.4f [ms]", c.connectionID, conn.RemoteAddr().String(), c.environ.OpCount, time.Since(startTime).Seconds()*1000)
	}()
	s.rwlock.Lock()
	s.clients[c.connectionID] = c
	connections := len(s.clients)
	s.rwlock.Unlock()
	metrics.ConnGauge.Set(float64(connections))
	c.run()
}

// Close Server
func (s *Server) Close() {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	if s.store != nil {
		s.store.Close()
		s.store = nil
	}
}

func (s *Server) getAllClients() []string {
	var clients []string
	if len(s.clients) > 0 {
		for _, c := range s.clients {
			clients = append(clients, c.conn.RemoteAddr().String())
		}
	}
	return clients
}

func (s *Server) getParams() [][]byte {
	var params [][]byte
	params = append(params, []byte(fmt.Sprintf("RunExpireGC: %t", s.cfg.RunExpGc)))
	return params
}
