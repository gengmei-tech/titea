package store

import (
	"bytes"
	"github.com/gengmei-tech/titea/pkg/util"
	"github.com/gengmei-tech/titea/pkg/util/number"
	"github.com/gengmei-tech/titea/server/types"
	"time"
)

// Meta contain mete info about a key
type Meta struct {
	// uuid uinque id
	ID []byte

	// redis data type
	Type byte

	// member count for hash,list,set,zset
	Count uint64

	// create time
	CreateAt uint64

	// expire time
	ExpireAt uint64

	// extra info, for string, it's value; for list, it's index;
	Extra []byte
}

// NewMeta init
func NewMeta(kind byte) *Meta {
	t := time.Now().Unix()
	m := Meta{
		ID:       util.UUID(),
		Type:     kind,
		ExpireAt: 0,
		Count:    0,
		CreateAt: uint64(t),
		Extra:    nil,
	}
	return &m
}

// Reset meta
func (m *Meta) Reset() {
	m.ID = util.UUID()
	m.Count = 0
	m.CreateAt = uint64(time.Now().Unix())
	m.ExpireAt = 0
	m.Extra = nil
}

// CheckIfExpire return true if expired else false
func (m *Meta) CheckIfExpire() bool {
	if m.ExpireAt > 0 && m.ExpireAt < uint64(time.Now().UnixNano()/1000/1000) {
		return true
	}
	return false
}

// CheckType check whether type is valid
func (m *Meta) CheckType(kind byte) bool {
	return m.Type == kind
}

// StringType return redis string type ,eg string,hash
func (m *Meta) StringType() string {
	t, _ := types.TYPES[m.Type]
	return t
}

// EncodeMetaKey header|m|key
func EncodeMetaKey(header []byte, key []byte) []byte {
	var buffer []byte
	buffer = append(buffer, header...)
	buffer = append(buffer, types.META)
	buffer = append(buffer, key...)
	return buffer
}

// EncodeMetaPrefix header|m
func EncodeMetaPrefix(header []byte) []byte {
	var buffer []byte
	buffer = append(buffer, header...)
	buffer = append(buffer, types.META)
	return buffer
}

// DecodeMetaKey header|m|key return key
func DecodeMetaKey(header []byte, key []byte) []byte {
	key = bytes.TrimPrefix(key, header)
	return key[1:]
}

// EncodeMetaValue meta value that save to TiKv
//type(1)|uuid(16)|count(8)|createAt(8)|expireAt(8)|extra
func (m *Meta) EncodeMetaValue() []byte {
	total := 17 + 24 + len(m.Extra)
	buffer := make([]byte, total)
	buffer[0] = m.Type
	copy(buffer[1:], m.ID)
	copy(buffer[17:], number.Uint64ToBytes(m.Count))
	copy(buffer[25:], number.Uint64ToBytes(m.CreateAt))
	copy(buffer[33:], number.Uint64ToBytes(m.ExpireAt))
	copy(buffer[41:], m.Extra)
	return buffer
}

// DecodeMetaValue  return Meta Struct
func DecodeMetaValue(value []byte) *Meta {
	var meta Meta
	meta.Type = value[0]
	meta.ID = value[1:17]
	meta.Count = number.BytesToUint64(value[17:25])
	meta.CreateAt = number.BytesToUint64(value[25:33])
	meta.ExpireAt = number.BytesToUint64(value[33:41])
	meta.Extra = value[41:]
	return &meta
}
