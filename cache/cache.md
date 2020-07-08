# Golang Cache 比较解析

golang 当中比较缓存分别是 freecache, bigcache, groupcache. 以下主要针对这三个cache进行源码分析.

## bigcache

项目: github.com/allegro/bigcache

实现的思路:

通过 hash 的方式进行分片. 每一个分片包含一个 map 和一个 ring buffer. 无论如何添加数据, 都会将它放置到对应的 ring 
buffer 中, 并且将位置保存在 map 中. 如果多次设置相同的元素, 则 ring buffer 中的旧值则会标记为无效.

每一个 map 的key都是一个 uint64 的 hash 值, 每个值对应一个存储着元数据的 ring buffer. 如果 hash 值碰撞了, 
bigcache 会忽略旧值 key (将旧的 hash 重置为0), 然后把新的值存储到 map 中. 使用 `map[uint64]uint32` 是避免 GC 
扫描.

存储: 

header = timestamp(8B) + hash(8B) + keysize(2B) + key + value

真正存储的内容是 `varint(header) + header` 

    
数据结构:

```cgo
type BigCache struct {
	shards       []*cacheShard // 每个shared都有一个读写锁
	lifeWindow   uint64 // 缓存时间窗口
	clock        clock  // 时钟
	hash         Hasher // hash算法
	config       Config
	shardMask    uint64 // shred mask
	maxShardSize uint32 // shred的内存大小, 0表示内存没有限制的, 默认值是0
}
```

1. 首先 shards 的个数是 2 的指数级别. 

2. 其次, 缓存的时间窗口 LifeWindow, 缓存 k-v 的 TTL 是在区间 `[LifeWindow, CleanWindow+LifeWindow]`, 无法
为每个 k-v 单独设置. 缓存失效或者删除不会导致大的内存移动. 

3. 清理时间窗口 CleanWindow, 每隔多长时间对 LifeWindow 进行移动. 

4. 最后, 可以限制缓存使用内存大小.


### 重要的方法

- Set 函数

````cgo
func (c *BigCache) Set(key string, entry []byte) error {
    // 根据key, 产生一个 uint64 的 hash 值
	hashedKey := c.hash.Sum64(key)
	
	// hashedKey & sharedMask 得到分片的位置
	shard := c.getShard(hashedKey)
	
	return shard.set(key, hashedKey, entry)
}
````

> shared(分片)的 set 方法, 带有锁.

```cgo
func (s *cacheShard) set(key string, hashedKey uint64, entry []byte) error {
	currentTimestamp := uint64(s.clock.epoch()) // 当前的时间, uint64
    
    // 全程加锁处理
	s.lock.Lock()

	// 已经存在的hash值, 重置, 非删除
	if previousIndex := s.hashmap[hashedKey]; previousIndex != 0 {
		if previousEntry, err := s.entries.Get(int(previousIndex)); err == nil {
		    // 重置, 就是将 timestamp 后面的 hashkey 设置为 0 [底层]
			resetKeyFromEntry(previousEntry)
		}
	}

	// 检查 "队列" 头部 [窗口TTL的好处, 有效的值总是在一个范围内], 如果过期, 使用 removeOldestEntry 进行删除
	if oldestEntry, err := s.entries.Peek(); err == nil {
		s.onEvict(oldestEntry, currentTimestamp, s.removeOldestEntry)
	}

	// 进行包装 timestamp (uint64), hashkey (uint64), key, entry
    // header + key + value
    // 其中 header 的内容是 timestamp(8) + hashkey(8) + keysize(2)
    // 使用了 binary.LittleEndian 的方式写入内容[节省内存空间]
    // 真正存储的值方式
    // varint[可变长度的data大小] + data [包装后的内容, 即 header(18) + key + value]
    // varint 一是为了节省空间, 二是为了读取内容的时候能够获取到全部内容
	w := wrapEntry(currentTimestamp, hashedKey, key, entry, &s.entryBuffer)

	// 进队列, 如果内存不足, 则删除头元素[最老的元素], 以使得当前缓存入队列
	for {
	    // 下面的有解析
		if index, err := s.entries.Push(w); err == nil {
			s.hashmap[hashedKey] = uint32(index) // hashkey <-> index [push的开始位置]
			s.lock.Unlock()
			return nil
		}
		if s.removeOldestEntry(NoSpace) != nil {
			s.lock.Unlock()
			return fmt.Errorf("entry is bigger than max shard size")
		}
	}
}
```

> 底层 ring bufer 的 Push 操作. **比较复杂**

````cgo
func (q *BytesQueue) Push(data []byte) (int, error) {
    // 计算所需的字节大小 headerEntrySize + dataLen
	dataLen := len(data)
	headerEntrySize := getUvarintSize(uint32(dataLen))
    
    // 查找位置
	if !q.canInsertAfterTail(dataLen + headerEntrySize) {
		if q.canInsertBeforeHead(dataLen + headerEntrySize) {
			q.tail = leftMarginIndex
		} else if q.capacity+headerEntrySize+dataLen >= q.maxCapacity && q.maxCapacity > 0 {
			return -1, &queueError{"Full queue. Maximum size limit reached."}
		} else {
			q.allocateAdditionalMemory(dataLen + headerEntrySize)
		}
	}

	index := q.tail
    
    // 存储数据
	q.push(data, dataLen)

	return index, nil
}
````

> 默认的 hash 算法 (fnv):

```cgo
const (
	// offset64 FNVa offset basis. 
	offset64 = 14695981039346656037
	// prime64 FNVa prime value. 
	prime64 = 1099511628211
)

// Sum64 gets the string and returns its uint64 hash value.
func (f fnv64a) Sum64(key string) uint64 {
	var hash uint64 = offset64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime64
	}

	return hash
}
```


- Get 函数

```cgo
func (c *BigCache) Get(key string) ([]byte, error) {
	hashedKey := c.hash.Sum64(key)
	shard := c.getShard(hashedKey)
	return shard.get(key, hashedKey)
}
```

```cgo
func (s *cacheShard) get(key string, hashedKey uint64) ([]byte, error) {
	s.lock.RLock() // 可读锁的基础上完成操作
	
	// keyhash -> 存储的index
	// 存储的 index 直接能读取到存储的内容, 即包装的内容
	wrappedEntry, err := s.getWrappedEntry(hashedKey)
	if err != nil {
		s.lock.RUnlock()
		return nil, err
	}

	// 读取包装内容当中的hashkey的值, hash冲突导致获取的元素非预期的元素
	if entryKey := readKeyFromEntry(wrappedEntry); key != entryKey {
		s.lock.RUnlock()
		s.collision()
		if s.isVerbose {
			s.logger.Printf("Collision detected. Both %q and %q have the same hash %x", key, entryKey, hashedKey)
		}
		return nil, ErrEntryNotFound
	}
	
	// 读取具体的内容
	entry := readEntry(wrappedEntry)
	s.lock.RUnlock()
	s.hit(hashedKey) // 命中

	return entry, nil
}
```

> ring buf 根据 index 获取内容

```cgo
// 获取 index 位置的原始内容
func (q *BytesQueue) peek(index int) ([]byte, int, error) {
	err := q.peekCheckErr(index)
	if err != nil {
		return nil, 0, err
	}

	blockSize, n := binary.Uvarint(q.array[index:])
	return q.array[index+n: index+n+int(blockSize)], n, nil
}

// 检查 index 是否合法
func (q *BytesQueue) peekCheckErr(index int) error {
	if q.count == 0 {
		return errEmptyQueue
	}

	if index <= 0 {
		return errInvalidIndex
	}

	if index >= len(q.array) {
		return errIndexOutOfBounds
	}
	return nil
}
```


- Delete 方法

```cgo
func (c *BigCache) Delete(key string) error {
	hashedKey := c.hash.Sum64(key)
	shard := c.getShard(hashedKey)
	return shard.del(hashedKey)
}
```

> shared 删除逻辑, 没有真正的立刻移除底层数据

```cgo
func (s *cacheShard) del(hashedKey uint64) error {
	// 读锁定下快速判断缓存是否存在
	s.lock.RLock()
	{
		itemIndex := s.hashmap[hashedKey]

		if itemIndex == 0 {
			s.lock.RUnlock()
			return ErrEntryNotFound
		}

		if err := s.entries.CheckGet(int(itemIndex)); err != nil {
			s.lock.RUnlock()
			return err
		}
	}
	s.lock.RUnlock()

	// 在缓存存在的状况下, 使用写锁进行删除, 删除在当前shared当中的hash, 以及重置hashkey
	// "真正的移除元素内容是在过期的时间点, 或者队列满了的状况下进行移除"
	s.lock.Lock()
	{   
	    // 再次查询, 避免已经被移除的情况, 并发考量
		itemIndex := s.hashmap[hashedKey]
		if itemIndex == 0 {
			s.lock.Unlock()
			return ErrEntryNotFound
		}
        
        // 获取内容
		wrappedEntry, err := s.entries.Get(int(itemIndex))
		if err != nil {
			s.lock.Unlock()
			return err
		}
        
        // 从 hashmap 当中移除 hash
		delete(s.hashmap, hashedKey)
		s.onRemove(wrappedEntry, Deleted) // 删除操作的回调函数
		if s.statsEnabled {
			delete(s.hashmapStats, hashedKey) // 统计信息的移除
		}
		resetKeyFromEntry(wrappedEntry) // 将 ring buf 的当中元素的 hashkey 重置
	}
	s.lock.Unlock()

	return nil
}
```