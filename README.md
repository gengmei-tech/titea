# TiTea


[![Build Status](https://travis-ci.com/gengmei-tech/titea.svg?branch=master)](https://travis-ci.com/gengmei-tech/titea)
[![Go Report Card](https://goreportcard.com/badge/github.com/gengmei-tech/titea)](https://goreportcard.com/report/github.com/gengmei-tech/titea)
![Project Version](https://img.shields.io/badge/version-1.0.0-brightgreen.svg)



# What is TiTea?
TiTea is a distributed redis protocol compatible NoSQL Database, providing a Redis protocol,  base on Tikv and PD, written in Go.


# Features
- Redis protocol compatible
- Linear scale-out ability
- Multi-tenancy support
- High availability

Thanks [TiKV](https://github.com/tikv/tikv) for supporting the core features


# Quick Run

```
# download docker-compose.yml
curl -s -O https://raw.githubusercontent.com/gengmei-tech/titea/master/docker-compose.yml
# Up
docker-compose up

# Then use redis-cli to connect
redis-cli -p 5379

# Just Like Use Redis
```

# Supported Commands

[supported commands](./docs/commands.md)

# License
TiTea is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.


# Thanks
- [TiDB](https://github.com/pingcap/tidb) 
- [Titan](https://github.com/meitu/titan)
- [Tidis](https://github.com/yongman/tidis)


