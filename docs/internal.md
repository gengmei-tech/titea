# 数据存储结构

- 假定命名空间为：```kv.default.default```
- 每个key都有一个相对应的元信息meta，meta结构如下：

```
type Meta struct {
	Namespace	[]byte		`json:"namespace"`        // 命名空间
	Level		uint8		`json:"levle"`            // 几级命令空间,目前都是3
	Type        string		`json:"type"`          // 对应redis数据类型
	Key         []byte		`json:"key"`           // key
	Count      	int		    `json:"count"`         // 元素数量
	ResetCount	uint		`json:"resetCount"`     // 相同key类型变更后记录重置次数
	CreateAt	int64		`json:"createAt"`          // 创建时间
	ExpireAt  	int64		`json:"expireAt"`       // 过期时间
	LastVAt   	int64		`json:"lastVAt"`        // 最近访问时间
	Extra       []byte		`json:"extra"`          // list类型时有数据其余没有
}

```


## 每种类型在tikv底层的存储结构



### 1. string

执行 ``` set key1 val1``` 操作，在tikv底层会存储2个key，分别为：

* meta

```
key:    kv.default.default.m.key1
value:  json({
  			Namespace:  "kv.default.default",
			Level:		  3,
			Type:		  "string",
			Key:		  "key1",
			Count:		  1,
			ResetCount: 0,
			CreateAt:   1538901288,
			ExpireAt:   0,
			LastVAt:    0,
			Extra:      nil
		})
```

* value

```
key:	kv.default.default.r.key1
value:  val1


```


### 2. hash

执行 ```hmset h1 f1 v1 f2 v2``` 操作，在tikv底层会存储3个key，分别为：

* meta

```
key:	kv.default.default.m.h1
value:  json({
  			Namespace:  "kv.default.default",
			Level:		  3,
			Type:		  "hash",
			Key:		  "h1",
			Count:		  2,
			ResetCount: 0,
			CreateAt:   1538901288,
			ExpireAt:   0,
			LastVAt:    0,
			Extra:      nil
		})
```
* f1

```
key:	kv.default.default.h.h1.f1
value:  v1
```

* f2

```
key:	kv.default.default.h.h1.f2
value:  v2

```

### 3. list

执行 ```lpush l1 v1 v2``` 操作，在tikv底层会存储3个key，分别为：

* meta


```
key:	kv.default.default.m.l1
value:  json({
  			Namespace:  "kv.default.default",
			Level:		  3,
			Type:		  "list",
			Key:		  "l1",
			Count:		  2,
			ResetCount: 0,
			CreateAt:   1538901288,
			ExpireAt:   0,
			LastVAt:    0,
			Extra:      []byte(headIndex8位 + tailIndex8位)
		})
```
* v1

```
key:	kv.default.default.l.l1.v1-index
value:  v1
```

* v2

```
key:	kv.default.default.l.l1.v2-index
value:  v2
```

### 4. set

执行 ``` sadd s1 m1 m2``` 操作，在tikv底层存储3个key， 分别为：

* meta

```
key:	kv.default.default.m.s1
value:  json({
  			Namespace:  "kv.default.default",
			Level:		  3,
			Type:		  "set",
			Key:		  "l1",
			Count:		  2,
			ResetCount: 0,
			CreateAt:   1538901288,
			ExpireAt:   0,
			LastVAt:    0,
			Extra:      nill,
		})
```

* m1

```
key:	kv.default.default.s.s1.m1
value:  []byte{0}
```

* m2

```
key:	kv.default.default.s.s1.m2
value:  []byte{0}
```

### 5. sorted set

执行 ```zadd z1 score1 m1 score2 m2``` 操作，在tikv底层存储5个key， 分别为：

* meta

```
key:	kv.default.default.m.z1
value:  json({
  			Namespace:  "kv.default.default",
			Level:		  3,
			Type:		  "zset",
			Key:		  "z1",
			Count:		  2,
			ResetCount: 0,
			CreateAt:   1538901288,
			ExpireAt:   0,
			LastVAt:    0,
			Extra:      nill,
		})
```

* m1

```
key:	kv.default.default.z.z1.m1
value: score1
```

* score1

```
key:	kv.default.default.z.z1.score1.m1
value: []byte{0}
```

* m2

```
key:	kv.default.default.z.z1.m2
value: score2
```

* score2

```
key:	kv.default.default.z.z1.score2.m2
value: []byte{0}
```

#### score说明
总体上来讲，score会被转成uint64的字节序列，因此，对于负数、小数的score都需要做转换。