### 1. 概要
命名空间由三部分组成：```appname```, ```group```, ```service```

* ```appname```: 系统分配，一级，在配置文件里指定，当多个应用共用一个tikv底层集群时，用来区分多个应用。如当tidb、kv共用一个底层的tikv集群时，kv系统使用```appname```前缀来标志数据是kv系统的。

* ```group```: 用户申请，二级，当多个团队共用kv系统时，用来隔离多个团队的数据。

* ```service```: 用户申请，三级，与```group```一同注册分配，用来隔离同个团队多个服务的数据。


### 2. 注册命名空间(```group.service```)

在正式使用kv系统前，需要先注册命名空间

```appname``` 由系统定义，不需要注册


```group.service``` 通过自定义的 ```register``` 命令来注册的，在使用系统前需要先注册命名空间，命令如下：

```register group.service creater [dbindex] ```

参数：

*  ```group.service``` : 必填参数，相同group仅注册一次, 相同group下不能注册相同的service，简言之，group.service不能重复

*  ```creater``` : 必填参数，申请注册人，备注

*  ```dbindex``` : 可选参数，对应redis里 ```select``` 操作的 ```dbindex```，如果提供了 ```dbindex``` 参数，则将 ```dbindex``` 与 ```group.service``` 做映射; 一个 ```dbindex``` 只能映射到一个 ```group.service``` ,  ```dbindex``` 从1开始。

#### 特殊说明
		
对于强类型语言的redis客户端，如 ```go、scala```，在执行register命令注册命名空间的时候，需要提供 ```dbindex``` 参数，然后在客户端执行 ```select dbindex``` 的时候，选择对应已注册的 ```dbindex``` ；对于若类型语言的redis客户端，如 ```python```，在执行 ```register``` 命令注册命名空间的时候，不需要提供 ```dbindex``` ，在客户端执行 ```select``` 操作的时候，直接执行 ```select group.service``` ,  ```group.service``` 为已注册的。


### 3. 本质

+ 存储： 从存储的角度来讲，在tikv底层存储的时候，命名空间作为一个前缀( ```appname.group.service......``` )加在每个key/value前面。
+ 统计： 从统计的角度来讲，命名空间的 ```group、service``` 作为统计的2个纬度，统计其下元素数量
+ 管理： 从管理控制角度来讲，命名空间可以作为流控的手段，控制某个 ```group``` 或 ```service``` 下数据数量(todo)

### 4. 默认及系统命名空间

+ 默认：默认的命名空间是```default.default```, 即 ```group``` 和 ```service``` 都是 ```default``` 。在没有 ```register``` 命令注册命名空间到 ```dbindex``` 的映射前，不要执行 ```select dbindex``` 操作。
+ 系统：系统命名空间的```group```为```sys```, ```service```包括```expire```和```stats```
