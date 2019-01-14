## 1. keys 命令特殊说明
keys 命令不同于redis的keys命令, 具体如下：
### 不支持redis keys命令的正则表达式匹配，仅支持3类操作：
* keys * :  获取命名空间下全部key
* keys prefix : 获取命名空间下prefix开头的全部key
* keys prefix* : 获取命名空间下prefix开头的全部key

### keys 命令支持2个可选参数, start(起始位置, 默认0) limit(返回数量, 默认5000, 最大5000)
* keys * (start） （limit)
* keys prefix （start） （limit）
* keys prefix* （start）（limit） 
