# Golang Cache 比较 - freecache

golang 当中比较缓存分别是 freecache, bigcache, groupcache. 以下主要针对这三个cache进行源码分析.

## freecache

项目: github.com/coocood/freecache

实现的思路:

freecache 通过减少指针的数量从而避免过多的 GC. 不管系统保存多少指针, 这里只有 512 个指针. 数据利用 key-value 的形
式被分配到 256 个段(segment)中. 每个段中包含两个指针, 一个是循环链用于存储 key-values, 另外一个是片的索引用于全局
搜索. 每个段都有独立的锁, 支持高并发访问.

freecache是一个近似LRU的算法. LRU是内存管理的一种方式, 即内存页置换方式, 对于已经分配的但是没有使用的数据进行缓存, 当
缓存满了之后, 按照一定的规则顺序对存储的数据进行置换.


![image](/images/develop_freecache.png)

数据结构:

```cgo
segmentCount = 256

type Cache struct {
    locks    [segmentCount]sync.Mutex
    segments [segmentCount]segment
}

type segment struct {
    rb            RingBuf // ring buffer 
    segId         int
    _             uint32
    missCount     int64
    hitCount      int64
    entryCount    int64      // 
    totalCount    int64      // number of entries in ring buffer, including deleted entries.
    totalTime     int64      // 用于计算最近最少使用的条目. 总的过期时长
    timer         Timer      // Timer giving current time
    ...
    vacuumLen     int64      // up to vacuumLen, new data can be written without overwriting old data.
    slotLens      [256]int32 // 单个 slot 实际存储 entryPtr 个数
    slotCap       int32      // 单个 slot 可以存储最大的 entryPtr 数量(容量)
    slotsData     []entryPtr // shared by all 256 slots
}

// 位置信息
type entryPtr struct {
    offset   int64  
    hash16   uint16 
    keyLen   uint16 
    _        uint32 
}

// entry header. 存储的 entry = entry header + entry
type entryHdr struct {
    accessTime uint32 // 4B, 访问时间
    expireAt   uint32 // 4B, 到期时间
    keyLen     uint16 // 2B, key长度
    hash16     uint16 // 2B, hashVal其中的16位
    valLen     uint32 // 4B, val长度
    valCap     uint32 // 4B, val容量
    deleted    bool   // 1B, 删除标记
    slotId     uint8  // 1B, 卡槽 id, 获取 entryPtr (值是 0-255)
    _          uint16 // 2B, 内存对齐
}
```

Cache 直接使用了 256 个 segment. 每一个 segment 都拥有一把排他锁.

创建 Cache 的时候可以指定系统使用的内存大小(至少是512KB), 然后将这些内存平均分配给每一个 segment.

每一个 segment 拥有一个 ring buf, 用于存储底层的数据.

每一个 segment 又会产生 256 * N 份卡槽, N是卡槽的容量`[备份数]`, 初始化的时候值为1, 这些卡槽用于存储 "数据 pointer", 
数据格式16B: `offset(8) + hash16(2) + keyLen(2)+ _(4)`. 这里数据的顺序解决了内存偏移量的问题, 可以使
用 unsafe 包进行快速转换.

hash64的用途:

`hash64[-7:0]` => 确定 segment 的位置

`hash64[-15:-8]` => 确定 slot 的位置 (一个位置信息组, 长度是N, 即segment.slotCap, 里面按照 hash16 排序)

`hash64[-31:-16]` => 确定 slot 当中的位置信息(`entryPtr`). 然后设置相应的值

底层 ring buf 存储的时候, 底层的内容是 len(8B) + header(24B) + key + val. 其中 header 内容是前面的`entryHdr`


### 重要的方法解析

- Set, 存储kv

```cgo
func (cache *Cache) Set(key, value []byte, expireSeconds int) (err error) {
    // xxhash.Sum64(), 使用汇编代码实现
    hashVal := hashFunc(key)
	
    // segmentAndOpVal是255, 采用 & 的方式快速定位 key 的 hash 对应的 segment 的index
    segID := hashVal & segmentAndOpVal
	
    // Lock 下完成数据存储
    cache.locks[segID].Lock()
    err = cache.segments[segID].set(key, value, hashVal, expireSeconds)
    cache.locks[segID].Unlock()
    return
}
```

> segment set 方法

```cgo
// 存储数据的头部 24B, 解决了内存对其的问题.
type entryHdr struct {
    accessTime uint32 // 4B, 访问时间
    expireAt   uint32 // 4B, 到期时间
    keyLen     uint16 // 2B, key长度
    hash16     uint16 // 2B, hashVal其中的16位
    valLen     uint32 // 4B, val长度
    valCap     uint32 // 4B, val容量
    deleted    bool   // 1B, 删除标记
    slotId     uint8  // 1B, 卡槽id
    _          uint16 // 2B
}

func (seg *segment) set(key, value []byte, hashVal uint64, expireSeconds int) (err error) {
    // key 的长度有限制, 最多不超过 64K
    if len(key) > 65535 {
        return ErrLargeKey
    }
	
    // ENTRY_HDR_SIZE 是 24
    maxKeyValLen := len(seg.rb.data)/4 - ENTRY_HDR_SIZE
    if len(key)+len(value) > maxKeyValLen {
        // Do not accept large entry.
        return ErrLargeEntry
    }
	
    // 计算 TTL
    now := seg.timer.Now()
    expireAt := uint32(0)
    if expireSeconds > 0 {
        expireAt = now + uint32(expireSeconds)
    }
	
    // 这里目前已经利用了 hashVal 的低 32 位
    // hashVal 的低 0-7 位已经用作计算 segment 的位置
    // hashVal 的低 8-15 位用作计算 slotid 的位置
    // hashVal 的低 16-31 位用作计算 hash16 
    slotId := uint8(hashVal >> 8) 
    hash16 := uint16(hashVal >> 16) 
    
    // ENTRY_HDR_SIZE 是 24, 已经解决内存分布, 高速转换
    var hdrBuf [ENTRY_HDR_SIZE]byte
    hdr := (*entryHdr)(unsafe.Pointer(&hdrBuf[0]))

    // 卡槽的开始位置 = "slotId * slot备份数量" 
    // 卡槽的结束位置 = 卡槽开始位置 + min(当前 slotId 的长度, slot备份数量)
    slot := seg.getSlot(slotId) // 获取卡槽, []entryPtr
	
    // match 表示查询到 key 
    // idx 查询到相应可以存储的位置
    idx, match := seg.lookup(slot, hash16, key) 
    if match {
        matchedPtr := &slot[idx]
        seg.rb.ReadAt(hdrBuf[:], matchedPtr.offset) // 读取 matchedPtr 当中的数据, 这个会导致 hdr 同步修改
		
        originAccessTime := hdr.accessTime
        hdr.accessTime = now
        hdr.expireAt = expireAt
        hdr.keyLen = uint16(len(key)) // 要存储的key的大小
        hdr.hash16 = hash16
        hdr.valLen = uint32(len(value)) // 要存储的value的长度
        hdr.slotId = slotId
		
        // value 没有越界, 直接存入并返回
        if hdr.valCap >= hdr.valLen {
            atomic.AddInt64(&seg.totalTime, int64(hdr.accessTime)-int64(originAccessTime))
            seg.rb.WriteAt(hdrBuf[:], matchedPtr.offset) // 覆盖header
            seg.rb.WriteAt(value, matchedPtr.offset+ENTRY_HDR_SIZE+int64(hdr.keyLen)) // 重写value
            atomic.AddInt64(&seg.overwrites, 1)
            return
        }
		
        // value 越界, 进行 slot 长度扩容
        seg.delEntryPtr(slotId, slot, idx)
        match = false
        // 针对 hdr.valCap 进行扩容, 直到其值大于等于 hdr.valLen
        for hdr.valCap < hdr.valLen {
            hdr.valCap *= 2
        }
        if hdr.valCap > uint32(maxKeyValLen-len(key)) {
            hdr.valCap = uint32(maxKeyValLen - len(key))
        }
    } else {
        hdr.slotId = slotId
        hdr.hash16 = hash16
        hdr.keyLen = uint16(len(key))
        hdr.accessTime = now
        hdr.expireAt = expireAt
        hdr.valLen = uint32(len(value))
        hdr.valCap = uint32(len(value))
        if hdr.valCap == 0 { // avoid infinite loop when increasing capacity.
            hdr.valCap = 1
        }
    }
    
    // 注意: hdr.valCap, 预先分配的大小
    entryLen := ENTRY_HDR_SIZE + int64(len(key)) + int64(hdr.valCap)
	
    // 清除操作, 保证空间足够 entryLen
    slotModified := seg.evacuate(entryLen, slotId, now)
    if slotModified {
        // the slot has been modified during evacuation, we need to looked up for the 'idx' again.
        // otherwise there would be index out of bound error.
        slot = seg.getSlot(slotId)
        idx, match = seg.lookup(slot, hash16, key)
    }
	
    // 获取 ring buf 的 offset 位置
    newOff := seg.rb.End()
	
    // 在 segment 的 slotData 当中写入当前插入的位置信息
    // 注意: 在此期间会发生 slotData 的扩容情况
    seg.insertEntryPtr(slotId, hash16, newOff, idx, hdr.keyLen)
	
    // 底层写入数据
    seg.rb.Write(hdrBuf[:])
    seg.rb.Write(key)
    seg.rb.Write(value)
    seg.rb.Skip(int64(hdr.valCap - hdr.valLen)) // 跳过一定长度, 方便后面追加操作
    atomic.AddInt64(&seg.totalTime, int64(now))
    atomic.AddInt64(&seg.totalCount, 1)
	
    // 减少 vacuumLen 可用长度
    seg.vacuumLen -= entryLen
    return
}

// 插入 entryPtr
func (seg *segment) insertEntryPtr(slotId uint8, hash16 uint16, offset int64, idx int, keyLen uint16) {
    if seg.slotLens[slotId] == seg.slotCap {
        seg.expand() // 等量扩容操作, 将 slotsData 扩展成现在的 2 倍, 同时进行
    }
    seg.slotLens[slotId]++
    atomic.AddInt64(&seg.entryCount, 1)
    slot := seg.getSlot(slotId)
    copy(slot[idx+1:], slot[idx:])
    slot[idx].offset = offset
    slot[idx].hash16 = hash16
    slot[idx].keyLen = keyLen
}

// 获取 slot 的位置
func (seg *segment) lookup(slot []entryPtr, hash16 uint16, key []byte) (idx int, match bool) {
    idx = entryPtrIdx(slot, hash16)
    for idx < len(slot) {
        ptr := &slot[idx]
        if ptr.hash16 != hash16 {
            break
        }
		
        // key 匹配
        match = int(ptr.keyLen) == len(key) && seg.rb.EqualAt(key, ptr.offset+ENTRY_HDR_SIZE)
        if match {
            return
        }
        idx++
    }
    return
}

// 二叉查找
func entryPtrIdx(slot []entryPtr, hash16 uint16) (idx int) {
    high := len(slot)
    for idx < high {
        mid := (idx + high) >> 1
        oldEntry := &slot[mid]
        if oldEntry.hash16 < hash16 {
            idx = mid + 1
        } else {
            high = mid
        }
    }
    return
}

//  seg 清除操作, 不断向前遍历
func (seg *segment) evacuate(entryLen int64, slotId uint8, now uint32) (slotModified bool) {
    var oldHdrBuf [ENTRY_HDR_SIZE]byte
    consecutiveEvacuate := 0 // 回收的次数
    for seg.vacuumLen < entryLen {
        // 最近一次的 offset
        oldOff := seg.rb.End() + seg.vacuumLen - seg.rb.Size()
		
        // 读取 header 的内容
        seg.rb.ReadAt(oldHdrBuf[:], oldOff)
		
        // header
        oldHdr := (*entryHdr)(unsafe.Pointer(&oldHdrBuf[0]))
        oldEntryLen := ENTRY_HDR_SIZE + int64(oldHdr.keyLen) + int64(oldHdr.valCap)
		
        // 已经标记删除, 这里可以删除
        if oldHdr.deleted {
            consecutiveEvacuate = 0 // 回收次数设置为 0
            atomic.AddInt64(&seg.totalTime, -int64(oldHdr.accessTime))
            atomic.AddInt64(&seg.totalCount, -1)
            seg.vacuumLen += oldEntryLen // [1]
            continue
        }
		
        // 过期, 最近最少使用
        expired := oldHdr.expireAt != 0 && oldHdr.expireAt < now
        leastRecentUsed := int64(oldHdr.accessTime)*atomic.LoadInt64(&seg.totalCount) <= atomic.LoadInt64(&seg.totalTime)
        if expired || leastRecentUsed || consecutiveEvacuate > 5 {
            seg.delEntryPtrByOffset(oldHdr.slotId, oldHdr.hash16, oldOff) // 标记删除对应的 k-v
            if oldHdr.slotId == slotId {
                slotModified = true
            }
            consecutiveEvacuate = 0 // 回收次数设置为 0
            atomic.AddInt64(&seg.totalTime, -int64(oldHdr.accessTime))
            atomic.AddInt64(&seg.totalCount, -1)
            seg.vacuumLen += oldEntryLen // [2]
            if expired {
                atomic.AddInt64(&seg.totalExpired, 1)
            } else {
                atomic.AddInt64(&seg.totalEvacuate, 1)
            }
        } else {
            // 删除最近访问过的旧条目[ring buf], 以提高缓存命中率, 这里并没有增加 vacuumLen
            newOff := seg.rb.Evacuate(oldOff, int(oldEntryLen))
            seg.updateEntryPtr(oldHdr.slotId, oldHdr.hash16, oldOff, newOff) // 更新位置信息
            consecutiveEvacuate++ // 回收次数设置为 0
            atomic.AddInt64(&seg.totalEvacuate, 1)
        }
    }
    return
}
```

> segment 的扩容操作 `(slot组大小)`.

```cgo
// 2 * N 扩容
func (seg *segment) expand() {
    newSlotData := make([]entryPtr, seg.slotCap*2*256)
    for i := 0; i < 256; i++ {
        off := int32(i) * seg.slotCap // 旧的 slot 的相对于 seg.slotsData 的偏移量
        copy(newSlotData[off*2:], seg.slotsData[off:off+seg.slotLens[i]]) // 搬移
    }
    seg.slotCap *= 2
    seg.slotsData = newSlotData
}
```


> segment get 方法

```cgo
func (seg *segment) get(key, buf []byte, hashVal uint64, peek bool) (value []byte, expireAt uint32, err error) {
    // 获取 key 的存储位置信息 entryPtr
    slotId := uint8(hashVal >> 8)
    hash16 := uint16(hashVal >> 16)
    slot := seg.getSlot(slotId)
    idx, match := seg.lookup(slot, hash16, key)
    if !match {
        err = ErrNotFound
        if !peek {
            atomic.AddInt64(&seg.missCount, 1)
        }
        return
    }
	
    // 查找到了, 需要更新一些信息
    ptr := &slot[idx]

    var hdrBuf [ENTRY_HDR_SIZE]byte
    seg.rb.ReadAt(hdrBuf[:], ptr.offset)
    hdr := (*entryHdr)(unsafe.Pointer(&hdrBuf[0]))
	
    if !peek {
        // TTL 检测
        now := seg.timer.Now()
        expireAt = hdr.expireAt

        if hdr.expireAt != 0 && hdr.expireAt <= now {
            seg.delEntryPtr(slotId, slot, idx)
            atomic.AddInt64(&seg.totalExpired, 1)
            err = ErrNotFound
            atomic.AddInt64(&seg.missCount, 1)
            return
        }
		
        // 更新 access time
        atomic.AddInt64(&seg.totalTime, int64(now-hdr.accessTime))
        hdr.accessTime = now
        seg.rb.WriteAt(hdrBuf[:], ptr.offset)
    }
	
    // buf 的避免重复分配内存.
    if cap(buf) >= int(hdr.valLen) {
        value = buf[:hdr.valLen]
    } else {
        value = make([]byte, hdr.valLen)
    }
    
    // 读取内容
    seg.rb.ReadAt(value, ptr.offset+ENTRY_HDR_SIZE+int64(hdr.keyLen))
	
    // 更新 counters
    if !peek {
        atomic.AddInt64(&seg.hitCount, 1)
    }
    return
}
```
                                          

> segment del 方法

```cgo
func (seg *segment) del(key []byte, hashVal uint64) (affected bool) {
    slotId := uint8(hashVal >> 8)
    hash16 := uint16(hashVal >> 16)
    slot := seg.getSlot(slotId)
    idx, match := seg.lookup(slot, hash16, key)
    if !match {
        return false
    }
	
    // 删除 entryPtr
    seg.delEntryPtr(slotId, slot, idx)
    return true
}
```

```cgo
func (seg *segment) delEntryPtr(slotId uint8, slot []entryPtr, idx int) {
    offset := slot[idx].offset
    var entryHdrBuf [ENTRY_HDR_SIZE]byte
    seg.rb.ReadAt(entryHdrBuf[:], offset) // 读取 entryPtr
    entryHdr := (*entryHdr)(unsafe.Pointer(&entryHdrBuf[0]))
    entryHdr.deleted = true // 标记删除
    seg.rb.WriteAt(entryHdrBuf[:], offset) // 回写, 这样下次读取可以看到标记删除
    copy(slot[idx:], slot[idx+1:])
    seg.slotLens[slotId]-- // 减少 slotLens
    atomic.AddInt64(&seg.entryCount, -1)
}
```