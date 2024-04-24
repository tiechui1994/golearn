## Redis 数据类型对应的底层数据结构

### 数据结构与参数配置

hash: ziplist 或 dict

```
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
```

list: ziplist 或 quicklist

```
list-max-ziplist-size -2
list-compress-depth 0
```

set: intset 或 dict

```
set-max-intset-entries 512
```

sortedset: ziplist 或 skiplist

```
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
```

string: int(可以使用整数表示的字符串) 或 embstr(小于等于44字符串) 或 raw(大于44的字符串)

> int 直接使用 robj 存储

### redisObject

从 Redis 内部实现的角度来看, 一个 database 内的这个映射关系是用一个 dict 来维护的. dict 的 key 固定用一种结构来表
达, 这就是动态字符串sds. 而 value 则比较复杂. 使用的是通用数据结构 redisObject.

通用数据结构 redisObject:

```cgo
typedef struct redisObject {
    unsigned  type:4;     // 5 OBJ_STRING, OBJ_LIST, OBJ_SET, OBJ_ZSET, OBJ_HASH
    unsigned  encoding:4; // 10 INT, RAW, EMBSTR, ZIPLIST, QUICKLIST, INTSET, HASHTABLE, SKIPLIST
    unsigned  lru:24;
    int       refcount;
    void* val;
} redisObject;
```

- type, 对象的数据类型, 4 个bit. 取值有5种, 分别对应Redis对外暴露的5种数据结构

- encoding, 对象的内部表示方式, 4 个bit.  可能取值是10种, 不过目前使用的只有8种

- lru, LRU 替换算法使用, 24个bit. 每一个 redisObject 都不一样

- refcount, 引用计数. 它允许 robj 对象在某些状况下被共享

- ptr, 数字指针, 指向真正的数据.


下面是字符串编码过程: 

```cgo
#define sdsEncodedObject(objptr) ( \
    objptr->encoding == REDIS_ENCODING_RAW || objptr->encoding == REDIS_ENCODING_EMBSTR )

// 尝试对字符串对象进行编码, 以节约内存.
robj *tryObjectEncoding(robj *o) {
    long value;

    sds s = o->ptr;
    size_t len;

    // 确保这是一个字符串对象, 这是我们在此函数中编码的唯一类型.
    // 其他类型使用编码的内存有效表示形式, 但由实现该类型的命令来处理.
    redisAssertWithInfo(NULL,o,o->type == REDIS_STRING);

    // 字符串编码既不是 RAW 也不是 EMBSTR 时, 返回本身(言外之意, 只在字符串的编码为 RAW 或者 EMBSTR 时尝试进行编码)
    if (!sdsEncodedObject(o)) return o;

     // 不对共享对象进行编码
     if (o->refcount > 1) return o;

    // 对字符串进行检查
    // 只对长度小于或等于 21 字节, 并且可以被解释为整数的字符串进行编码
    len = sdslen(s);
    if (len <= 21 && string2l(s,len,&value)) {
         // 字符串可以编码为 long, 尝试使用共享库. 
         // 注意: 在使用maxmemory时, 避免使用共享整数, 因为每个对象都需要具有私有LRU字段才能使LRU算法正常工作.
        if (server.maxmemory == 0 &&
            value >= 0 &&
            value < REDIS_SHARED_INTEGERS)
        {
            decrRefCount(o);
            incrRefCount(shared.integers[value]);
            return shared.integers[value];
        } else {
            // 编码为整数, 释放掉之前的指针
            if (o->encoding == REDIS_ENCODING_RAW) sdsfree(o->ptr);
            o->encoding = REDIS_ENCODING_INT;
            o->ptr = (void*) value;
            return o;
        }
    }

    // 尝试将 RAW 编码的字符串编码为 EMBSTR 编码
    // REDIS_ENCODING_EMBSTR_SIZE_LIMIT 是 39
    if (len <= REDIS_ENCODING_EMBSTR_SIZE_LIMIT) {
        robj *emb;

        if (o->encoding == REDIS_ENCODING_EMBSTR) return o;
        emb = createEmbeddedStringObject(s,sdslen(s));
        decrRefCount(o);
        return emb;
    }

    // 这个对象没办法进行编码, 尝试从 SDS 中移除所有空余空间
    if (o->encoding == REDIS_ENCODING_RAW &&
        sdsavail(s) > len/10)
    {
        o->ptr = sdsRemoveFreeSpace(o->ptr);
    }

    /* Return the original object. */
    return o;
}

// 创建一个 REDIS_ENCODING_EMBSTR 编码的字符对象
// 这个字符串对象中的 sds 会和字符串对象的 redisObject 结构一起分配
// 因此这个字符也是不可修改的
robj *createEmbeddedStringObject(char *ptr, size_t len) {
    robj *o = zmalloc(sizeof(robj)+sizeof(struct sdshdr)+len+1);
    struct sdshdr *sh = (void*)(o+1); // 注意: p+1 是内存偏移一个 sizeof(*p), 而不是 1 个字节

    o->type = REDIS_STRING;
    o->encoding = REDIS_ENCODING_EMBSTR;
    o->ptr = sh+1; // 内存偏移一个 sizeof(struct sdshdr)
    o->refcount = 1;
    o->lru = LRU_CLOCK();

    sh->len = len;
    sh->free = 0;
    if (ptr) {
        memcpy(sh->buf,ptr,len);
        sh->buf[len] = '\0';
    } else {
        memset(sh->buf,0,len+1);
    }
    return o;
}
```

### list 当中的 BLPOP, BRPOP, BRPOPLPUSH 的阻塞操作

BLPOP, BRPOP, BRPOPLPUSH 三个命令都可能造成客户端被阻塞. 以下将这些名称统称为列表的阻塞原语.

阻塞原语并不是一定造成客户端阻塞:

- 只有当命令被用于空列表时, 它们才会阻塞客户端

- 如果被处理的列表不为空, 它们就执行无阻塞版本的 LPOP, RPOP 或 RPOPLPUSH

阻塞:

阻塞一个客户端需要以下的步骤:

1. 将客户端的状态设置为 "正在阻塞", 并记录阻塞这个客户端的各个键, 以及阻塞的最长时间等数据.

2. 将客户端的信息记录到 `server.db[i]->blocking_keys` 中 (i为客户端所使用的数据库的编号)

3. 继续维持客户端和服务器之间的网络连接, 但不再向客户端传送任何信息, 造成客户端阻塞.

步骤2是解除阻塞的关键, `server.db[i]->blocking_keys` 是一个字典, 字典的键是造成客户端阻塞的键, 字典的值是一个链表,
链表里保存了所有因为这个键而被阻塞的客户端.

当客户端被阻塞之后, 脱离阻塞状态的三个方法:

1. 被动脱离: 有其他客户端为阻塞的键推入新数据

2. 主动脱离: 到达执行阻塞原语设定的最大阻塞时间

3. 强制脱离: 客户端与服务器连接断开.

> 阻塞因 LPUSH, RPUSH, LINSERT 等添加命令而被取消

通过将新元素推入造成客户端阻塞的某个键, 可以让相应的客户端从阻塞状态中脱离出来(取消阻塞客户端的数量取决于推入元素的数据).

LPUSH, RPUSH, LINSERT 添加新元素到列表的命令, 底层是由一个 `pushGenericCommand` 的函数实现.

当向一个空键推入新元素时, 函数执行的两件事:

1. 检查这个键是否存在于 `server.db[i]->blocking_keys` 当中, 如果存在, 那么说明至少又一个客户端因为这个 key 而被
阻塞, 程序会为这个 key 创建一个 readList (`redis.h`) 结构, 并将它添加到 `server.ready_keys` 链表中.

2. 将给定的值添加到列表键中.

readList 记录了 key 和 db.


Redis 主进程在执行完 `pushGenericCommand` 函数之后, 会继续调用 `handleClientBlockedOnLists` 函数, 这个函数执
行的操作:

1. 如果 `server.ready_keys` 不为空, 弹出该链表的头元素, 并取出元素中的 readyList 值.

2. 根据 readyList 当中保存的 key 和 db, 在 `server.db[i]->blocking_keys` 所有因 key 而被阻塞的客户端

3. 如果 key 不为空, 那么从 key 中弹出一个元素, 并弹出客户端客户端链表的第一个客户端, 然后将被弹出的元素返回给被弹出客
户端作为阻塞原语的返回值.

4. 根据 readyList 结构的属性, 删除 `server.db[i]->blocking_keys` 中相应的客户端, 取消客户端的阻塞状态

5. 继续执行上述 3 和 4, 直到没有 key 可弹出, 或者所有因为 key 而被阻塞的客户端都取消阻塞为止.

6. 继续执行1 
 
 

### 事务

Redis 通过 MULTI, DISCARD, EXEC, WATCH 四个命令实现事务功能.

事务提供了一种 "将多个命令打包, 然后一次性, 按顺序执行" 的机制, 并且事务在执行的期间不会主动中断 -- 服务器在执行完事务
中所有命令之后, 才会继续处理其他客户端的其他命令.

MULTI 命令的执行标记着事务的开始. 这个命令唯一做的就是, 将客户端的 REDIS_MULTI 选项打开, 让客户端从非事务状态切换到
事务状态.

当客户端进入事务状态之后, 服务器在收到来自客户端的命令时, 不会立即执行命令, 而是将这些命令全部放进一个事务队列里, 然后返
回 QUEUED, 表示命令已入队.

> 注: 客户端进入事务状态之后,客户端发送的命令就会被放入到事务队列里, 但不是所有命令都会被放进事务队列, 其中的例外就是 
EXEC, DISCARD, MULTI, WATCH 这四个命令. 这四个命令发送到服务器时, 会立即被服务器执行.

当客户端处于事务状态,那么当 EXEC 命令执行时,服务器根据客户端所保存的事务队列, 以FIFO的方式执行事务队列中的命令.

当事务队列里所有命令被执行完成之后, EXEC 命令会将回复队列作为自己的执行结果返回给客户端, 客户端从事务状态返回到非事务状
态, 至此, 事务执行完毕.

DISCARD 命令用于取消一个事务,它清空客户端的整个事务队列, 然后将客户端从事务状态切换回非事务状态, 最后返回字符 OK 给客户
端, 说明事务已被取消.

Redis 的事务是不可嵌套的, 当客户端处于事务状态, 而客户端又再向服务器发送 MULTI 时, 服务器只是简单地向客户端发送一个错误,
然后继续等待其他命令的入队. `MULTI 命令的发送不会造成整个事务失败, 也不会修改队列中已有的数据.`

WATCH 只能在客户端进入事务状态之前执行, `在事务状态下发送 WATCH 命令会引发一个错误, 但它不会造成事务失败, 也不会修改队
列中已有的数据`.

WATCH 用于在事务之前监听任意数量的key, 当调用 EXEC 命令执行事务时, 如果任意一个被监视的 key 已被其他客户端修改了, 那
么整个事务不在执行, 直接返回失败.

#### WATCH 命令的实现

每个代表数据库的 redisDb (`redis.h`) 结构类型中, 都保存了一个 watched_keys 字典, 字典的 key 是这个数据库被监视的
key, 而字典的 val 是一个链表, 链表保存了所有监视这个键的客户端.

WATCH 命令的作用, 就是将当前客户端和要监视的key在 watched_keys 中进行关联.


在任何对何数据库空间 (key space) 进行修改的命令成功执行之后 (例如: FLUSHDB, SET, DEL, LPUSH, SADD, ZREM 等),
touchWatchedKey(`multi.c`) 函数都会被调用 -- 检查数据库的 watched_keys 字典, 看是否有客户端在监视已经被命令修改
的 key, 如果有, 程序将所有监视这个/这些被修改 key 的客户端的 REDIS_DIRTY_CAS 选项打开:

当客户端发送 EXEC 命令, 触发事务执行时, 服务器会对客户端的状态进行检查:

- 如果客户端的 REDIS_DIRTY_CAS 选项已经被打开, 那么说明被客户端监视的 key 至少又一个已经被修改了, 事务的安全性已经被
破坏. 服务器会放弃执行这个事务, 直接向客户端返回空, 表示事务执行失败.

- 如果 REDIS_DIRTY_CAS选项没有被打开, 那么说明索引监听的 key 都安全, 服务器正式执行事务.


#### ACID

Redis 事务保证了其中的一致性(C) 和隔离性(I), 但并不保证原子性(A) 和持久性(D).

- 原子性 (Atomicity):

单个 Redis 命令的执行是原子性, 但 Redis 没有在事务上增加任何维持原子性的机制, 所以 Redis 事务的执行并不是原子性的.

如果一个事务列表中的所有命令都被执行成功, 那么称这个事务执行成功.

另一方面, 如果Redis服务器进程在执行事务的过程中被停止 -- 比如接到KILL信号,崩溃等等, 那么事务执行失败. 

当事务失败时,Redis也不会进行任何的重试或回滚操作.          

- 一致性(Consistency):

Redis 一致性问题可以分为三个部分: 入队错误, 执行错误, Redis进程被终结

> 入队错误

在命令入队过程中, 如果客户端向服务器发送错误的命令(比如命令参数不对等), 服务器会向客户端返回一个出错信息,并且将客户端的事
务状态设置为 REDIS_DIRTY_EXEC.

当客户端执行 EXEC 命令时,Redis 会拒绝执行状态为 REDIS_DIRTY_EXEC 的事务, 并返回错误.

> 执行错误

如果命令在事务执行过程中发生错误(比如, 对一个不同类型的 key 执行了错误的操作), 那么 Redis 只会将错误包含在事务的结果中,
这不会引起事务中断或整个失败, 不会影响已经执行事务命令的结果, 也不会影响后面要执行事务命令, 对事务一致性没有要求没影响.

> Redis进程被 KILL

在执行事务过程中, Redis进程被KILL, 那么根据持久化模式, 可能的情况:

1) 内存模式: Redis没有采取任何持久化机制,那么重启之后的数据库总是空白的, 所以数据总数一致的.

2) RDB模式: 在执行事务时, Redis不会中断事务去执行保存RDB的工作, 只有在事务执行之后, 保存RDB的工作才有可能开始. 所以
以RDB模式下的Redis服务器进程在事务中被KILL时, 事务内执行的命令, 不管成功了多少, 都不会保存到 RDB 文件中. 恢复数据库
需要的RDB文件, 而这个RDB文件保存的是最近一次数据快照(snapshot), 所以数据可能不是最新的, 但只要RDB文件本身没有其他问题
而出错, 那么还原后的数据库是一致的.

3) AOF模式: 因为保存 AOF 文件的工作是在后台线程进行, 所以即使是在事务执行的中途, 保存AOF文件的工作也在继续执行, 因此
根据事务语句是否被写入并保存到AOF文件, 有两种情况:

3.1) 如果事务语句未写入到AOF文件, 或者AOF未被SYNC保存到磁盘, 当进程被KILL之后, Redis可以根据最近一次成功保存到磁盘
的AOF文件来还原数据库, 只要AOF文件本身没有其他问题, 那么还原后的数据库总数一致的, 但其中的数据不一定是最新的.

3.2) 如果事务的部分语句写入到AOF文件, 并且 AOF 文件被成功保存, 那么不完整的事务执行信息就会遗留在 AOF 文件, 当重启
Redis时, 程序检测AOF文件并不完整, Redis会退出, 并报告错误. 需要使用 redis-check-aof 工具将部分成功的事务命令移除之
后, 才能再次启动Redis. 还原之后的数据总是一致的, 而且数据也是最新的.

> 隔离性(Isolation)

Redis 是单进程程序, 并且保证在执行事务时, 不会对事务进行中断, 事务可以运行直到执行完所有事务队列中的命令为止. 因此, 
Redis 的事务总数带有隔离性的.

> 持久性(Durability)

因为事务不过是用队列包裹的一组 Redis 命令, 并没有任何额外的持久化功能, 所以事务的持久性由 Redis 所使用的持久化模式决定
的:

1. 单纯的内存模式, 事务是不持久化的

2. 在RDB模式下, 服务器可能在事务执行之后, RDB文件更新之前这段时间失败,所以RDB 模式下 Redis 事务也不能持久化.

3. 在AOF的 "总是SYNC" 模式下, 事务的每条命令在执行成功之后, 都会立即调用 fsync 或 ddatasync 将事务数据写入到 AOF
文件. 但是, 这种保存是由后台线程进行的, 主线程不会阻塞直到保存成功, 所有命令执行成功到数据保存到硬盘之间, 还是有一段非常
小的间隔, 所以这种模式下的事务也是不持久的. AOF的其他模式也是一样的.

