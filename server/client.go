package server

import (
	"bufio"
	"io"
	"math"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gengmei-tech/titea/pkg/redis"
	"github.com/gengmei-tech/titea/server/metrics"
	"github.com/gengmei-tech/titea/server/namespace"
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	log "github.com/sirupsen/logrus"
)

// Client represent a client connection
type Client struct {
	server       *Server
	connectionID uint32

	store *store.Store

	// params
	cmd  string
	args [][]byte
	argc int

	// connection authentation isAuthed 是否验证 needAuth 是否需要验证
	needAuth bool
	isAuthed bool

	// connection
	conn net.Conn

	// redis read or write
	reader *redis.Reader
	writer *redis.Writer

	// context
	environ *store.Environ
}

func newClient(s *Server) *Client {
	client := &Client{
		server:       s,
		connectionID: atomic.AddUint32(&baseConnID, 1),
		store:        s.store.Read(),
	}
	if s.auth == "" {
		client.needAuth = false
	} else {
		client.needAuth = true
	}
	client.isAuthed = false
	return client
}

func (c *Client) setConn(conn net.Conn) {
	c.conn = conn
	c.reader = redis.NewReader(bufio.NewReader(conn))
	c.writer = redis.NewWriter(bufio.NewWriter(conn))
	c.environ = store.CreateEnviron(namespace.GetDefaultNamespace())
}

// Run
func (c *Client) run() {
	defer func() {
		c.close()
	}()

	var req [][]byte
	var err error

	for {
		if req, err = c.reader.ParseRequest(); err != nil {
			// 结束
			if err == io.EOF {
				break
			}
			log.Warn(err.Error())
			break
		}

		if err = c.handleRequest(req); err != nil {
			log.Warn(err.Error())
			break
		}
	}
}

func (c *Client) handleRequest(req [][]byte) error {
	if len(req) == 0 {
		return c.writer.Error(terror.ErrCommand)
	}

	c.cmd = strings.ToLower(string(req[0]))
	c.args = req[1:]
	c.argc = len(c.args)

	// Need Auth
	if c.needAuth && !c.isAuthed {
		return c.writer.Error(terror.ErrAuthRequired)
	}
	// Error
	if len(c.cmd) == 0 {
		return c.writer.Error(terror.ErrCommand)
	}
	return c.execute()
}

func (c *Client) close() {
	c.server.rwlock.Lock()
	delete(c.server.clients, c.connectionID)
	connections := len(c.server.clients)
	c.server.rwlock.Unlock()
	metrics.ConnGauge.Set(float64(connections))
	c.conn.Close()
	return
}

func (c *Client) execute() error {
	// redis-cli will send "command" when connected
	command, ok := lookupCommand(c.cmd)
	if !ok {
		return c.writer.Error(terror.ErrCommand)
	}

	// check args
	if command.Arity > 0 && c.argc != command.Arity {
		return c.writer.Error(terror.ErrCmdParams)
	} else if command.Arity < 0 && c.argc < int(math.Abs(float64(command.Arity))) {
		return c.writer.Error(terror.ErrCmdParams)
	}

	// if value is "", so Fill(tikv don't support empty value)
	if command.Flags == types.FlagWrite && c.argc > 1 {
		for i, r := range c.args[1:] {
			if len(r) == 0 {
				c.args[i+1] = redis.EmptyFill()
			}
		}
	}

	// default is read
	if command.Flags == types.FlagWrite {
		c.store.WriteReset()
	} else if command.Flags == types.FlagRead {
		c.store.SetRead()
	}

	c.environ.SetType(command.Type)
	c.environ.OpCount++

	startTime := time.Now()
	defer func() {
		cost := time.Since(startTime)
		metrics.CommandDuration.WithLabelValues(c.environ.Namespace.Group, c.environ.Namespace.Service, c.cmd).Observe(cost.Seconds())
		if c.argc > 0 {
			log.Infof("[consume] Cmd:%s, Params:%s, Consume:%.2f ms", c.cmd, c.args[0], cost.Seconds()*1000)
		}
	}()

	// 执行命令
	return command.Function(c)
}
