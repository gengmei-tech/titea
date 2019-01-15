
# Supported Commands 

### String

- [x] get
- [x] set
- [x] setex
- [x] mget
- [x] mset
- [x] incr
- [x] incrby
- [x] decr
- [x] decrby
- [x] strlen
- [x] setnx
- [x] getset

### Hash

- [x] hget
- [x] hstrlen
- [x] hexists
- [x] hlen
- [x] hmget
- [x] hset
- [x] hmset
- [x] hkeys
- [x] hvals
- [x] hgetall
- [x] hdel
- [x] hincrby


### List

- [x] lpush
- [x] lpop
- [x] rpush
- [x] rpop
- [x] llen
- [x] lindex
- [x] lrange
- [x] lset

### Set

- [x] sadd
- [x] scard
- [x] sismember
- [x] smembers
- [x] srem
- [x] sdiff
- [x] sunion
- [x] sinter

### ZSet

- [x] zadd
- [x] zcard
- [x] zrange
- [x] zrevrange
- [x] zscore
- [x] zrem
- [x] zrank
- [x] zrevrank
- [x] zincrby

###  Keys

- [x] del
- [x] exists
- [x] type
- [x] expire
- [x] expireat
- [x] pexpire
- [x] pexpireat
- [x] ttl
- [x] pttl
- [x] keys

### Connection

- [x] select
- [x] ping
- [x] echo
- [x] auth
- [x] quit

### Server

- [x] flushdb
- [x] flushall
- [x] client list
- [x] info
- [x] dbsize


# Special Case

## keys command

keys command differ from redis keys command: 

#### regular expression as redis don't supported, all as follow: 
* keys * :  get all keys under namespace
* keys prefix : get all keys with prefix under namespace
* keys prefix* : just like `keys prefix`, get all keys with prefix under namespace

#### two optional params is supported, start and limit(default 5000, max 5000), 
* keys * (start） （limit)
* keys prefix （start） （limit）
* keys prefix* （start）（limit） 
