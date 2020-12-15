# 关于 mysql 的问题的一些思考?

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
