package types

const (
	// STRING string
	STRING byte = 'r'
	// HASH hash
	HASH byte = 'h'
	// LIST list
	LIST byte = 'l'
	// SET set
	SET byte = 's'
	// ZSET zset
	ZSET byte = 'z'
	// META meta
	META byte = 'm'
	// GC gc
	GC byte = 'g'
	// DATA data
	DATA byte = 'd'
	// EXPIRE expire
	EXPIRE byte = 'e'
)

// command type
const (
	// CSTRING string command
	CSTRING byte = STRING
	// CLIST list command
	CLIST byte = LIST
	// CHSAH hash command
	CHSAH byte = HASH
	// CSET set command
	CSET byte = SET
	// CZSET zset command
	CZSET byte = ZSET
	// CKEY key command
	CKEY byte = 'k'
	// CCONN connection command
	CCONN byte = 'c'
	// CSERVER server command
	CSERVER byte = 'v'
	// CUDC user defined command
	CUDC byte = 'u'
)

const (
	// SSTRING string
	SSTRING = "string"
	// SHASH hash
	SHASH = "hash"
	// SLIST list
	SLIST = "list"
	// SSET set
	SSET = "set"
	// SZSET zset
	SZSET = "zset"
)

// TYPES redis type
var TYPES = map[byte]string{
	STRING: SSTRING,
	HASH:   SHASH,
	LIST:   SLIST,
	SET:    SSET,
	ZSET:   SZSET,
}

const (
	// FlagRead read command
	FlagRead byte = 'r'
	// FlagWrite write command
	FlagWrite byte = 'w'
	// FlagPlain no read no write
	FlagPlain byte = 'p'
)

// SystemPrefix system prefix
const SystemPrefix = "kv"

// StringType return redis type
func StringType(t byte) string {
	s, _ := TYPES[t]
	return s
}
