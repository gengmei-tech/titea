package number

import (
	"bytes"
	"encoding/binary"
	"math"
)

// BytesToInt64 transfer byte to int64
func BytesToInt64(n []byte) int64 {
	return int64(binary.BigEndian.Uint64(n))
}

// BytesToUint64 transfer byte to uint64
func BytesToUint64(n []byte) uint64 {
	return binary.BigEndian.Uint64(n)
}

// Uint64ToBytes transfer uint64 to byte
func Uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

// Uint32ToBytes transfer uint32 to byte
func Uint32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	return b
}

// Int64ToBytes transfer int64 to byte
func Int64ToBytes(n int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

// Float64ToUint64 transfer float64 to uint64
func Float64ToUint64(f float64) uint64 {
	u := math.Float64bits(f)
	if f >= 0 {
		u |= 0x8000000000000000
	} else {
		u = ^u
	}
	return u
}

// Uint64ToFloat64 transfer uint64 to float64
func Uint64ToFloat64(u uint64) float64 {
	if u&0x8000000000000000 > 0 {
		u &= ^uint64(0x8000000000000000)
	} else {
		u = ^u
	}
	return math.Float64frombits(u)
}
