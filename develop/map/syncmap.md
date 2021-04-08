# go sync.Map 源码解析

sync.Map 设计思路: 

1.空间换取时间, 通过冗余的两个数据结构(read, dirty), 减少加锁对性能影响.

2.动态调整, miss次数多了之后, 将dirty的数据提升为read中. 这里只是指针的拷贝, 很快.

3.double checking, 尽量减少加锁.

4.延迟删除, 删除一个key值时只是打上标记, 在迁移dirty数据的时候才清理删除的数据(这个发生在Store的时候, 这里的数据的迁移
是对底层的数据的拷贝, 可能是性能问题).

5.优先从read当前读取,更新,删除. read的读取是不需要锁的

Map类型针对两种常见用例进行了优化:

(1)给定键的条目仅写入一次但多次读取(例如仅在增长的高速缓存中),

(2)当多个goroutine读取, 写入时, 并覆盖不相交的键集的条目. 

在这两种情况下, 与与单独的 Mutex 或 RWMutex 配对的Go映射相比, 使用Map可以显着减少锁争用.


> 注意: Map的零值就可以开始使用,  但是首次使用后不得复制 Map.

## 数据结构

```cgo
type Map struct {
	// 当涉及到 dirty 数据的操作的时候, 需要使用这个锁
	mu Mutex

	// 一个只读的数据结构, 因为只读, 所以不会有读写冲突. 所以从这个数据中读取总是安全的.
	// 实际也会更新这个数据的 entries, 如果 entry 是未删除的(unexpunged), 并不需要加锁.
	// 如果entry已经被删除了, 需要加锁, 以便更新 dirty 数据.
	read atomic.Value // readOnly
	
	
	// dirty 数据包含当前的 map 包含的 entries, 它包含最新的 entries(包括read种未删除的数据, 虽有冗余, 但是提升
	// dirty 字段为 read 的时候非常快, 不用一个一个的复制, 而是直接将这个数据结构作为 read 字段的一部分), 有些数据
	// 还可能没有移动到 read 字段中.
	//
	// 对于 dirty 的操作需要加锁, 因为对它的操作可能会有读写竞争.
	// 
	// 当 dirty 为空的时候, 比如初始化或者刚提升完, 下一次的写操作会复制read字段中未删除的数据到这个数据中.
	dirty map[interface{}]*entry

	// 当从Map中读取entry的时候, 如果 read 中不包含这个 entry, 会尝试从 dirty 当中读取, 这个时候会将 misses 加1,
	// 当 misses 累计到 dirty 的长度的时候, 就会将 dirty 提升为 read, 避免从 dirty 中 miss 太多次. 因为操作 
	// dirty 需要加锁.
	misses int
}
```

使用了冗余是数据结构 `read`, `dirty`. `dirty` 中会包含 `read` 中为删除的 `entries`, 新增加的 `entries` 会加入
到 `dirty` 中.

`read` 是数据结构:

```cgo
type readOnly struct {
	m       map[interface{}]*entry
	amended bool // 如果 Map.dirty 包含了一些在 readOnly.m 不存在 key, 这个值为 true.
}
```

`amended` 表示 `Map.dirty` 包含了一些在 `readOnly.m` 不存在 key的状况, 所以如果从 `Map.read` 找不到的数据的话, 
还要进一步到 `Map.dirty` 中查找.


```cgo
// entry 是对应于特定 key 的映射中的 solt
type entry struct {
	// p指向为 entry 中存储的 interface{}值.
    //
    // 如果 p == nil, 则该 entry 已被删除, 并且 m.dirty == nil.
    //
    // 如果 p == expunged, 则该 entry 已被删除, m.dirty != nil, 并且 m.dirty 中不存在该entry.
    //
    // 其他, 该 entry 有效并记录在 m.read.m[key] 中; 如果m.dirty != nil, 则记录在 m.dirty[key] 中.
    //
    // 可以通过用nil进行原子替换来删除entry: 下次创建 m.dirty 时, 它将自动用 expunged 替换nil并使 m.dirty[key]
    // 保持未设置状态.
    //
    // 如果 p != expunged, 则可以通过原子替换来更新条目的关联值. 
    // 如果 p == expunged, 则只有在首先设置 m.dirty[key] = e 之后才能更新entry的关联值, 以便使用 dirty 查找该
    // entry.
	p unsafe.Pointer // *interface{}
}
```


## 查询

```cgo
// Load 返回存储在Map中的键值, 如果没有值, 则返回nil.
// ok 表明是否在Map中找到了值.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	
	// 如果没有找到, 并且read.amended为true, 说明dirty中有新数据, 从dirty中查找, 需要加锁
	if !ok && read.amended {
		m.mu.Lock()
		
		// 双检查, 避免加锁的时候, m.dirty 提升为 m.read, 这个时候 m.read 可能被替换了
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		
		// 如果 m.read 中还是不存在, 并且 m.dirty 中有新的数据
		if !ok && read.amended {
		    // 从 m.dirty 中查找
			e, ok = m.dirty[key]
			// 不管 m.dirty 中存不存在, 都将 misses 计数加1. missLocked() 中满足条件后就会提升 m.dirty
			m.missLocked()
		}
		m.mu.Unlock()
	}
	
	if !ok {
		return nil, false
	}
	
	return e.load()
}

// load, 原子加载 entry 的值
func (e *entry) load() (value interface{}, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expunged {
		return nil, false
	}
	return *(*interface{})(p), true
}


// 加锁的状况下, 更新 misses 的值
func (m *Map) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	
	// m.misses >= len(m.dirty), 需要将 dirty 提升为 read, 数据指针的拷贝, 速度很快.
	m.read.Store(readOnly{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}
```


## 存储

```cgo
// 存在
func (m *Map) Store(key, value interface{}) {
	read, _ := m.read.Load().(readOnly)
	// 当前的 key 在 read 当中已经存在, 尝试去更新它. 否则需要加锁更新.
	if e, ok := read.m[key]; ok && e.tryStore(&value) {
		return
	}
    
	m.mu.Lock()	
	read, _ = m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
	    // read 当中存在 key
	    // e 标记删除, 将 e.p 设置为 nil, 需要重新更新 dirty 当中的对于的值 
		if e.unexpungeLocked() {
			// 该 entry 先前已删除, 这意味着存在一个非零的 dirty 映射, 并且该 entry 不在其中.
			m.dirty[key] = e
		}
		e.storeLocked(&value) // 存储
	} else if e, ok := m.dirty[key]; ok {
	    // dirty 当中存在 key
		e.storeLocked(&value)
	} else {
	    // read 和 dirty 当中都不存在该 key, 并且此时的 read 和 dirty 是没有差异的
		if !read.amended {
			// 我们将第一个新键添加到 dirty 映射.
            // 确保已分配它, 并将 read 映射标记为存在差异.
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		// dirty 当中存储 value
		m.dirty[key] = newEntry(value)
	}
	m.mu.Unlock()
}


// 如果entry尚未被清除, 则tryStore将存储一个值.
//
// 如果删除了 entry, 则tryStore返回false并使该 entry 保持不变.
func (e *entry) tryStore(i *interface{}) bool {
	for {
		// 加载指针
		p := atomic.LoadPointer(&e.p)
		
		// 情况2: entry 已经被删除了
		if p == expunged {
			return false
		}
		
		// 情况1: entry 尚未被删除, 使用CAS保存i的值
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return true
		}
	}
}


// unexpungeLocked 确保该 entry 未标记为 expunged(已删除).
//
// 如果该 entry 先前已删除, 则必须在解锁 m.mu 之前将其添加到 dirty 映射中.
func (e *entry) unexpungeLocked() (wasExpunged bool) {
    // 如果 e.p == expunged, 则将其设置为 nil, 返回 true;
    // 否则保持 e.p 值不变, 返回false
	return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}


// 加锁状况下存储 entry, 该 entry 必须未被标记删除的
func (e *entry) storeLocked(i *interface{}) {
	atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

// 加锁状况下, 从 read 当中迁移数据到 dirty
func (m *Map) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnly)
	m.dirty = make(map[interface{}]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

// 判断 entry 是否被删除, 如果被删除, 则不需要迁移到 dirty 
func (e *entry) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
	    // CAS, 
	    // 如果 e.p == nil, 则将其设置为 expunged(标记删除), 返回值是 true
	    // 否则, 保持不动, 返回值是false
		if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
			return true
		}
		
		// 再次加载 e.p
		p = atomic.LoadPointer(&e.p)
	}
	
	return p == expunged
}
```


## 删除


```cgo
func (m *Map) Delete(key interface{}) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	
	// read 当中不存在 key, 并且 read 和 dirty 存在差异, 需要从 dirty 当中移除
	if !ok && read.amended {
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		// 加锁状况下, 双重确认
		if !ok && read.amended {
			delete(m.dirty, key) // 从 dirty 当中移除 key
		}
		m.mu.Unlock()
	}
	
	if ok {
		e.delete() // 删除 entry
	}
}

func (e *entry) delete() (hadValue bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		// entry 已经删除(标记删除和真正的删除)
		if p == nil || p == expunged {
			return false
		}
		// 删除, 这里是删除, 不是标记删除
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return true
		}
	}
}
```


## 遍历

```cgo
// 如果f返回false, 则range停止迭代.
//
// Range不一定与Map内容的任何一致快照相对应: 一个key最多会被访问一次, 但是如果同时store或delete key, Range可能会
// 在此期间从任何点反映该键的任何映射范围调用.
//
// Range时间复杂度可能是O(N), N为Map中的元素数. 即使在恒定的调用次数后f返回false, 也是如此.
func (m *Map) Range(f func(key, value interface{}) bool) {
	read, _ := m.read.Load().(readOnly)
	
	// 如果 dirty 当中有新的数据, 则提升 dirty 为 read, 然后遍历
	if read.amended {
		// 提升 dirty
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly) // 双检查
		if read.amended {
			read = readOnly{m: m.dirty}
			m.read.Store(read) // 将 dirty 提升为 read. 注意: 这里的提升不存在底层数据的拷贝, 因此速度很快.
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}
    
    // 遍历 read, 这是安全的遍历
	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}
```
