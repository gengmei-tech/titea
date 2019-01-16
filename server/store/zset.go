package store

import (
	"bytes"
	"github.com/gengmei-tech/titea/pkg/util/number"
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	"strconv"
)

const (
	// WITHSCORE zset key contain score
	WITHSCORE = byte(1)

	// WITHOUTSCORE zset key not contain score
	WITHOUTSCORE = byte(0)
)

// ZSet implement redis zset operation
type ZSet struct {
	store *Store

	// environ
	environ *Environ

	// hash key
	key []byte

	// metaKey
	metaKey []byte

	// Meta
	meta *Meta
}

// ZMember zset member
type ZMember struct {
	Field []byte
	Value []byte
	Score float64
}

// InitZSet init
func InitZSet(environ *Environ, store *Store, key []byte) (*ZSet, error) {
	mkey := EncodeMetaKey(environ.Header, key)
	value, err := store.Get(mkey)
	if err != nil {
		return nil, err
	}
	zset := ZSet{
		store:   store,
		environ: environ,
		key:     key,
		metaKey: mkey,
	}
	if value != nil {
		meta := DecodeMetaValue(value)
		if !meta.CheckType(types.ZSET) {
			return nil, terror.ErrTypeNotMatch
		}
		zset.meta = meta
	}
	return &zset, nil
}

// ExistsForWrite exist for write or not exist
func (zset *ZSet) ExistsForWrite() error {
	if zset.meta == nil {
		zset.meta = NewMeta(types.ZSET)
	} else {
		if zset.meta.CheckIfExpire() {
			if err := zset.destroy(); err != nil {
				return err
			}
			// reset meta
			zset.meta.Reset()
		}
	}
	return nil
}

// ExistsForRead exist for read or not exist
func (zset *ZSet) ExistsForRead() error {
	if zset.meta == nil {
		return terror.ErrKeyNotExist
	}
	if zset.meta.CheckIfExpire() {
		zset.store.WriteReset()
		// 删除meta
		if err := zset.store.Delete(zset.metaKey); err != nil {
			return err
		}
		if err := zset.destroy(); err != nil {
			return err
		}
		if err := zset.store.Commit(); err != nil {
			zset.environ.FailedTxn("ExistsForRead")
			return err
		}
		return terror.ErrKeyNotExist
	}
	return nil
}

// EncodeDataKey header|type(1)|uuid|withoutScore(1)|member
func (zset *ZSet) EncodeDataKey(member []byte) []byte {
	buffer := make([]byte, zset.environ.HeaderLen+18+len(member))
	copy(buffer[0:], zset.environ.Header)
	buffer[zset.environ.HeaderLen] = types.DATA
	copy(buffer[zset.environ.HeaderLen+1:], zset.meta.ID)
	buffer[zset.environ.HeaderLen+17] = WITHOUTSCORE
	copy(buffer[zset.environ.HeaderLen+18:], member)
	return buffer
}

// DecodeDataKey header|d|uuid|withoutScore(1)|member
func (zset *ZSet) DecodeDataKey(key []byte) []byte {
	index := zset.environ.HeaderLen + 18
	return key[index:]
}

// EncodeScoreKey header|d|uuid|withScore(1)|score(8)|field
func (zset *ZSet) EncodeScoreKey(member []byte, score float64) []byte {
	buffer := make([]byte, zset.environ.HeaderLen+26+len(member))
	copy(buffer[0:], zset.environ.Header)
	buffer[zset.environ.HeaderLen] = types.DATA
	copy(buffer[zset.environ.HeaderLen+1:], zset.meta.ID)
	index := zset.environ.HeaderLen + 17
	buffer[index] = WITHSCORE
	index++
	// 8字节
	copy(buffer[index:], number.Uint64ToBytes(number.Float64ToUint64(score)))
	index = index + 8
	copy(buffer[index:], member)
	return buffer
}

// DecodeScoreKey header|d|uuid|withScore(1)|score(8)|member
func (zset *ZSet) DecodeScoreKey(key []byte) ([]byte, float64) {
	index := zset.environ.HeaderLen + 18
	score := number.Uint64ToFloat64(number.BytesToUint64(key[index : index+8]))
	member := key[index+8:]
	return member, score
}

// Prefix header|d|uuid
func (zset *ZSet) Prefix() []byte {
	buffer := make([]byte, zset.environ.HeaderLen+17)
	copy(buffer[0:], zset.environ.Header)
	buffer[zset.environ.HeaderLen] = types.DATA
	copy(buffer[zset.environ.HeaderLen+1:], zset.meta.ID)
	return buffer
}

// ScorePrefix header|d|uuid|withScore
func (zset *ZSet) ScorePrefix() []byte {
	buffer := make([]byte, zset.environ.HeaderLen+18)
	copy(buffer[0:], zset.environ.Header)
	buffer[zset.environ.HeaderLen] = types.DATA
	copy(buffer[zset.environ.HeaderLen+1:], zset.meta.ID)
	index := zset.environ.HeaderLen + 17
	buffer[index] = WITHSCORE
	return buffer
}

// Add members
func (zset *ZSet) Add(members ...ZMember) (uint64, error) {
	keyValues := make(map[string]ZMember)
	dataKeys := make([][]byte, len(members))
	for i, member := range members {
		dataKey := zset.EncodeDataKey(member.Field)
		dataKeys[i] = dataKey
		keyValues[string(dataKey)] = member
	}
	result, err := zset.store.BatchGet(dataKeys)
	if err != nil {
		return 0, err
	}
	var addCnt uint64
	for key, member := range keyValues {
		scoreKey := zset.EncodeScoreKey(member.Field, member.Score)
		if val, ok := result[key]; ok {
			if bytes.Compare(val, member.Value) != 0 {
				// score 变化了 删除之前的score
				oldScore, err := strconv.ParseFloat(string(val), 64)
				if err != nil {
					return 0, err
				}
				oldScoreKey := zset.EncodeScoreKey(member.Field, oldScore)
				if err = zset.store.Delete(oldScoreKey); err != nil {
					return 0, err
				}
				if err := zset.store.Set(scoreKey, member.Value); err != nil {
					return 0, err
				}
			}
		} else {
			if err := zset.store.Set([]byte(key), member.Value); err != nil {
				return 0, err
			}
			if err := zset.store.Set(scoreKey, member.Value); err != nil {
				return 0, err
			}
			addCnt++
		}
	}
	if addCnt > 0 {
		zset.meta.Count += addCnt
		if err := zset.store.Set(zset.metaKey, zset.meta.EncodeMetaValue()); err != nil {
			return 0, err
		}
	}
	if err := zset.store.Commit(); err != nil {
		zset.environ.FailedTxn("Add")
		return 0, err
	}
	return addCnt, nil
}

// Incrby incr score
func (zset *ZSet) Incrby(member []byte, step float64) ([]byte, error) {
	dataKey := zset.EncodeDataKey(member)
	value, err := zset.store.Get(dataKey)
	if err != nil {
		return nil, err
	}
	var score float64
	if value == nil {
		score = step
	} else {
		score, err := strconv.ParseFloat(string(value), 64)
		oldScoreKey := zset.EncodeScoreKey(member, score)
		if err = zset.store.Delete(oldScoreKey); err != nil {
			return nil, err
		}
		score += step
	}
	scoreKey := zset.EncodeScoreKey(member, score)
	value = []byte(strconv.FormatFloat(step, 'f', 2, 64))
	if err = zset.store.Set(dataKey, value); err != nil {
		return nil, err
	}
	if err = zset.store.Set(scoreKey, value); err != nil {
		return nil, err
	}
	if err := zset.store.Commit(); err != nil {
		zset.environ.FailedTxn("Incrby")
		return nil, err
	}
	return nil, nil
}

// Card count members
func (zset *ZSet) Card() uint64 {
	return zset.meta.Count
}

// Score get member score
func (zset *ZSet) Score(member []byte) ([]byte, error) {
	dataKey := zset.EncodeDataKey(member)
	value, err := zset.store.Get(dataKey)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Remove remove member from zset
func (zset *ZSet) Remove(members ...[]byte) (uint64, error) {
	dataKeys := make([][]byte, len(members))
	keyValues := make(map[string][]byte)
	for i, member := range members {
		dataKey := zset.EncodeDataKey(member)
		dataKeys[i] = dataKey
		keyValues[string(dataKey)] = member
	}
	result, err := zset.store.BatchGet(dataKeys)
	if err != nil {
		return 0, err
	}
	var delCnt uint64
	for key, value := range keyValues {
		if val, ok := result[key]; ok {
			// 存在
			score, err := strconv.ParseFloat(string(val), 64)
			if err != nil {
				return 0, err
			}
			scoreKey := zset.EncodeScoreKey(value, score)
			if err = zset.store.Delete(scoreKey); err != nil {
				return 0, err
			}
			if err = zset.store.Delete([]byte(key)); err != nil {
				return 0, err
			}
			delCnt++
		}
	}
	if delCnt > 0 {
		zset.meta.Count -= delCnt
		if zset.meta.Count <= 0 {
			if err = zset.store.Delete(zset.metaKey); err != nil {
				return 0, err
			}
		} else {
			if err = zset.store.Set(zset.metaKey, zset.meta.EncodeMetaValue()); err != nil {
				return 0, err
			}
		}
	}
	if err := zset.store.Commit(); err != nil {
		zset.environ.FailedTxn("Remove")
		return 0, err
	}
	return delCnt, nil
}

// Range return member within range, begin with 0, offset default 0
func (zset *ZSet) Range(offset uint64, limit uint64, withScores bool) ([][]byte, error) {
	scorePrefix := zset.ScorePrefix()
	result, err := zset.store.ScanAll(scorePrefix, nil, offset, limit)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	var values [][]byte
	for i := 0; i < len(result); i = i + 2 {
		member, _ := zset.DecodeScoreKey(result[i])
		values = append(values, member)
		if withScores {
			values = append(values, result[i+1])
		}
	}
	return values, nil
}

// Rank O(n)
func (zset *ZSet) Rank(member []byte, reverse bool) (uint64, error) {
	dataKey := zset.EncodeDataKey(member)
	value, err := zset.store.Get(dataKey)
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, terror.ErrKeyNotValid
	}
	score, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return 0, err
	}
	scorePrefix := zset.ScorePrefix()
	scoreKey := zset.EncodeScoreKey(member, score)
	index, err := zset.store.Index(scorePrefix, scoreKey)
	if err != nil {
		return 0, err
	}
	if index > zset.meta.Count {
		return 0, terror.ErrOutOfIndex
	}
	if reverse {
		index = zset.meta.Count - index - 1
	}
	return index, nil
}

// add to gc and remove from expire set
func (zset *ZSet) destroy() error {
	// add to gc
	if err := AddToGc(zset.environ, zset.store, zset.meta.ID); err != nil {
		return err
	}
	// remove from expire set
	return RemoveExpire(zset.environ, zset.store, zset.key, zset.meta.ExpireAt)
}
