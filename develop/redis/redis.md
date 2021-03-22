## redis数据类型对应的底层数据结构

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