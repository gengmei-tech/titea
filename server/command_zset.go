package server

import (
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"strconv"
	"strings"
)

// redis response, empty list or empty set
var emptyListOrSet = make([][]byte, 0)

func zaddCommand(c *Client) error {
	if c.argc%2 == 0 {
		return c.writer.Error(terror.ErrCmdParams)
	}
	var members []store.ZMember
	for i := 1; i < c.argc; i += 2 {
		score, err := strconv.ParseFloat(string(c.args[i]), 64)
		if err != nil {
			return c.writer.Error(terror.ErrCmdParams)
		}
		member := store.ZMember{Field: c.args[i+1], Value: c.args[i], Score: score}
		members = append(members, member)
	}
	if len(members) == 0 {
		return c.writer.Integer(0)
	}
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err := zset.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	addCnt, err := zset.Add(members...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(addCnt))
}

func zcardCommand(c *Client) error {
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = zset.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	return c.writer.Integer(int64(zset.Card()))
}

// inclusize ranges
func zrangeCommand(c *Client) error {
	return zrangeGenericCommand(c, false)
}

// inclusize ranges
func zrevrangeCommand(c *Client) error {
	return zrangeGenericCommand(c, true)
}

func zrangeGenericCommand(c *Client, reverse bool) error {
	start, err := strconv.ParseInt(string(c.args[1]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	stop, err := strconv.ParseInt(string(c.args[2]), 10, 64)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	var withscores = false
	if c.argc > 3 {
		if withscores, err = transferWithscores(c); err != nil {
			return c.writer.Error(terror.ErrCmdParams)
		}
	}
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = zset.ExistsForRead(); err != nil {
		return c.writer.Array(emptyListOrSet)
	}
	start, limit, ok := transferStartEnd(int64(zset.Card()), start, stop, reverse)
	if !ok {
		return c.writer.Array(emptyListOrSet)
	}
	result, err := zset.Range(uint64(start), uint64(limit), withscores)
	if err != nil {
		return c.writer.Error(err)
	}
	if result == nil {
		return c.writer.Array(emptyListOrSet)
	}
	if reverse {
		var i = 0
		var j = len(result) - 1
		if withscores {
			j = j - 1
			for i < j {
				result[i], result[j] = result[j], result[i]
				i++
				j++
				result[i], result[j] = result[j], result[i]
				i++
				j -= 3
			}
		} else {
			for i < j {
				result[i], result[j] = result[j], result[i]
				i++
				j--
			}
		}
	}
	return c.writer.Array(result)
}

func zscoreCommand(c *Client) error {
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = zset.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	score, err := zset.Score(c.args[1])
	if err != nil {
		return c.writer.Error(err)
	}
	if score == nil {
		return c.writer.Null()
	}
	return c.writer.BulkByte(score)
}

func zremCommand(c *Client) error {
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = zset.ExistsForRead(); err != nil {
		return c.writer.Integer(0)
	}
	delCnt, err := zset.Remove(c.args[1:]...)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.Integer(int64(delCnt))
}

// zrank key member
func zrankCommand(c *Client) error {
	return zrankGenericCommand(c, false)
}

func zrevrankCommand(c *Client) error {
	return zrankGenericCommand(c, true)
}

func zrankGenericCommand(c *Client, reverse bool) error {
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err = zset.ExistsForRead(); err != nil {
		return c.writer.Null()
	}
	index, err := zset.Rank(c.args[1], reverse)
	if err != nil {
		return c.writer.Null()
	}
	return c.writer.Integer(int64(index))
}

// zincrby key increment member
func zincrbyCommand(c *Client) error {
	score, err := strconv.ParseFloat(string(c.args[1]), 74)
	if err != nil {
		return c.writer.Error(terror.ErrCmdParams)
	}
	zset, err := store.InitZSet(c.environ, c.store, c.args[0])
	if err != nil {
		return c.writer.Error(err)
	}
	if err := zset.ExistsForWrite(); err != nil {
		return c.writer.Error(err)
	}
	value, err := zset.Incrby(c.args[2], score)
	if err != nil {
		return c.writer.Error(err)
	}
	return c.writer.BulkByte(value)
}

// return start count ok
func transferStartEnd(total, start, end int64, reverse bool) (int64, int64, bool) {
	if start < 0 {
		start += total
		if start < 0 {
			start = 0
		}
	} else if start > total {
		return 0, 0, false
	}

	if end < 0 {
		end += total
		if end < 0 {
			return 0, 0, false
		}
	} else if end > total {
		end = total - 1
	}

	if start > end {
		return 0, 0, false
	}

	if !reverse {
		return start, end - start + 1, true
	}
	start, end = total-end-1, total-start
	return start, end - start, true
}

// get withscores fields
func transferWithscores(c *Client) (bool, error) {
	if c.argc == 4 {
		if strings.ToLower(string(c.args[3])) == "withscores" {
			return true, nil
		}
	} else if c.argc == 7 {
		if strings.ToLower(string(c.args[3])) == "withscores" {
			return true, nil
		} else if strings.ToLower(string(c.args[6])) == "withscores" {
			return true, nil
		}
	}
	return false, nil
}

//// parse limit off
//func transferOffsetLimit(c *Client, total uint64) (uint64, uint64, bool, error) {
//	var (
//		index	uint8
//		offset	uint64
//		limit	uint64
//		err		error
//	)
//
//	for i, c := range c.args[3:] {
//		if strings.ToLower(string(c)) == "limit" {
//			index = uint8(i+3)
//		}
//	}
//
//	fmt.Println("total:", total)
//	fmt.Println("index:", index)
//
//	if index < 2 || index > 5 {
//		return 0, 0, false, nil
//	}
//
//	if offset, err = strconv.ParseUint(string(c.args[index+1]), 10, 64); err != nil {
//		return 0, 0, false, err
//	}
//
//	if limit, err = strconv.ParseUint(string(c.args[index+2]), 10, 64); err != nil {
//		return 0, 0, false, err
//	}
//
//	if offset < 0 {
//		offset = 0
//	}
//
//	if limit < 0 {
//		return 0, 0, false, nil
//	}
//
//	if offset > total {
//		return 0, 0, false, nil
//	}
//
//	if limit > total {
//		limit = total
//	}
//	return offset, limit, true, nil
//}

// score -inf +inf
//func transferScore(client *Client, value []byte) ([]byte, bool, error) {
//	var (
//		key		 		[]byte
//		score			float64
//		withFrontier 	bool
//		err			 	error
//	)
//	strScore := strings.ToLower(string(value))
//	switch strScore {
//		case "-inf":
//			key = nil
//			withFrontier = true
//		case "+inf":
//			key = nil
//			withFrontier = true
//		default:
//			if string(value[0]) == "(" {
//				if score, err = strconv.ParseFloat(string(value[1:]), 64); err != nil {
//					return nil, false, err
//				}
//				key = client.codec.EncodeZScoreKey(client.args[0], nil, score)
//				withFrontier = false
//			} else {
//				if score, err = strconv.ParseFloat(strScore, 64); err != nil {
//					return nil, false, err
//				}
//				key = client.codec.EncodeZScoreKey(client.args[0], nil, score)
//				withFrontier = true
//		}
//	}
//	return key, withFrontier, nil
//}
