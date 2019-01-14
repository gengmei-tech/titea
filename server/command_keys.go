package server

import (
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
	"time"
)

// 获取meta 根据meta的类型调用不同的删除函数 del key1 key2 key3 同一类型的key一起删除 返回成功删除的个数
func delCommand(c *Client) error {
	pk := store.InitKey(c.environ, c.store)
	delCnt, err := pk.Del(c.args[0:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(delCnt))
}

// exists key1 key2 返回exists的数量
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

// 返回0 || 1
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

// expireTime 为毫秒过期时间
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
//tip: 最大返回5000个元素 start 不支持负数
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
