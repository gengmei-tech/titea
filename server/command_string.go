package server

import (
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
	"strings"
	"time"
)

func getCommand(c *Client) error {
	s := store.InitString(c.environ, c.store)
	value, err := s.Get(c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if value == nil {
		return c.writer.Null()
	}
	return c.writer.Byte(value)
}

func mgetCommand(c *Client) error {
	string := store.InitString(c.environ, c.store)
	values, err := string.MGet(c.args[0:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(values)
}

// Don't support type changed
func setCommand(c *Client) error {
	var (
		expireAt uint64
		isNX     = false
		isXX     = false
	)
	if c.argc > 2 {
		for i, p := range c.args[2:] {
			con := strings.ToLower(string(p))
			switch con {
			case "ex":
				sec, err := strconv.ParseInt(string(c.args[i+3]), 10, 64)
				if err != nil || sec < 0 {
					return c.writer.Error(terror.ErrCmdParams)
				}
				expireAt = uint64(sec+int64(time.Now().Unix())) * 1000
				break
			case "px":
				msec, err := strconv.ParseInt(string(c.args[i+3]), 10, 64)
				if err != nil || msec < 0 {
					return c.writer.Error(terror.ErrCmdParams)
				}
				expireAt = uint64(msec + (time.Now().UnixNano() / 1000 / 1000))
				break
			case "nx":
				isNX = true
				break
			case "xx":
				isXX = true
				break
			}
		}
	}
	s := store.InitString(c.environ, c.store)
	if err := s.Set(c.args[0], c.args[1], expireAt, isNX, isXX); err != nil {
		return c.writer.Null()
	}
	return c.writer.String("OK")
}

// set key seconds value
func setexCommand(c *Client) error {
	// expire > 0
	sec, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil || sec < 0 {
		return c.writer.Error(terror.ErrCmdParams)
	}
	expireAt := uint64(sec+int64(time.Now().Unix())) * 1000
	s := store.InitString(c.environ, c.store)
	if err = s.Set(c.args[0], c.args[1], expireAt, false, false); err != nil {
		return c.writer.Error(err)
	}
	return c.writer.String("OK")
}

func msetCommand(c *Client) error {
	if c.argc%2 != 0 {
		return c.writer.Error(terror.ErrCmdParams)
	}
	s := store.InitString(c.environ, c.store)
	keyValues := make(map[string][]byte)
	for i := 0; i < c.argc; i += 2 {
		keyValues[string(c.args[i])] = c.args[i+1]
	}
	if err := s.MSet(keyValues); err != nil {
		return c.writer.Error(err)
	}
	return c.writer.String("OK")
}

func setnxCommand(c *Client) error {
	s := store.InitString(c.environ, c.store)
	err := s.Set(c.args[0], c.args[1], 0, true, false)
	if err != nil {
		return c.writer.Integer(0)
	}
	return c.writer.Integer(1)
}

func getsetCommand(c *Client) error {
	s := store.InitString(c.environ, c.store)
	value, err := s.GetSet(c.args[0], c.args[1])
	if err != nil {
		return c.writer.Error(err)
	}
	if value == nil {
		return c.writer.Null()
	}
	return c.writer.BulkByte(value)
}

func incrCommand(c *Client) error {
	return incrGenericCommand(c, 1)
}

func incrbyCommand(c *Client) error {
	step, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	return incrGenericCommand(c, step)
}

func decrCommand(c *Client) error {
	return incrGenericCommand(c, -1)
}

func decrbyCommand(c *Client) error {
	step, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	return incrGenericCommand(c, step*-1)
}

func strlenCommand(c *Client) error {
	s := store.InitString(c.environ, c.store)
	lens, err := s.Strlen(c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(lens))
}

func incrGenericCommand(c *Client, step int64) error {
	s := store.InitString(c.environ, c.store)
	num, err := s.Incr(c.args[0], step)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(num))
}
