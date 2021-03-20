## quicklist

涉及的两个参数:

```
list-max-ziplist-size -2
list-compress-depth 0
```

quicklist 概述:

list 具有这样的特点: 它是一个有序列表, 便于在表的两端追加和删除数据, 而对于中间位置的存取具有 O(N) 的时间复杂度. 这就
是一个双向链表具有的特点.

list 的内部实现 quicklist 正是一个双向链表, 而且是一个 ziplist 的双向链表.

quicklist 的结构设计, 是一个空间和时间的折中:

- 双向链表便于在表的两端进行 push 和 pop 操作, 但是它的内存开销比较大.(节点需要保存两个指针; 节点内存独立, 地址不连续,
容易产生内存碎片)

- ziplist 由于是一块连续内存, 存储效率很高. 但是不利于修改, 每次数据变动都会引发一次内存 realloc. 尤其是当 ziplist 
长度很长时, realloc 导致大量数据拷贝, 性能降低.

quicklist 的节点上的 ziplist 需要保存一个合理的长度, 才能发挥最大性能, 这个长度取决于不同的场景. Redis 提供了一个配
置参数 `list-max-ziplist-size`, 可以让使用者进行自行调整.

```
list-max-ziplist-size -2
```

该参数的取值: 

如果是正数, 表示节点 ziplist 存储的 entry 的最大个数;
 
如果是负数, 则表示按照占用字节数限定节点 ziplist的长度. 取值是范围是 `[-1, -5]`, `-5`(每个quicklist节点上的ziplist
的大小不能超过64kb), `-4`(32kb), `-3`(16kb), `-2`(8kb), `-1`(4kb).

列表的设计目标是能够用来存储很长的数据列表的. 当列表很长时, 最容易被访问的很可能是两端的数据, 中间的数据被访问的频率很低.
如果应用符合这个场景, list 还提供了一个选项, 将中间节点数据进行压缩, 从而节省空间. 配置参数 `list-compress-depth`

```
list-compress-depth 0
```

该参数表示 quicklist 两端不被压缩的节点个数. 0 表示都不被压缩(默认值). N 表示 quicklist 两端各有 N 个节点(ziplist)
不被压缩, 中间的节点压缩.

quicklist 定义

```cgo
typedef struct quicklistNode {
    struct quicklistNode *prev;
    struct quicklistNode *next;
    
    unsigned char *zl; // 如果当前几点没有被压缩, 指向ziplist; 否则, 指向 quicklistLZF
    unsigned int sz;   // zl指向的ziplist的总大小.(不论当前节点是否被压缩)
    unsigned int count:16;    // ziplist 当中的数据项个数
    unsigned int encoding:2;  // RAW=1, LZF=2
    unsigned int container:2; // 预留字段, 固定值2. NONE=1, ZIPLIST=2
    unsigned int recompress:1; // 当查看当前节点被压缩时, 会对当前数据进行解压缩, 该字段设为1作为一个标记, 后续压缩
    unsigned int attempted_compress:1; // 自动化使用
    unsigned int extra: 10; // 扩展字段
} quicklistNode;

typedef struct quicklist {
    quicklistNode* head;
    quicklistNode* tail;
    unsigned long count; // 存储的元素个数
    unsigned int len;    // ziplist 节点个数
    int fill:16;
    unsigned int compress:16; // 两端未被压缩的几点个数
} quicklist;


typedef struct quicklistLZF {
    unsigned int sz; // 压缩后的ziplist大小
    char compresss[];
} quicklistLZF;
```

