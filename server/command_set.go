package server

import (
	"github.com/deckarep/golang-set"
	"github.com/gengmei-tech/titea/server/store"
)

const (
	opDiff = iota
	opInter
	opUnion
)

func saddCommand(c *Client) error {
	set, err := store.InitSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = set.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	addCnt, err := set.Add(c.args[1:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(addCnt))
}

func scardCommand(c *Client) error {
	set, err := store.InitSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = set.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	return c.writer.Integer(int64(set.Card()))
}

func sismemberCommand(c *Client) error {
	set, err := store.InitSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = set.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	status, err := set.IsMember(c.args[1])
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(status))
}

func smembersCommand(c *Client) error {
	set, err := store.InitSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = set.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	result, err := set.Members()
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(result)
}

func sremCommand(c *Client) error {
	set, err := store.InitSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = set.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	delCnt, err := set.Remove(c.args[1:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(delCnt))
}

func sdiffCommand(c *Client) error {
	result, err := sopGeneric(c, opDiff, c.args...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(result)
}

func sunionCommand(c *Client) error {
	result, err := sopGeneric(c, opUnion, c.args...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(result)
}

func sinterCommand(c *Client) error {
	result, err := sopGeneric(c, opInter, c.args...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Array(result)
}

func sopGeneric(c *Client, opType int, keys ...[]byte) ([][]byte, error) {
	var sets []mapset.Set
	for _, key := range keys {
		set, err := store.InitSet(c.environ, c.store, key)
		if err != nil {
			return nil, err
		}
		if err = set.ExistsForRead(); err != nil {
			sets = append(sets, mapset.NewSet())
		}
		members, err := set.Members()
		if err != nil {
			return nil, err
		}
		if members == nil {
			sets = append(sets, mapset.NewSet())
		}
		mem := make([]interface{}, len(members))
		for i, m := range members {
			mem[i] = string(m)
		}
		sets = append(sets, mapset.NewSetFromSlice(mem))
	}
	result := sets[0]
	for _, ms := range sets[1:] {
		switch opType {
		case opDiff:
			result = result.Difference(ms)
			break
		case opInter:
			result = result.Intersect(ms)
			break
		case opUnion:
			result = result.Union(ms)
			break
		}
	}
	sls := result.ToSlice()
	values := make([][]byte, len(sls))
	for i, val := range sls {
		values[i] = []byte(val.(string))
	}
	return values, nil
}
