## dict

dict 为了实现增量式重哈希(incremental rehashing), dict 的数据结构包括两个哈希表. 在重哈希期间, 数据从第一个哈希表
向第二个哈希表迁移.

```cgo
typedef struct dictEntry {
    void* key;
    union {
       void* val;
       uint64_t u64;
       int64_t  i64;
       double d;
    } v;
    struct dictEntry* next;
} dictEntry;

typedef struct dictType {
    unsigned int (*hashFunction)(const void* key);     // 哈希算法
    void* (*keyDup)(void* privdata, const void* key);  // 拷贝函数
    void* (*valDup)(void* privdata, const void* obj);
    int   (*keyCompare)(void* privdata, const void* key1, const void* key2); // 
    void  (*keyDestructor)(void* privdata, void* key);
    void  (*valDestructor)(void* privdata, void* obj);
} dictType;


typedef struct dictht {
    dictEntry** table;
    unsigned long size;     // 数组长度, 2的整数幂
    unsigned long sizemask; // 数组的mask
    unsigned long used;     // hashtable 当中存储的元素的个数
} dictht;

typedef struct dict{
    dictType* type; // 自定义的方式使得 dict 的 key 和 value 能够存储任何类型.
    void* privdata; // 私有数据指针. 由调用者在创建 dict 时传入.
    dictht ht[2];   // 哈希表, 在进行rehash时, ht[0] 和 ht[1] 都有效, 平常状况下,  只有 ht[0] 有效
    long rehashidx; // 在进行 rehash 的时候使用, 正常状况下值是-1
    int  iterators; // 当前正在遍历的 iterators 的个数.
} dict;
```