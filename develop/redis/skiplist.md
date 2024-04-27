## Redis - skiplist

### skiplist 数据结构

普通链表:

![images](/images/develop_redis_sk_list.png)


多层链表(跳跃表):

![images](/images/develop_redis_sk_sklist.png)


skiplist 正是受这种多层链表的想法的启发而设计出来的. 实际上, 按照上面的多层链表的方式, 上一层链表的节点个数是下一层链表
节点个数的一半, 这样查询给就非常类似一个二分查找. 如果严格维护2:1对应的关系, 那么插入和删除的时间复杂度将蜕变到O(n).

skiplist 为了避免这一问题, 它不要求上下相邻两层链表之间个数有严格对应关系, 而是随机为每一个节点随机出一个层数.

```cgo
// skiplist 的节点
typedef struct zskiplistNode {
    robj   *obj;   // 成员对象
    double score;  // 分值
    struct zskiplistNode *backward; // 后退指针

    // 层
    struct zskiplistLevel {
        struct zskiplistNode *forward; // 前进指针
        unsigned int span;             // 跨度, span用于计算元素排名(rank)
    } level[];
} zskiplistNode;

// skiplist 头节点
typedef struct zskiplist {
    struct zskiplistNode *header, *tail;
    unsigned long length;
    int level;
} zskiplist;


typedef struct zset {
    // 字典, 键为成员, 值为分值
    dict *dict;

    // 跳跃表, 按分值排序成员
    // 用于支持平均复杂度为 O(log N) 的按分值定位成员操作以及范围操作
    zskiplist *zsl;
} zset;
```

> 注意: 
> 1. 在 skiplist 的节点中存储了 robj 和 score, 按照 score 的分数进行排序. robj 是 member(整数, 字符串)
> 2. zset 的编码有两种, 使用的是 ziplist 或 skiplist (也就是上述的 zset 结构, 这时使用了 dict 存储 member->score 
的映射关系, 方便快速查找 score), 编码转换的条件后面会介绍到的.


在 Go当中, 设计出 skiplist 的数据结构:

两个常识: 

**第 i 层的 node, 其 next 有 i 个后续节点.**

**查询是自上而下, 自左向右进行的**        

```
type Node struct{
    val interface{}  // 当前节点的值
    next []*Node     // 指向下一个节点的数组
}

type SkipList struct{
    header *Node  // 保存所有的头节点, 不存储任何值 
    level int     // 表示当前的level
}

func insert(sk *SkipList, val int)  {
    prefix := map[int]*Node{} // 前驱节点
    node := sk.header
    // 查询顺序: 自上而下, 自左向右 
    for level:=sk.level-1; level>=0; level-- {
        for {
            if node.next[level] == nil {
                break
            }
            
            if node.next[level].val.(int) < val {
                node = node.next[level] // 自左向右
            } else {
                break
            }
        }
        
        prefix[level] = node // 记录前驱节点
    }
    
    // 插入
    level := random()
    node := &Node{
        val: val
        next: make([]*Node, level)
    }
    
    if level > sk.level {
        for i:=level-1; i > sk.level-1; i-- {
            sk.header.next[i] = node
        }
    }
    
    for level := sk.level-1; level >= 0; level-- {
        pre := prefix[level]
        node.next[level] = pre.next[level]
        pre.next[level] = node
    }
    
    if level > sk.level {
        sk.level = level
    }
}

func delete(sk *SkipList, val int) {
    prefix := map[int]*Node{} // 前驱节点
    node := sk.header
    // 查询顺序: 自上而下, 自左向右 
    for level := sk.level-1; level>=0; level-- {
        for {
            if node.next[level] == nil {
                break
            }
            if node.next[level].val == val {
                prefix[i] = node       // 记录删除节点的前驱节点
                break
            } else if node.next[level].val.(int) < val {
                node = node.next[level] // 自左向右
            } else {
                break
            }
        }
    }
    
    for i, v := range update {
        if v == skipList.Header {
            skipList.Level--
        }
        v.next[i] = v.next[i].next[i]
    }
}
```


skiplist 与 平衡树, 哈希表比较:

1. skiplist和各种平衡树(AVL, 红黑树等)的元素是有序的, 哈希表是无序的. 

2. 范围查找, 平衡树比 skiplist 操作要复杂. 平衡树, 在找到指定范围的小值之后, 还需要以中序遍历查找不超过最大值的节点.
skiplist上进行范围查找, 只需要找到最小值之后, 在对第一层链表上进行遍历即可实现.

3. 插入和删除, 会引起平衡树的调整, 而 skiplist 的插入和删除只需要修改相邻节点的指针.

4. 内存, skiplist比平衡树更灵活. 平衡树每个节点平均包含两个指针, skiplist每个节点的平均指针是 1/(1-p), p是每一层的
概率.

5. 查找, 基本相当.


> redis zset的实现:

- 当数据较少时, zset是由一个ziplist来实现的.

- 当数据较多时, zset是由一个 dict + skiplist 实现的. dict 查询数据到分数的对应关系, 而 skiplist 可以根据分数查找
数据.


zset 命令:

```
zadd key score member  // 添加

zincr key incr member // member 分数增加

zcount key min max    // count 指定区间分数的成员数
zlexcount key min max // count 指定字典区间内的成员数

zrange key start stop     // 指定排名的成员
zrangebylex key min max   // 指定字典区间的成员
zrangebyscore key min max // 指定分数区间内的成员

zrevrange key start stop // 指定区间的成员(倒序)
zrevrangebyscore key max min // 指定分数区间的成员(倒序)
zrevrangebylex key max min // 指定字典区间的成员(倒序)

zrem key member // 删除 member
zremrangebylex key min max // 删除字典区间的成员
zremrangebyscore key min max // 删除区间分数的成员
zremrangebyrank key min max // 删除区间排名的成员

zrank key member // member的排名
zrevrank key member // member排名(倒序)

zcore key member // 分数

zscan key cursor // 迭代集合的元素(成员, 分数)  

zcard key // 集合成员数
```

在使用 skiplist+dict 方式下编码 zset:

- zscore 的查询, 由 dict 来提供的(当编码是 ziplist 是可以查询获取的).

- 为了支持排名(rank), Redis 对 skiplist 做了扩展, 使得根据 rank 能够快速查找到 member, 或根据 score 查到 member
之后, 也同样获得 rank. 根据 rank 查找, 时间复杂度都是 O(lgn).

- `zrevrange start stop` 的查询, 是根据 rank 查 member, 扩展后的 skiplist 提供.

- `zrevrank member` 是先在dict当中查找 score, 然后拿 score 到 skiplist 中查找, 查找到也同时获得了 rank.

> 关于时间复杂度, zscore 是 O(1), zrevrank, zrevrange, zrevrangebyscore. 要查询 skiplist, zrevrank的时间复
杂度是 O(lgn),  zrevrange, zrevrangebyscore 时间复杂度是 O(lgn + M)


### skiplist 与 zset
 
涉及的两个参数:

```
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
```

zset 当数据量较少的时候, 使用 ziplist 实现, 当数据量大的时候, zset 使用 zset 数据结构(skiplist+dict) 进行实现的.

基于 ziplist 实现 zset 主要是节省内存. 由于 zset 的每一项是由 member 和 score 组成的, 因此在 zadd 插入一条数据的  
时候, 底层的 ziplist 就会插入两个数据项, member在前, score在后. 查找的时候, 当然是只能顺序查找, 每一步前进两个数据项.

当满足下面两个条件之一, zset 会从 ziplist 转向 skiplist:

- 当 zset 中的元素数量个数, 即(member, score)对的数目超过 `zset-ziplist-max-entries` 的值.

- 当 zset 中插入任意一个member的长度超过 `zset-ziplist-max-value` 时