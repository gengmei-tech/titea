package server

import (
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
)

func lpushCommand(c *Client) error {
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	addCnt, err := list.LPush(c.args[1:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(addCnt))
}

func lpopCommand(c *Client) error {
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	value, err := list.LPop()
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.BulkByte(value)
}

func rpushCommand(c *Client) error {
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	addCnt, err := list.RPush(c.args[1:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(addCnt))
}

func rpopCommand(c *Client) error {
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	value, err := list.RPop()
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.BulkByte(value)
}

func llenCommand(c *Client) error {
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	return c.writer.Integer(int64(list.Len()))
}

// index begin with 0
func lindexCommand(c *Client) error {
	index, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	if index < 0 {
		index = index + int64(list.Len())
		// out of bound
		if index < 0 {
			return c.writer.Null()
		}
	}
	// out of bound
	if index >= int64(list.Len()) {
		return c.writer.Null()
	}
	value, err := list.Index(index)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.BulkByte(value)
}

// start begin with 0
func lrangeComamnd(c *Client) error {
	start, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	end, err := strconv.ParseInt(string(c.args[2]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForRead(); err != nil {
		return c.writer.Array(emptyListOrSet)
	}

	if start < 0 {
		start = start + int64(list.Len())
		if start < 0 {
			return c.writer.Array(emptyListOrSet)
		}
	} else if start >= int64(list.Len()) {
		return c.writer.Array(emptyListOrSet)
	}

	if end < 0 {
		end = end + int64(list.Len())
		if end < 0 {
			return c.writer.Array(emptyListOrSet)
		}
	} else if end >= int64(list.Len()) {
		end = int64(list.Len()) - 1
	}

	// here start and stop both be positive
	if start > end {
		return c.writer.Array(emptyListOrSet)
	}
	limit := end - start + 1
	result, err := list.Range(uint64(start), uint64(limit))
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(result)
}

func lsetCommand(c *Client) error {
	index, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}

	list, err := store.InitList(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = list.ExistsForRead(); err != nil {
		return c.writer.Error(err)
	}
	if index < 0 {
		index = index + int64(list.Len())
		if index < 0 {
			return c.writer.Null()
		}
	} else if index >= int64(list.Len()) {
		return c.writer.Null()
	}
	if err = list.Set(index, c.args[2]); err != nil {
		return c.writer.Error(err)
	}
	return c.writer.String("OK")
}
