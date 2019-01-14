### 1. 下载安装

```
cd $GOPATH/src
mkdir gengmei
cd gengmei
git clone git@git.wanmeizhensuo.com:system/gm-kv.git kv
cd kv
// 编译
make build
// 运行
bin/gm-kv --config=./config/gmkv.toml
```
### 2. 运行使用

```
redis-cli -p 5379
127.0.0.1:5379> get a
"wang"
127.0.0.1:5379> set b zhang
OK
```

### 3. 配置文件

##### 配置文件默认在根目录下config文件夹


```
# Configuration file for gm-kv

# Multi app(tidb and others) share common tikv use appname to distinguish
appname = "kv"

# KV server host
host = "0.0.0.0"

# KV server port
port = 5379

# max connection(disable current)
max_connection = 5000

auth = ""

[backend]
# Placement driver addresses, tikv://etcd-node1:port,etcd-node2:port?cluster=1&disableGC=false
pd-address = "192.168.15.11:2379,192.168.15.12:2379"

# Transaction retry count when commit failed in case of conflict
txn_retry = 2


[log]
# Log level: debug, info, warn, error, fatal
level = "debug"

# Log format, one of json, text
format = "text"

# To save log
# log-path = ""

# Max log file size
max-size = 300

max-days = 3

[ttl]
# Expire check interval in second
check_interval = 30

# Loop count in one check
check_loop  = 10


[metric]
# Push to prometheus
job = "gmkv"

# Prometheus pushgateway address, leaves it empty will disable prometheus push.
address = "192.168.15.11:9091"

# Prometheus client push interval in second, set "0" to disable prometheus push.
interval = "20s"

```

### 4. 支持的命令

#### keys

	+----------+--------------------------------------+
	|  command  |               format                |
	+-----------+-------------------------------------+
	|    del    | del key1 key2 ...                   |
	+-----------+-------------------------------------+
	|   exists  | exists key                          | 
	+-----------+-------------------------------------+
	|    type   | type key                            |
	+-----------+-------------------------------------+
	|   expire  | expire key seconds                  |
	+-----------+-------------------------------------+
	|  expireat | expireat key timestamp              |
	+-----------+-------------------------------------+
	|  pexpire  | pexpire key milliseconds            |
	+-----------+-------------------------------------+
	| pexpireat | pexpireat key milli-timestamp       |
	+-----------+-------------------------------------+
	|    ttl    | ttl key                             |
	+-----------+-------------------------------------+
	|    pttl   | pttl key                            |
	+-----------+-------------------------------------+
	|    keys   | keys key-prefix                     |
	+-----------+-------------------------------------+ 

#### string

    +-----------+-------------------------------------+
    |  command  |               format                |
    +-----------+-------------------------------------+
    |    get    | get key                             |
    +-----------+-------------------------------------+
    |    set    | set key value [EX sec|PX ms][NX|XX] | 
    +-----------+-------------------------------------+
    |    mget   | mget key1 key2 ...                  |
    +-----------+-------------------------------------+
    |    mset   | mset key1 value1 key2 value2 ...    |
    +-----------+-------------------------------------+
    |    incr   | incr key                            |
    +-----------+-------------------------------------+
    |   incrby  | incr key step                       |
    +-----------+-------------------------------------+
    |    decr   | decr key                            |
    +-----------+-------------------------------------+
    |   decrby  | decrby key step                     |
    +-----------+-------------------------------------+
    |   strlen  | strlen key                          |
    +-----------+-------------------------------------+
    |   setex   | setex key seconds value             |
    +-----------+-------------------------------------+
    |   setnx   | setnx key value                     |
    +-----------+-------------------------------------+
    |   getset  | getset key value                    |
    +-----------+-------------------------------------+
    
#### hash

    +------------+------------------------------------------+
    |  Commands  | 				Format                        |
    +------------+------------------------------------------+
    |    hget    | hget key field                           |
    +------------+------------------------------------------+
    |   hstrlen  | hstrlen key                              |
    +------------+------------------------------------------+
    |   hexists  | hexists key                              |
    +------------+------------------------------------------+
    |    hlen    | hlen key                                 |
    +------------+------------------------------------------+
    |    hmget   | hmget key field1 field2 field3...        |
    +------------+------------------------------------------+
    |    hdel    | hdel key field1 field2 field3...         |
    +------------+------------------------------------------+
    |    hset    | hset key field value                     |
    +------------+------------------------------------------+
    |   hsetnx   | hsetnx key field value                   |
    +------------+------------------------------------------+
    |    hmset   | hmset key field1 value1 field2 value2... |
    +------------+------------------------------------------+
    |    hkeys   | hkeys key                                |
    +------------+------------------------------------------+
    |    hvals   | hvals key                                |
    +------------+------------------------------------------+
    |   hgetall  | hgetall key                              |
    +------------+------------------------------------------+
    
#### list

    +------------+-----------------------+
    |  commands  |         format        |
    +------------+-----------------------+
    |    lpush   | lpush key             |
    +------------+-----------------------+
    |    lpop    | lpop key              |
    +------------+-----------------------+
    |    rpush   | rpush key             |
    +------------+-----------------------+
    |    rpop    | rpop key              |
    +------------+-----------------------+
    |    llen    | llen key              |
    +------------+-----------------------+
    |   lindex   | lindex key index      |
    +------------+-----------------------+
    |   lrange   | lrange key start stop |
    +------------+-----------------------+
    |    lset    | lset key index value  |
    +------------+-----------------------+
    |    ltrim   | ltrim key start stop  |
    +------------+-----------------------+
   
#### set

    +-------------+--------------------------------+
    |   commands  |             format             |
    +-------------+--------------------------------+
    |     sadd    | sadd key member1 [member2 ...] |
    +-------------+--------------------------------+
    |    scard    | scard key                      |
    +-------------+--------------------------------+
    |  sismember  | sismember key member           |
    +-------------+--------------------------------+
    |   smembers  | smembers key                   |
    +-------------+--------------------------------+
    |     srem    | srem key member                |
    +-------------+--------------------------------+
    |    sdiff    | sdiff key1 key2                |
    +-------------+--------------------------------+
    |    sunion   | sunion key1 key2               |
    +-------------+--------------------------------+
    |    sinter   | sinter key1 key2               |
    +-------------+--------------------------------+
    |  sdiffstore | sdiffstore key1 key2 key3      |
    +-------------+--------------------------------+
    | sunionstore | sunionstore key1 key2 key3     |
    +-------------+--------------------------------+
    | sinterstore | sinterstore key1 key2 key3     |
    +-------------+--------------------------------+
   
#### sorted set

    +------------------+---------------------------------------------------------------+
    |     commands     |                             format                            |
    +------------------+---------------------------------------------------------------+
    |       zadd       | zadd key member1 score1 [member2 score2 ...]                  |
    +------------------+---------------------------------------------------------------+
    |       zcard      | zcard key                                                     |
    +------------------+---------------------------------------------------------------+
    |      zrange      | zrange key start stop [WITHSCORES]                            |
    +------------------+---------------------------------------------------------------+
    |     zrevrange    | zrevrange key start stop [WITHSCORES]                         |
    +------------------+---------------------------------------------------------------+
    |   zrangebyscore  | zrangebyscore key min max [WITHSCORES][LIMIT offset count]    |
    +------------------+---------------------------------------------------------------+
    | zrevrangebyscore | zrevrangebyscore key max min [WITHSCORES][LIMIT offset count] |
    +------------------+---------------------------------------------------------------+
    | zremrangebyscore | zremrangebyscore key min max                                  |
    +------------------+---------------------------------------------------------------+
    |    zrangebylex   | zrangebylex key min max [LIMIT offset count]                  |
    +------------------+---------------------------------------------------------------+
    |  zrevrangebylex  | zrevrangebylex key max min [LIMIT offset count]               |
    +------------------+---------------------------------------------------------------+
    |  zremrangebylex  | zremrangebylex key min max                                    |
    +------------------+---------------------------------------------------------------+
    |      zcount      | zcount key                                                    |
    +------------------+---------------------------------------------------------------+
    |     zlexcount    | zlexcount key                                                 |
    +------------------+---------------------------------------------------------------+
    |      zscore      | zscore key member                                             |
    +------------------+---------------------------------------------------------------+
    |       zrem       | zrem key member1 [member2 ...]                                |
    +------------------+---------------------------------------------------------------+
   
#### connection

	+-----------+-------------------------------------+
	|  command  |               format                |
	+-----------+-------------------------------------+
	|  select   | select dbindex OR select namespace  |
	+-----------+-------------------------------------+
	|   ping    | ping                                | 
	+-----------+-------------------------------------+
	|   echo    | echo message                        |
	+-----------+-------------------------------------+
	|   auth    | auth password                       |
	+-----------+-------------------------------------+
	|   quit    | quit                                |
	+-----------+-------------------------------------+
    
#### server

	+-----------+-------------------------------------+
	|  command  |               format                |
	+-------------------------------------------------+
	|  flushdb  | flushdb                             |
	+-----------+-------------------------------------+
	| flushall  | flushall                            |
	+-----------+-------------------------------------+
	|  client   | client list                         |
	+-----------+-------------------------------------+
	|  info     | info                                |
	+-----------+-------------------------------------+
    
#### udc(user definded command)

	+-----------+--------------------------------------------------+
	|  command  |               format                             |
	+--------------------------------------------------------------+
	|  register | register group.service  creater  [dbindex]       |
	+-----------+--------------------------------------------------+
	|   flush   | flush prefix(appname.[group.service])            |
	+-----------+--------------------------------------------------+	

### 5. 压测
对于redis自带的benchmark，多个客户端会发送相同的key，对导致底层tikv的锁冲突问题，所以gm-kv提供了一个简单的benchmark:

```
// 编译
make benchmark
// 运行
bin/benchmark -h 127.0.0.1 -p 5379 -c 50 -n 2000 -l false -t set,get

参数: 
-h: gmkv host
-p: gmkv port
-c: cocurrent clients
-n: number of requests for each client
-l: loop forever
-t: test command, for now set/get
```

