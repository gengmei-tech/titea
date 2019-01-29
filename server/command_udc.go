package server

import (
	"bytes"
	"github.com/gengmei-tech/titea/server/namespace"
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	"strconv"
)

// user defined command
// register group.service dbindex creator, registre map from redis dbindex to namespace
func registerCommand(c *Client) error {
	dbindex, err := strconv.ParseUint(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	if dbindex == 0 {
		return c.writer.Error(terror.ErrCmdParams)
	}
	if err = namespace.RegisterNamespace(c.store, string(c.args[0]), string(c.args[2]), dbindex); err != nil {
		return c.writer.Error(err)
	}
	return c.writer.String("OK")
}

// flush with prefix
func flushCommand(c *Client) error {
	if bytes.HasPrefix(c.args[0], []byte(types.SystemPrefix)) {
		svr := store.InitServer(c.store)
		svr.FlushPrefix(c.args[0])
		return c.writer.String("OK")
	}
	return c.writer.Error(terror.ErrCmdParams)
}

func scanCommand(c *Client) error {
	if !bytes.HasPrefix(c.args[0], []byte(types.SystemPrefix)) {
		return c.writer.Error(terror.ErrCmdParams)
	}
	var offset uint64
	var err error
	if c.argc == 2 {
		offset, err = strconv.ParseUint(string(c.args[1]), 10, 64)
		if err != nil {
			return c.writer.Error(err)
		}
	}
	svr := store.InitServer(c.store)
	keys, err := svr.ScanPrefix(c.args[0], offset)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(keys)
}

// count with prefix
func countCommand(c *Client) error {
	if !bytes.HasPrefix(c.args[0], []byte(types.SystemPrefix)) {
		return c.writer.Error(terror.ErrCmdParams)
	}
	svr := store.InitServer(c.store)
	total, _ := svr.Count(c.args[0])
	return c.writer.Integer(int64(total))
}
