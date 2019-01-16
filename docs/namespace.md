
# 命名空间 - Namespace

## 命名空间是什么

- TiTea的多租户是通过命名空间来实现
- 逻辑上, 命名空间用来隔离不同的用户, 不同的项目可以创建各自独立的命名空间, 没有数量限制
- 物理上, 命名空间通过前缀的方式隔离不同的数据, 不同命名空间下的数据, 前缀不同
- 使用上, 类似于redis的DB, 不同的命名空间对应不同的DB; redis默认DB是0, TiTea默认namespace是default.default


## 创建namespace

默认命名空间是default.default, 对应的redis DB是0

命名空间的创建是通过自定义的命令register来实现的, register命令如下：

```
register group.service dbindex creator

参数说明:
- group.service: 命名空间名称, 必须填写
- dbindex: 对应的redis DB, 必须填写, 大于0, 
- creator: 创建人, 必须添加, 作为记录 

register命令最直白的意思是将redis的整型DB映射为字符串的namespace

```

### 为什么命名空间是形式是group.service

- 二级的命名空间, 相同的group下不能有相同的service
- 可以分别按group、server、group.service做不同的流控处理(待完成)
- 可以分别按group、server、group.service做Prometheus metrics 不同label的统计


### 为什么需要namespace到DB的映射

- 直接使用redis的DB不方便统计及显示,使用group.service在统计上能方便的看出来是哪个group,哪个service
- 需要namespace来隔离不同的用户的数据, 又需要客户端像操作redis一样不增加额外的负担
- 通过映射的方式, 使用前由管理员执行register, 客户端不需要更改代码, 像操作redis一样, 通过select db来选择数据库

