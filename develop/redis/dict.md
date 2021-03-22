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


```cgo
/* 
 * 在字典不存在安全迭代器的情况下, 对字典进行单步 rehash .
 *
 * 字典有安全迭代器的情况下不能进行 rehash, 因为两种不同的迭代和修改操作可能会弄乱字典.
 *
 * 这个函数被多个通用的查找, 更新操作调用, 它可以让字典在被使用的同时进行 rehash.
 */
static void _dictRehashStep(dict *d) {
    if (d->iterators == 0) dictRehash(d,1);
}

/*
 * 执行 N 步渐进式 rehash.
 *
 * 返回 1 表示仍有键需要从 0 号哈希表移动到 1 号哈希表, 
 * 返回 0 则表示所有键都已经迁移完毕.
 *
 * 注意: 每步 rehash 都是以一个哈希表索引(桶)作为单位的, 一个桶里可能会有多个节点, 被 rehash 的桶里的所有节点都
 * 会被移动到新哈希表.
 *
 * T = O(N)
 */
int dictRehash(dict *d, int n) {
    // 只可以在 rehash 进行中时执行
    if (!dictIsRehashing(d)) return 0;

    // 进行 N 步迁移
    while(n--) {
        dictEntry *de, *nextde;

        // 如果 0 号哈希表为空, 那么表示 rehash 执行完毕
        if (d->ht[0].used == 0) {
            // 释放 0 号哈希表
            zfree(d->ht[0].table);
            // 将原来的 1 号哈希表设置为新的 0 号哈希表
            d->ht[0] = d->ht[1];
            // 重置旧的 1 号哈希表
            _dictReset(&d->ht[1]);
            // 关闭 rehash 标识
            d->rehashidx = -1;
            // 返回 0, 向调用者表示 rehash 已经完成
            return 0;
        }

        // 确保 rehashidx 没有越界.
        assert(d->ht[0].size > (unsigned)d->rehashidx);

        // 略过数组中为空的索引, 找到下一个非空索引
        while(d->ht[0].table[d->rehashidx] == NULL) d->rehashidx++;

        // 指向该索引的链表表头节点
        de = d->ht[0].table[d->rehashidx];
        
        // 将链表中的所有节点迁移到新哈希表
        while(de) {
            unsigned int h;

            // 保存下个节点的指针
            nextde = de->next;

            // 计算新哈希表的哈希值, 以及节点插入的索引位置
            h = dictHashKey(d, de->key) & d->ht[1].sizemask;

            // 插入节点到新哈希表
            de->next = d->ht[1].table[h];
            d->ht[1].table[h] = de;

            // 更新计数器
            d->ht[0].used--;
            d->ht[1].used++;

            // 继续处理下个节点
            de = nextde;
        }
        // 将刚迁移完的哈希表索引的指针设为空
        d->ht[0].table[d->rehashidx] = NULL;
        // 更新 rehash 索引
        d->rehashidx++;
    }

    return 1;
}
```