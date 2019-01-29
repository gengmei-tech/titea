package server

import (
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
	"time"
)

func delCommand(c *Client) error {
	pk := store.InitKey(c.environ, c.store)
	delCnt, err := pk.Del(c.args[0:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(delCnt))
}

func existsCommand(c *Client) error {
	pk := store.InitKey(c.environ, c.store)
	count, err := pk.Exists(c.args[0:]...)
	if err != nil {
		return c.writer.Error(nil)
	}
	return c.writer.Integer(int64(count))
}

func typeCommand(c *Client) error {
	pk := store.InitKey(c.environ, c.store)
	t, err := pk.Type(c.args[0])
	if err != nil || t == nil {
		return c.writer.Null()
	}
	return c.writer.Byte(t)
}

func expireCommand(c *Client) error {
	second, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	expireTime := second*1000 + (time.Now().UnixNano() / 1000 / 1000)
	return expireGenericCommand(c, expireTime)
}

func expireatCommand(c *Client) error {
	expireAt, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	return expireGenericCommand(c, expireAt*1000)
}

func pexpireCommand(c *Client) error {
	msec, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	expireTime := msec + (time.Now().UnixNano() / 1000 / 1000)
	return expireGenericCommand(c, expireTime)
}

func pexpireatCommand(c *Client) error {
	expireAt, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	return expireGenericCommand(c, expireAt)
}

func ttlCommand(c *Client) error {
	pk := store.InitKey(c.environ, c.store)
	ttl, err := pk.TTL(c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if ttl <= 0 {
		return c.writer.Integer(int64(ttl))
	}
	return c.writer.Integer(ttl / 1000)
}

func pttlCommand(c *Client) error {
	pk := store.InitKey(c.environ, c.store)
	pttl, err := pk.TTL(c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(pttl)
}

// expireTime millisecond
func expireGenericCommand(c *Client, expireTime int64) error {
	pk := store.InitKey(c.environ, c.store)
	status, err := pk.Expire(c.args[0], expireTime)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(status))
}

//keys * start limit
//keys prefix start limit
//keys prefix*
//tip: max 5000
func keysCommand(c *Client) error {
	var start int64
	var limit int64
	var err error
	if c.argc > 1 {
		if start, err = strconv.ParseInt(string(c.args[1]), 10, 64); err != nil {
			return c.writer.Error(terror.ErrCmdParams)
		}
		if c.argc > 2 {
			if limit, err = strconv.ParseInt(string(c.args[2]), 10, 64); err != nil {
				return c.writer.Error(terror.ErrCmdParams)
			}
		}
	}
	pk := store.InitKey(c.environ, c.store)
	keys, err := pk.Keys(c.args[0], uint64(start), uint64(limit))
	if err != nil {
		return c.writer.Error(err)
	}
	if keys == nil {
		return c.writer.Null()
	}
	return c.writer.Array(keys)
}
