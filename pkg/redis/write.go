package redis

import (
	"bufio"
	"bytes"
	"strconv"
)

// Writer allows for writing RESP messages.
type Writer struct {
	w         *bufio.Writer
	autoFlush bool
}

// NewWriter creates a new RESP writer.
func NewWriter(wr *bufio.Writer) *Writer {
	return &Writer{
		w:         wr,
		autoFlush: true,
	}
}

//Error builds a RESP error
func (r *Writer) Error(e error) error {
	_, err := r.w.Write([]byte("-" + e.Error() + "\r\n"))
	if r.autoFlush {
		r.w.Flush()
	}
	return err
}

// Null return null to client
func (r *Writer) Null() error {
	_, err := r.w.Write([]byte("$-1\r\n"))
	if r.autoFlush {
		r.w.Flush()
	}
	return err
}

// Byte single byte
func (r *Writer) Byte(b []byte) error {
	return r.String(string(b))
}

// BulkByte bulk bytes to response
func (r *Writer) BulkByte(b []byte) error {
	return r.BulkString(string(b))
}

//String builds a RESP simplestring
func (r *Writer) String(s string) error {
	_, err := r.w.Write([]byte("+" + s + "\r\n"))
	if r.autoFlush {
		r.w.Flush()
	}
	return err
}

// BulkString bulk string to response
func (r *Writer) BulkString(s string) error {
	length := strconv.Itoa(len(s))
	_, err := r.w.Write([]byte("$" + length + "\r\n" + s + "\r\n"))
	if r.autoFlush {
		r.w.Flush()
	}
	return err
}

// Integer builds a RESP integer
func (r *Writer) Integer(v int64) error {
	s := strconv.FormatInt(v, 10)
	_, err := r.w.Write([]byte(":" + s + "\r\n"))
	if r.autoFlush {
		r.w.Flush()
	}
	return err
}

// Array [][]byte to response
func (r *Writer) Array(data [][]byte) error {
	r.autoFlush = false
	s := strconv.Itoa(len(data))
	r.w.Write([]byte("*" + s + "\r\n"))
	for i := range data {
		if data[i] == nil {
			r.Null()
			continue
		}

		if EmptyCheck(data[i]) {
			r.Null()
			continue
		}
		r.BulkString(string(data[i]))
	}
	r.w.Flush()
	r.autoFlush = true
	return nil
}

// EmptyFill client set "", but tikv don't support, so fillEmpty
func EmptyFill() []byte {
	return []byte{0}
}

// EmptyCheck reverse EmptyFill
func EmptyCheck(val []byte) bool {
	return bytes.Equal(val, EmptyFill())
}
