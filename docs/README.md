## 简介

+ gm-kv 是 redis 协议的实现
+ 除 redis 协议实现外，增加了一部分自定义功能，如命名空间管理、metrics上报等
+ 与 redis 不同的是，数据存储在 tikv，最终是硬盘形式，而非内存，不受内存容量限制
+ 借助tikv自身的分布式功能(PD控制), 实现分布式存储
+ 无状态，可多台配置，一台宕机不会影响整体服务, 客户端需要支持retry，如其中一台挂了，则connect到另一台


## 模块

+ [快速入门](./quickstart.md)
+ [系统结构](./system-structure.md)
+ [命名空间管理](./namespace.md)
+ [协议实现](./internal.md)


## 参考链接

+ [redis](http://www.redis.cn/)  http://www.redis.cn/
+ [Tikv/PD](https://pingcap.com/index.html)  https://pingcap.com/index.html

##	重要说明

+ 用户自定义的key不要包含 ".", 一般redis习惯是用 ":" 来分割
+ dbindex map: 2 => strategy.doris 已分配给策略组