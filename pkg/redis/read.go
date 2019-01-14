package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	okReply   interface{} = "OK"
	pongReply interface{} = "PONG"
)

// Reader represent a reader for RESP or telnet commands.
type Reader struct {
	rd    *bufio.Reader
	buf   []byte
	start int
	end   int
	data  []byte
}

// NewReader returns a command reader which will read RESP or telnet commands.
func NewReader(rd io.Reader) *Reader {
	return &Reader{
		rd:  bufio.NewReader(rd),
		buf: make([]byte, 4096),
	}
}

// Parse RESP
func (r *Reader) Parse() (interface{}, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("short resp line")
	}
	switch line[0] {
	case '+':
		switch {
		case len(line) == 3 && line[1] == 'O' && line[2] == 'K':
			// Avoid allocation for frequent "+OK" response.
			return okReply, nil
		case len(line) == 5 && line[1] == 'P' && line[2] == 'O' && line[3] == 'N' && line[4] == 'G':
			// Avoid allocation in PING command benchmarks :)
			return pongReply, nil
		default:
			return string(line[1:]), nil
		}
	case '-':
		return string(line[1:]), nil
	case ':':
		n, err := parseInt(line[1:])
		return n, err
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(r.rd, p)
		if err != nil {
			return nil, err
		}
		if line, err := r.ReadLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, errors.New("bad bulk string format")
		}
		return p, nil
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		b := make([]interface{}, n)
		for i := range b {
			b[i], err = r.Parse()
			if err != nil {
				return nil, err
			}
		}
		return b, nil
	}
	return nil, errors.New("unexpected response line")
}

// ParseRequest Parse client -> server command request, must be array of bulk strings
func (r *Reader) ParseRequest() ([][]byte, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("short resp line")
	}
	switch line[0] {
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		b := make([][]byte, n)
		for i := range b {
			b[i], err = r.ParseBulk()
			if err != nil {
				return nil, err
			}
		}
		return b, nil
	default:
		return nil, fmt.Errorf("not invalid array of bulk string type, but %c", line[0])
	}
}

// ReadLine read request
func (r *Reader) ReadLine() ([]byte, error) {
	p, err := r.rd.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return nil, errors.New("long resp line")
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, errors.New("bad resp line terminator")
	}
	return p[:i], nil
}

// parseLen parses bulk string and array lengths.
func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, errors.New("malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, errors.New("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// parseInt parses an integer reply.
func parseInt(p []byte) (int64, error) {
	if len(p) == 0 {
		return 0, errors.New("malformed integer")
	}

	var negate bool
	if p[0] == '-' {
		negate = true
		p = p[1:]
		if len(p) == 0 {
			return 0, errors.New("malformed integer")
		}
	}

	var n int64
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, errors.New("illegal bytes in length")
		}
		n += int64(b - '0')
	}

	if negate {
		n = -n
	}
	return n, nil
}

// ParseBulk parse bulk
func (r *Reader) ParseBulk() ([]byte, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("short resp line")
	}
	switch line[0] {
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(r.rd, p)
		if err != nil {
			return nil, err
		}
		if line, err := r.ReadLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, errors.New("bad bulk string format")
		}
		return p, nil
	default:
		return nil, fmt.Errorf("not invalid bulk string type, but %c", line[0])
	}
}
