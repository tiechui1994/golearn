# 关于 mysql join 的问题的一些思考?

## join 原理

mysql 只支持一种 join 算法: Nested-Loop Join (嵌套循环连接), 但 Nested-Loop Join 有三种变种:


1. **Simple Nested-Loop Join**

如下图, r 为驱动表, s 为匹配表, 可以看到从 r 中分别取出 r1, r2, ... rN 区匹配s表的左右列, 然后再合并数据. 对 s 表
进行 rN 次访问, 对数据库开销大.

![image](/images/sql_join_simple.png)


2. **Index Nested-Loop Join**

这个要求非驱动表(匹配表s)上有索引, 可以通过索引来减少比较, 加速查询.

在查询时, 驱动表(r)会根据关键字段的索引进行查找, 当在索引上找到符合的值, 再回表进行查询, 也就是只有当匹配到索引以后才会
进行回表查询.

如果非驱动表(s)的关键字段是主键的话, 性能会非常高, 如果不是主键, 要进行多次(至少是1次)回表查询, 先关联索引, 然后根据索
引的主键id进行回表查询, 性能比索引是主键要慢.

![image](/images/sql_join_index.png)


3. **Block Nested-Loop Join**

如果 join 列没有索引, 就会采用 `Block Nested-Loop Join`.

可以看到中间有个 `join buffer` 缓冲区, 是将驱动表的所有 join 相关的列都先缓存到 `join buffer` 中, 然后批量与匹配
表进行匹配, 将第一种多次比较合并为一次, 降低了非驱动表(s)的访问频率.

默认情况下 join_buffer_size=256K, 在查找的时候 MySQL 会将所有需要的列缓存到 `join buffer` 当中, 包括 select 列,
而不仅仅只缓存关键列. 在一个有N个 JOIN 关联的 SQL 当中会在执行时候分配 N-1 个 `join buffer`.

![image](/images/sql_join_buffer.png)


## join 日常问题思考

现在有四张表:

```
techer: 
    techerid, pk
    username

student:
    studentid, pk
    username

class:
    classid, pk
    name

score:
    techerid
    studentid
    classid
    score
```

上面表的内容很简单的, 这里不做过多的介绍.


1. LEFT JOIN 当中,

1) ON 当中的条件, 会对笛卡尔集合产生什么样子的结果?

2) 在 ON 加过滤条件, 且在 WHERE 也加过滤条件, 两者过滤条件有啥区别, 会怎么样?

3) 只在 ON 加过滤条件, 会怎么样?

4) 只在 WHERE 加过滤条件, 会怎么样? 

```
teacher LEFT JOIN score

// 存在 classid ?
// 存在 studentid ?
```

结果怎样? 性能如何?

先回答第一个问题, `JOIN` 肯定会产生笛卡尔集合. 这个笛卡尔集合
