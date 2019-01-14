package server

// hset| hmset 2个写操作, 如果遇到过期且主动过期还没有执行, 此时立即将meta的ExpireAt设置为0,
// 然后同步清除改key下的field, 这种情况发生的概率低, 故为同步且只能为同步; 如果异步, 则在执行删除的时候,有可能同步写的field已经写入,
// 此时再delete, 会将数据全部删掉, 故只能为同步;

import (
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
)

// 参数检测 统一在调用前 此处不再检测参数个数是否满足需求
// 字段不存在时返回nil
func hgetCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	// 不存在或者已经过期
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	value, err := hash.Field(c.args[1])
	if err != nil {
		return c.writer.Error(err)
	}
	if value == nil {
		return c.writer.Null()
	}
	return c.writer.BulkByte(value)
}

// 字段不存在时返回0
func hstrlenCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	// 不存在或者已经过期
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	value, err := hash.Field(c.args[1])
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(len(value)))
}

// 字段不存在返回0
func hexistsCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	// 不存在或者已经过期
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	value, err := hash.Field(c.args[1])
	if err != nil || value == nil {
		return c.writer.Integer(0)
	}
	return c.writer.Integer(1)
}

func hlenCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	// 不存在或者已经过期
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	return c.writer.Integer(int64(hash.Count()))
}

func hmgetCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	// 不存在或者已经过期
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	result, err := hash.Fields(c.args[1:])
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(result)
}

// 字段存在则返回0(重写) 否则返回1
func hsetCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	status, err := hash.Set(c.args[1], c.args[2])
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(status))
}

func hmsetCommand(c *Client) error {
	if (c.argc-1)%2 != 0 {
		return c.writer.Error(terror.ErrCmdParams)
	}
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	items := make(map[string][]byte)
	for i := 1; i < c.argc-1; i = i + 2 {
		items[string(c.args[i])] = c.args[i+1]
	}
	if _, err = hash.MSet(items); err != nil {
		return c.writer.Error(err)
	}
	return c.writer.String("OK")
}

func hkeysCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	keys, err := hash.Keys()
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(keys)
}

func hvalsCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	keys, err := hash.Values()
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(keys)
}

func hgetallCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	keys, err := hash.All()
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(keys)
}

func hdelCommand(c *Client) error {
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	delCnt, err := hash.Remove(c.args[1:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(delCnt))
}

//HINCRBY key field increment
func hincrbyCommand(c *Client) error {
	step, err := strconv.ParseInt(string(c.args[2]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	hash, err := store.InitHash(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = hash.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	num, err := hash.Incrby(c.args[1], step)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(num)
}
