package server

import "github.com/gengmei-tech/titea/server/types"

type cmdFunction func(c *Client) error

type command struct {
	Name     string
	Function cmdFunction
	Arity    int
	Flags    byte
	FirstKey int
	Step     int
	Type     byte
}

var commandTable = map[string]command{
	"get":    {"get", getCommand, 1, types.FlagRead, 0, 0, types.CSTRING},
	"set":    {"set", setCommand, -2, types.FlagWrite, 0, 0, types.CSTRING},
	"setex":  {"setex", setexCommand, 3, types.FlagWrite, 0, 0, types.CSTRING},
	"mget":   {"mget", mgetCommand, -1, types.FlagRead, 0, 1, types.CSTRING},
	"mset":   {"mset", msetCommand, -2, types.FlagWrite, 0, 2, types.CSTRING},
	"incr":   {"incr", incrCommand, 1, types.FlagWrite, 0, 0, types.CSTRING},
	"incrby": {"incrby", incrbyCommand, 2, types.FlagWrite, 0, 0, types.CSTRING},
	"decr":   {"decr", decrCommand, 1, types.FlagWrite, 0, 0, types.CSTRING},
	"decrby": {"decrby", decrbyCommand, 2, types.FlagWrite, 0, 0, types.CSTRING},
	"strlen": {"strlen", strlenCommand, 1, types.FlagRead, 0, 0, types.CSTRING},
	"setnx":  {"setnx", setnxCommand, 2, types.FlagWrite, 0, 0, types.CSTRING},
	"getset": {"getset", getsetCommand, 2, types.FlagWrite, 0, 0, types.CSTRING},

	"hget":    {"hget", hgetCommand, 2, types.FlagRead, 0, 0, types.CHSAH},
	"hstrlen": {"hstrlen", hstrlenCommand, 2, types.FlagRead, 0, 0, types.CHSAH},
	"hexists": {"hexists", hexistsCommand, 2, types.FlagRead, 0, 0, types.CHSAH},
	"hlen":    {"hlen", hlenCommand, 1, types.FlagRead, 0, 0, types.CHSAH},
	"hmget":   {"hmget", hmgetCommand, -2, types.FlagRead, 0, 0, types.CHSAH},
	"hset":    {"hset", hsetCommand, 3, types.FlagWrite, 0, 0, types.CHSAH},
	"hmset":   {"hmset", hmsetCommand, -3, types.FlagWrite, 0, 0, types.CHSAH},
	"hkeys":   {"hkeys", hkeysCommand, 1, types.FlagRead, 0, 0, types.CHSAH},
	"hvals":   {"hvals", hvalsCommand, 1, types.FlagRead, 0, 0, types.CHSAH},
	"hgetall": {"hgetall", hgetallCommand, 1, types.FlagRead, 0, 0, types.CHSAH},
	"hdel":    {"hdel", hdelCommand, -1, types.FlagWrite, 0, 0, types.CHSAH},
	"hincrby": {"hincrby", hincrbyCommand, 3, types.FlagWrite, 0, 0, types.CHSAH},

	"sadd":      {"sadd", saddCommand, -2, types.FlagWrite, 0, 0, types.CSET},
	"scard":     {"scard", scardCommand, 1, types.FlagRead, 0, 0, types.CSET},
	"sismember": {"sismember", sismemberCommand, 2, types.FlagRead, 0, 0, types.CSET},
	"smembers":  {"smembers", smembersCommand, 1, types.FlagRead, 0, 0, types.CSET},
	"srem":      {"srem", sremCommand, -2, types.FlagWrite, 0, 0, types.CSET},
	"sdiff":     {"sdiff", sdiffCommand, -2, types.FlagRead, 0, 1, types.CSET},
	"sunion":    {"sunion", sunionCommand, -2, types.FlagRead, 0, 1, types.CSET},
	"sinter":    {"sinter", sinterCommand, -2, types.FlagRead, 0, 1, types.CSET},

	"lpush":  {"lpush", lpushCommand, -2, types.FlagWrite, 0, 0, types.CLIST},
	"lpop":   {"lpop", lpopCommand, 1, types.FlagWrite, 0, 0, types.CLIST},
	"rpush":  {"rpush", rpushCommand, -2, types.FlagWrite, 0, 0, types.CLIST},
	"rpop":   {"rpop", rpopCommand, 1, types.FlagWrite, 0, 0, types.CLIST},
	"llen":   {"llen", llenCommand, 1, types.FlagRead, 0, 0, types.CLIST},
	"lindex": {"lindex", lindexCommand, 2, types.FlagRead, 0, 0, types.CLIST},
	"lrange": {"lrange", lrangeComamnd, 3, types.FlagRead, 0, 0, types.CLIST},
	"lset":   {"lset", lsetCommand, 3, types.FlagWrite, 0, 0, types.CLIST},

	"zadd":      {"zadd", zaddCommand, -3, types.FlagWrite, 0, 0, types.CZSET},
	"zcard":     {"zcard", zcardCommand, 1, types.FlagRead, 0, 0, types.CZSET},
	"zrange":    {"zrange", zrangeCommand, -3, types.FlagRead, 0, 0, types.CZSET},
	"zrevrange": {"zrevrange", zrevrangeCommand, -3, types.FlagRead, 0, 0, types.CZSET},
	"zscore":    {"zscore", zscoreCommand, 2, types.FlagRead, 0, 0, types.CZSET},
	"zrem":      {"zrem", zremCommand, -2, types.FlagWrite, 0, 0, types.CZSET},
	"zrank":     {"zrank", zrankCommand, 2, types.FlagRead, 0, 0, types.CZSET},
	"zrevrank":  {"zrevrank", zrevrankCommand, 2, types.FlagRead, 0, 0, types.CZSET},
	"zincrby":   {"zincrby", zincrbyCommand, 3, types.FlagWrite, 0, 0, types.CZSET},

	// command keys
	"del":       {"del", delCommand, -1, types.FlagWrite, 0, 1, types.CKEY},
	"exists":    {"exists", existsCommand, -1, types.FlagRead, 0, 1, types.CKEY},
	"type":      {"type", typeCommand, 1, types.FlagRead, 0, 0, types.CKEY},
	"expire":    {"expire", expireCommand, 2, types.FlagWrite, 0, 0, types.CKEY},
	"expireat":  {"expireat", expireatCommand, 2, types.FlagWrite, 0, 0, types.CKEY},
	"pexpire":   {"pexpire", pexpireCommand, 2, types.FlagWrite, 0, 0, types.CKEY},
	"pexpireat": {"pexpireat", pexpireatCommand, 2, types.FlagWrite, 0, 0, types.CKEY},
	"ttl":       {"ttl", ttlCommand, 1, types.FlagRead, 0, 0, types.CKEY},
	"pttl":      {"pttl", pttlCommand, 1, types.FlagRead, 0, 0, types.CKEY},
	"keys":      {"keys", keysCommand, -1, types.FlagRead, 0, 0, types.CKEY},

	// command connection
	"select": {"select", selectCommand, 1, types.FlagPlain, 0, 0, types.CCONN},
	"ping":   {"ping", pingCommand, 0, types.FlagPlain, 0, 0, types.CCONN},
	"echo":   {"echo", echoCommand, 1, types.FlagPlain, 0, 0, types.CCONN},
	"auth":   {"auth", authCommand, 1, types.FlagPlain, 0, 0, types.CCONN},
	"quit":   {"quit", quitCommand, 0, types.FlagPlain, 0, 0, types.CCONN},
	// redis-cli 连接后发送的验证命令
	"command": {"command", pingCommand, 0, types.FlagPlain, 0, 0, types.CCONN},

	// command server
	"flushdb":  {"flushdb", flushdbCommand, 0, types.FlagWrite, 0, 0, types.CSERVER},
	"flushall": {"flushall", flushallCommand, 0, types.FlagWrite, 0, 0, types.CSERVER},
	"client":   {"client", clientCommand, -1, types.FlagPlain, 0, 0, types.CSERVER},
	"info":     {"info", infoCommad, 0, types.FlagPlain, 0, 0, types.CSERVER},
	"dbsize":   {"dbsize", dbsizeCommad, 0, types.FlagRead, 0, 0, types.CSERVER},

	// command user defined command
	"register": {"reginster", registerCommand, 3, types.FlagWrite, 0, 0, types.CUDC},
	"flush":    {"flush", flushCommand, -1, types.FlagWrite, 0, 0, types.CUDC},
	"scan":     {"scan", scanCommand, -1, types.FlagRead, 0, 0, types.CUDC},
	"count":    {"count", countCommand, 1, types.FlagRead, 0, 0, types.CUDC},
}

func lookupCommand(cmd string) (*command, bool) {
	if c, ok := commandTable[cmd]; ok {
		return &c, true
	}
	return nil, false
}
