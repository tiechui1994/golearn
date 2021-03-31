## MySQL 常用的技巧

1. `GROUP BY` 后进行条件选择, 例如: 选择各科成绩最高的学生

```sql
SELECT a.score, a.name, a.courseid
FROM (
  SELECT max(score) as score, courseid
  FROM tscore 
  GROUP BY courseid
) temp INNER JOIN tscore ON temp.score=tscore.score AND temp.courseid=tscore.courseid
```

> 注意: 上面的语句的 sql_mode 的值当中不能包含 `ONLY_FULL_GROUP_BY`

2. MySQL使用 `on duplicate key update` 引起主键不连续自增
 
为了效率用到 `ON DUPLICATE KEY UPDATE` 进行自动判断是更新还是插入(MySQL判断记录是否存在的依据是主键或者唯一索引,
insert 在主键或者唯一索引已经存在的情况下会插入失败). 这个会导致表的主键id(连续自增), 不是连续自增, 而是跳跃的增加.

解决这个问题的方式:

- 拆分成两个操作, 先更新, 更新无效再插入

1) 先根据唯一索引更新表

2) 在根据上一步的返回值, 返回值大于0, 则说明更新成功不在需要插入, 等于0则需要进行插入该条数据.

- 修改 `innodb_atomic_lock_mode` 参数

`innodb_atomic_lock_mode` 有三种模式: 0, 1, 2, 数据库默认是 1, 就会发生上面的状况, 每次使用 `INSERT INTO ...
ON DUPLICATE KEY UPDATE` 的时候就会id自增, 无论发生了 insert 还是 update. 

1) traditonal级别 (`innodb_atomic_lock_mode=0`), 该自增锁是表锁级别的, 且必须等待当前 SQL 执行完成之后或者回滚掉
才释放. 高并发状况下自增锁开销比较大. 

2) consecutive级别 (`innodb_atomic_lock_mode=1`), 如果是单一的 insert SQL, 可以立即获取该锁, 并立即释放, 而
不必等待当前 SQL 执行完成(除非在其他事务中已经有session获取了自增锁). 另外当SQL是批量 insert 时, 例如 `INSERT INTO 
... SELECT ...`, `LOAD DATA`, `REPLACE ... SELECT ...` 时, 这时还是表级锁, 可以理解成退化为必须等待当前SQL执
行完成才释放. 唯一的缺陷是产生的自增值不一定是完全连续的. 可以认为, 该值为1时相对比较轻量级的锁, 也不会对复制产生影响.

3) interleaved级别 (`innodb_atomic_lock_mode=2`), 所有的 insert 类的SQL都可以立马获取锁并释放, 效率最高. 但是
引入一个新问题: 当 binlog_format 是 statement 时, 这时的复制没法保证安全, 因为批量的 insert, 例如 `INSERT INTO 
... SELECT ...` 语句在这个状况下, 也可以立马获取到一大批的自增id值, 不必锁整个表, slave 在回放时必然产生错乱.
