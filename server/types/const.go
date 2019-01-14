package types

// Type in tikv
const (
	STRING byte = 'r'
	HASH   byte = 'h'
	LIST   byte = 'l'
	SET    byte = 's'
	ZSET   byte = 'z'
	META   byte = 'm'
	GC     byte = 'g'
	DATA   byte = 'd'
	EXPIRE byte = 'e'
)

// Command type
const (
	CSTRING byte = STRING
	CLIST   byte = LIST
	CHSAH   byte = HASH
	CSET    byte = SET
	CZSET   byte = ZSET
	CKEY    byte = 'k'
	CCONN   byte = 'c'
	CSERVER byte = 'v'
	CUDC    byte = 'u'
)

// Command type string
const (
	SSTRING = "string"
	SHASH   = "hash"
	SLIST   = "list"
	SSET    = "set"
	SZSET   = "zset"
)

// TYPES redis type
var TYPES = map[byte]string{
	STRING: SSTRING,
	HASH:   SHASH,
	LIST:   SLIST,
	SET:    SSET,
	ZSET:   SZSET,
}

// Command Flag
const (
	FlagRead  byte = 'r'
	FlagWrite byte = 'w'
	FlagPlain byte = 'p'
)

// SystemPrefix system prefix
const SystemPrefix = "kv"

// StringType return redis type
func StringType(t byte) string {
	s, _ := TYPES[t]
	return s
}
