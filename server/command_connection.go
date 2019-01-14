package server

import (
	ns "github.com/gengmei-tech/titea/server/namespace"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
)

// select group.service
func selectCommand(c *Client) error {
	dbindex, err := strconv.ParseInt(string(c.args[0]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	if dbindex < 0 {
		return c.writer.Error(terror.ErrCmdParams)
	}
	namespace, err := ns.SelectNamespace(uint64(dbindex))
	if err != nil {
		return c.writer.Error(err)
	}
	c.environ.Update(namespace)
	return c.writer.String("OK")
}

func pingCommand(c *Client) error {
	return c.writer.String("PONG")
}

func echoCommand(c *Client) error {
	return c.writer.Byte(c.args[0])
}

func authCommand(c *Client) error {
	if c.server.auth == "" {
		return c.writer.Error(terror.ErrAuthNoNeed)
	} else if string(c.args[0]) != c.server.auth {
		c.isAuthed = false
		return c.writer.Error(terror.ErrAuthFailed)
	} else {
		c.isAuthed = true
		return c.writer.String("OK")
	}
}

func quitCommand(c *Client) error {
	c.close()
	return c.writer.String("OK")
}
