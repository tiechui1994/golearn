## ziplist

### ziplist 的数据结构

```
<zlbytes><zltail><zllen><entry1>...<entryN><zlend>
```

zlbytes 4字节, 整个 ziplist 占用的内存字节数. **这个变量在扩容的时候使用**

zltail 4字节, 到达 ziplist 表最后一个entry(的头部)的偏移量. **通过这个偏移量, 可以直接找到表尾节点.**

zllen  2字节, ziplist 当中 entry 的数量. 当这个值小于 2^16-1 时, 这个值就是 ziplist 中节点的数量; 当值是 2^16-1
时, 节点的数量需要遍历才能获得.

entryX 可变长度, ziplist所保存的节点. 

zlend, 1字节, 值是 0xFF, 用于标记 ziplist 的末端.


- entry 的数据结构:

```
<prevlen><len><data>
```

prevlen 表示前一个节点的长度, 通过这个值, 可以进行指针计算, 从而跳转到上一个节点. 从而达到倒序遍历的目的.

prevlen 占用1个字节或5个字节:

1字节: 如果前一个节点长度小于 254 字节, 使用一个字节保存它的值.

5字节: 如果前一个节点长度大于等于 254 字节, 那么第一个字节设置为 254, 后面的4个字节保存实际的长度. 

> 注: 由于 255 是作为 zlend 的特定值, 所以这里使用 254 作为分界线.

len 当中包含了编码类型和数据的长度, 使用变长编码, 分为9种情况:

00xxxxxx, 1个字节, 存储字符数组, 数组的长度小于等于 2^6-1 字节
01xxxxxx xxxxxxxx, 2字节, 存储字符数组, 数组长度小于等于 2^14-1 字节
10000000 xxxxxxxx xxxxxxxx xxxxxxxx xxxxxxxx 5字节, 存储字符数组, 数组长度小于等于 2^32-1 字节

11000000 1 字节, int16类型整数
11010000 1 字节, int32类型整数
11100000 1 字节, int64类型整数
11110000 1 字节, 24 bit有符号整数
11111110 1 字节, 8 bit有符号整数
1111xxxx 1 字节, 4 bit有符号整数 (介于0-12之间)


data: 真实的数据.

例如: `hello` 案例

```
prelen: ?
len: 1字节 (00000101)
data: 5字节, 内容是 hello 
```


### ziplist api

- 将新节点插入到 ziplist 节点 p 的过程:

1) 计算 prelen 和 prelensize (如果 p 指向列表末端 `p[0] == 0xFF`, 则需要根据 `zltail` 获取 ptail 节点, 否则
直接获取 p 节点的 prelen); 尝试将插入节点的值转换为整数, 从而计算出新插入节点的编码方式 encoding 和 len, 最终计算出
新插入节点所需的空间大小 reqlen.

2) 计算 nextdiff, p节点的 prelensize 和 新节点的 prelensize 的差(可能值: 0,4,-4)

> 注: 新插入的节点位置是 p, 原来p节点(包括p)之后的所有节点都要后移一位. 因此 p 节点的 prelen 的值是当前节点的长度. 但
是这个值可能发生内存字节大小的变化.

3) 记录到达 ziplist 到达 p 的偏移量 (因为后续需要重新分配内存)

4) 重新进行内存重新分配(zlbytes+reqlen+nextdiff).

5) 将 p-nextdiff, 移动到 p+reqlen, 重新设置 zltail. 以及进行可能的级联更新.

6) 设置新节点的各项属性 prelen, len, data

7) 更新 ziplist 的相关属性, zlbytes, zllen


> 带参数的宏和函数很相似, 但有本质上的区别: 宏展开仅仅是字符串的替换, 不会对表达式进行计算; 宏在编译之前就被处理掉了, 
它没有机会参与编译, 也不会占用内存. 而函数是一段可以重复使用的代码, 会被编译, 会给它分配内存, 每次调用函数, 就是执行这
块内存中的代码.

```cgo
#define ZIP_END 255
#define ZIP_BIGLEN 254

#define ZIP_DECODE_PREVLENSIZE(ptr, prevlensize) do {           \
    if ((ptr)[0] < ZIP_BIGLEN) {                                \
        (prevlensize) = 1;                                      \
    } else {                                                    \
        (prevlensize) = 5;                                      \
    }                                                           \
} while(0);

#define ZIP_DECODE_PREVLEN(ptr, prevlensize, prevlen) do {      \
                                                                \
    /* 先计算被编码长度值的字节数 */                                 \
    ZIP_DECODE_PREVLENSIZE(ptr, prevlensize);                   \
                                                                \
    /* 再根据编码字节数来取出长度值 */                               \
    if ((prevlensize) == 1) {                                   \
        (prevlen) = (ptr)[0];                                   \
    } else if ((prevlensize) == 5) {                            \
        assert(sizeof((prevlensize)) == 4);                     \
        memcpy(&(prevlen), ((char*)(ptr)) + 1, 4);              \
        memrev32ifbe(&prevlen);                                 \
    }                                                           \
} while(0);

/* Insert item at "p". */
/*
 * 根据指针 p 所指定的位置, 将长度为 slen 的字符串 s 插入到 zl 中.
 *
 * 函数的返回值为完成插入操作之后的 ziplist
 *
 * T = O(N^2)
 */
static unsigned char *__ziplistInsert(unsigned char *zl, unsigned char *p, unsigned char *s, unsigned int slen) {
    // 记录当前 ziplist 的长度
    size_t curlen = intrev32ifbe(ZIPLIST_BYTES(zl)), reqlen, prevlen = 0;
    size_t offset;
    int nextdiff = 0;
    unsigned char encoding = 0;
    long long value = 123456789; /* initialized to avoid warning. Using a value
                                    that is easy to see if for some reason
                                    we use it uninitialized. */
    zlentry entry, tail;

    /* Find out prevlen for the entry that is inserted. */
    // 计算 prelen, 前一个节点的长度.
    if (p[0] != ZIP_END) {
        // 如果 p[0] 不指向列表末端, 说明列表非空, 并且 p 正指向列表的其中一个节点
        // 然后直接取出当前节点的上一个节点的 prevlen 和 prevlensize
        ZIP_DECODE_PREVLEN(p, prevlensize, prevlen);
    } else {
        // 如果 p 指向表尾末端, 那么需要检查列表是否为空:
        // 1) 如果 ptail 指向 ZIP_END, 那么列表为空;
        // 2) 如果 ptail 没有指向 ZIP_END, 那么 ptail 将指向列表的最后一个节点.
        // ptail 是最后一个 entry (的起始位置) 的指针
        unsigned char *ptail = ZIPLIST_ENTRY_TAIL(zl); // z1+zltail
        if (ptail[0] != ZIP_END) {
            // 表尾节点为新节点的前置节点, 计算该节点的长度: prerawlensize+lensize+len
            prevlen = zipRawEntryLength(ptail); 
        }
    }
    
    // 尝试将字符串 s 转换为整数 value, 如果成功:
    // 1) value 将保存转换后的整数值. 
    // 2) encoding 则保存适用于 value 的编码方式
    // 无论使用什么编码, reqlen 都保存节点值的长度.
    // T = O(N)
    if (zipTryEncoding(s,slen,&value,&encoding)) {
        /* 'encoding' is set to the appropriate integer encoding */
        reqlen = zipIntSize(encoding);
    } else {
        reqlen = slen;
    }
    
    // 计算编码 prevlen 所需的大小(1或5).
    // T = O(1)
    reqlen += zipPrevEncodeLength(NULL, prevlen);
    // 计算编码 slen 所需的大小.( 它是由 encoding 和 slen 决定的)
    // T = O(1)
    reqlen += zipEncodeLength(NULL,encoding,slen);

    // 只要新节点不是被添加到列表末端, 检查看 p 所指向的节点能否编码新节点的长度.
    // nextdiff 保存了新旧编码之间的字节大小差, 如果这个值大于 0 那么说明需要对 p 所指向的节点进行扩展.
    // prelen 占用字节长度只能是 1 和 5, 可能的结果是 0, 4
    nextdiff = (p[0] != ZIP_END) ? zipPrevLenByteDiff(p, reqlen) : 0;

    // 由于重分配空间可能会改变 zl 的地址. 所以在分配之前, 需要记录 zl 到 p 的偏移量, 然后在分配之后依靠偏移量
    // 还原 p .
    offset = p-zl;
    
    // curlen 是 ziplist 原来的长度
    // reqlen 是整个新节点的长度
    // nextdiff 是新节点的后继节点扩展 header 的长度 (要么 0 字节, 要么 4 个字节)
    // T = O(N)
    zl = ziplistResize(zl,curlen+reqlen+nextdiff); // 调用 realloc 方法重新分配内存
    p = zl+offset; // 恢复 p 节点
    
    // 计算尾节点的偏移量
    if (p[0] != ZIP_END) {
        // 新元素之后还有节点, 因为新元素的加入, 需要对这些原有节点进行调整: 移动现有元素, 为新元素的插入空间腾出位置
        // 将 p - nextdiff 之后的内容, 移动到 p + reqlen 之后
        // T = O(N)
        memmove(p+reqlen, p-nextdiff, curlen-offset-1+nextdiff);

        /* Encode this entry's raw length in the next entry. */
        // 将新节点的长度编码至后置节点(更新后置节点的prevrawlen值), p+reqlen 定位到后置节点
        // reqlen 是新节点的长度
        // T = O(1)
        zipPrevEncodeLength(p+reqlen, reqlen);

        /* Update offset for tail */
        // 更新到达表尾的偏移量，将新节点的长度也算上
        ZIPLIST_TAIL_OFFSET(zl) =
            intrev32ifbe(intrev32ifbe(ZIPLIST_TAIL_OFFSET(zl))+reqlen);

        // 如果新节点的后面有多于一个节点, 需要将 nextdiff 记录的字节数也计算到表尾偏移量中.
        // 这样才能让表尾偏移量正确对齐表尾节点.
        // T = O(1)
        tail = zipEntry(p+reqlen);
        if (p[reqlen+tail.headersize+tail.len] != ZIP_END) {
            ZIPLIST_TAIL_OFFSET(zl) =
                intrev32ifbe(intrev32ifbe(ZIPLIST_TAIL_OFFSET(zl))+nextdiff);
        }
    } else {
        /* This element will be the new tail. */
        // 新元素是新的表尾节点
        ZIPLIST_TAIL_OFFSET(zl) = intrev32ifbe(p-zl);
    }

    /* When nextdiff != 0, the raw length of the next entry has changed, so
     * we need to cascade the update throughout the ziplist */
    // 当 nextdiff != 0 时, 新节点的后继节点的(prevrawlen)长度已经被改变, 
    // 所以需要级联地更新后续的节点
    if (nextdiff != 0) {
        offset = p-zl;
        zl = __ziplistCascadeUpdate(zl,p+reqlen); // 时间复杂度 O(N^2)
        p = zl+offset;
    }

    // 将前置节点的长度写入新节点的 header
    p += zipPrevEncodeLength(p, prevlen);
    // 将节点值的长度写入新节点的 header
    p += zipEncodeLength(p,encoding, slen);
    // 写入节点值
    if (ZIP_IS_STR(encoding)) {
        memcpy(p,s,slen);
    } else {
        zipSaveInteger(p,value,encoding);
    }

    // 更新列表的节点数量计数器
    ZIPLIST_INCR_LENGTH(zl,1);

    return zl;
}
```


### ziplist 与 hash

hash 是 redis 中可以用来存储一个对象结构的比较理想的数据类型. 一个对象的各个属性, 正好对于hash结构的各个 field.

实际上, hash 随着数据的增大, 其底层数据结构的实现是会发生变化的, 当然存储效率也就不想同. 在 field 比较少, 各个 value
值也比较小时, hash 采用 ziplist 实现; 随着 field 增多和 value 值增大, hash 可能会变成 dict 来实现.

当随着数据的插入, hash 底层数据结构的这个 ziplist 就可能转换为 dict. 到底插入多少才会转换呢?

```
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
```

上述的两个配置, 在如果满足下面两个条件之一, ziplist 会转换成 dict:

- 当 hash 当中是数据项 (即field-value对) 的数目超过 512 时, 也就是 ziplist 的数据项超过 1024 时候.

- 当 hash 中插入的任意一个 value (这里的插入的value是hash的 field 或 value 的长度) 的长度超过 64 的时候.

