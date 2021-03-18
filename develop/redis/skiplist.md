## skiplist

涉及的两个参数:

```
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
```

普通链表:

![images](/images/develop_redis_sk_list.png)


多层链表(跳跃表):

![images](/images/develop_redis_sk_sklist.png)


skiplist 正是受这种多层链表的想法的启发而设计出来的. 实际上, 按照上面的多层链表的方式, 上一层链表的节点个数是下一层链表
节点个数的一半, 这样查询给就非常类似一个二分查找. 如果严格维护2:1对应的关系, 那么插入和删除的时间复杂度将蜕变到O(n).

skiplist 为了避免这一问题, 它不要求上下相邻两层链表之间个数有严格对应关系, 而是随机为每一个节点随机出一个层数.

```cgo
typedef struct zskiplistNode {
    robj   *obj;   // 成员对象
    double score;  // 分值
    struct zskiplistNode *backward; // 后退指针

    // 层
    struct zskiplistLevel {
        struct zskiplistNode *forward; // 前进指针
        unsigned int span;             // 跨度
    } level[];
} zskiplistNode;

typedef struct zskiplist {
    struct zskiplistNode *header, *tail;
    unsigned long length;
    int level;
} zskiplist;
```


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


Redis sorted set的实现:

- 当数据较少时, sorted set是由一个ziplist来实现的

- 当数据较多时, sorted set是由一个 dict + skiplist 实现的. dict 查询数据到分数的对应关系, 而 skiplist 可以根据
分数查找数据.


sorted与skiplist的关系:

- zscore 的查询, 是由 dict 实现的.

- 为了支持排名(rank), Redis 对 skiplist 做了扩展, 使得根据排名能够快速查找到数据, 或根据分数查到数据之后, 也同样获
得排名. 根据排名查找, 时间复杂度也是 O(lgn)

- zrevrage的查询, 是根据排名查数据, 扩展后的 skiplist 提供

- zrevrank是现在dict当中查找分数, 然后拿分数到 skiplist 中查找, 查找到也同时获得了排名.

- 关于时间复杂度, zscore 是 O(1), zrevrank, zrevrange, zrevrangebyscore 要查询 skiplist, zrevrank的时间复
杂度是 O(lgn),  zrevrange, zrevrangebyscore 时间复杂度是 O(lgn + M)

