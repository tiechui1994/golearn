# Golang Cache 比较解析

golang 当中比较缓存分别是 freecache, bigcache, groupcache. 以下主要针对这三个cache进行源码分析.

## lru 的实现

`List + map[interface{}]Element`.

使用 `map[interface{}]Element` 存储缓存的所有kv元素, 使用 List 来调整访问的顺序.

列表最前面的元素是最近访问的, 最后面的元素的不经常访问的元素. 淘汰元素的时候, 从最尾端开始.(lru算法的内容)

K-V 结构:

```cgo
type Key interface{}

// 在 value 当中包含了 key.
type entry struct{
    key Key
    value interface{}
}
```

- 添加元素

```cgo
func (c *Cache) Add(key Key, value interface{}) {
	// 初始化
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	
	// 查找 key 对应的 value, 如果找到, 则将该元素移动到最前面
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	
	// 该元素不存在, 则在最前端新添加一个元素
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	
	// 元素淘汰处理
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}
```

- 元素淘汰

```cgo
// 淘汰元素
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	
	// 从最尾端开始淘汰, 删除
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e) // 从链表移除
	kv := e.Value.(*entry)
	delete(c.cache, kv.key) // 从缓存当中移除
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value) // 淘汰回调
	}
}
```

- 获取元素

```cgo
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	
	// 直接访问 cache, 然后将元素移动到最前面
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}
```


## groupcache

项目: github.com/golang/groupcache

实现的思路:

groupcache 是一个分布式的缓存框架. 


### 初始化注册函数

- func(groupName string) PeerPicker

```cgo
// peer 必须现实 ProtoGetter, 它作用是向其他对端 peer 发送请求获取未在本端缓存的数据
// 缓存同步的重要手段
type ProtoGetter interface {
	Get(ctx context.Context, in *pb.GetRequest, out *pb.GetResponse) error
}

// PickPeer, 根据 key 获取到该 key 存储在哪个 peer
type PeerPicker interface {
	// ok 为 true, 表示已经开始了远程请求
	// ok 为 false, 则表示 key 属于当前的 peer
	PickPeer(key string) (peer ProtoGetter, ok bool)
}

var (
    // 根据 groupName 获取 peer, 后续可以根据 peer 去获取缓存的内容(远程的)
	portPicker func(groupName string) PeerPicker
)

// 注册获取 PeerPicker 函数, 不区分 groupNmae. 单机缓存注册使用
func RegisterPeerPicker(fn func() PeerPicker) {
	if portPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	portPicker = func(_ string) PeerPicker { return fn() }
}

// 注册根据 groupName 获取 PeerPicker 函数, 区分 groupName, 分布式缓存注册使用
func RegisterPerGroupPeerPicker(fn func(groupName string) PeerPicker) {
	if portPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	portPicker = fn
}

// 注意, 上述的两个注册函数最多只能调用一个.

// 根据 groupName 获取 PeerPicker  
func getPeers(groupName string) PeerPicker {
	if portPicker == nil {
		return NoPeers{}
	}
	pk := portPicker(groupName)
	if pk == nil {
		pk = NoPeers{}
	}
	return pk
}
```

- func(*Group) 和 func()

```cgo
var newGroupHook func(*Group)
var initPeerServer     func()

// 注册一个"创建Group"的回调函数, 每次创建Group的时候, 都会调用该函数
// 用于新增修改 Peer 信息. 由于创建 Group 的时候不需要提供 Peer, 因此在此回调函数当中进行修改 Peer
func RegisterNewGroupHook(fn func(*Group)) {
	if newGroupHook != nil {
		panic("RegisterNewGroupHook called more than once")
	}
	newGroupHook = fn
}

// 注册当第一个 Group 创建的时候, 初始化 Server 的回调函数. 该函数只会回调一次.
func RegisterServerStart(fn func()) {
	if initPeerServer != nil {
		panic("RegisterServerStart called more than once")
	}
	initPeerServer = fn
}
```

- Get 逻辑

```cgo

func (g *Group) Get(ctx context.Context, key string, dest Sink) error {
	g.peersOnce.Do(g.initPeers) // 初始化当前的 Group 的 peer
	g.Stats.Gets.Add(1) // 统计信息
	if dest == nil {
		return errors.New("groupcache: nil dest Sink")
	}
	// 从当前 Group 的 mainCache, hotCache 查询
	value, cacheHit := g.lookupCache(key)

	if cacheHit {
		g.Stats.CacheHits.Add(1)
		return setSinkView(dest, value)
	}
    
    
    // 重点:
    // 为避免双重反序列化或复制而进行的优化: 
    // 跟踪目标是否已填充, 一个 caller (如果是 local) 将对此进行设置; 调用则不会.
    // 常见的情况可能是一个 caller.
	destPopulated := false
	value, destPopulated, err := g.load(ctx, key, dest) // 远程加载缓存
	if err != nil {
		return err
	}
	if destPopulated {
		return nil
	}
	return setSinkView(dest, value)
}
```


```cgo
// 通过本地调用getter或将其发送到另一台计算机来加载key
func (g *Group) load(ctx context.Context, key string, dest Sink) (value ByteView, destPopulated bool, err error) {
	g.Stats.Loads.Add(1)
	// loadGroup 使用 flightGroup, 避免并发状况下相同的 key 被多次远程加载 
	viewi, err := g.loadGroup.Do(key, func() (interface{}, error) {
		// 再次检查缓存, 因为singleflight只能删除重复并发的调用. 
		// 2个并发请求可能是 miss Cache, 从而导致2个load()调用.
		// 不幸的是, goroutine 调度会导致此回调连续两次运行. 
		// 如果不再次检查缓存, 即使此键只有一个条目, cache.nbytes也会增加到下面.
		if value, cacheHit := g.lookupCache(key); cacheHit {
			g.Stats.CacheHits.Add(1)
			return value, nil
		}
		
		g.Stats.LoadsDeduped.Add(1)
		var value ByteView
		var err error
		// 获取 key 对应的 Peer, 然后从该 peer 当中去获取Cache
		if peer, ok := g.peers.PickPeer(key); ok {
			value, err = g.getFromPeer(ctx, peer, key)
			if err == nil {
				g.Stats.PeerLoads.Add(1) // Peer 获取成功, 直接返回
				return value, nil
			}
			g.Stats.PeerErrors.Add(1) // Peer 获取失败
		}
		
		// 从本地的 peer 再次获取 key
		value, err = g.getLocally(ctx, key, dest)
		if err != nil {
			g.Stats.LocalLoadErrs.Add(1)
			return nil, err
		}
		g.Stats.LocalLoads.Add(1)
		destPopulated = true // only one caller of load gets this return value
		g.populateCache(key, value, &g.mainCache) // 将当前的缓存填充到主存当中
		return value, nil
	})
	
	if err == nil {
		value = viewi.(ByteView)
	}
	return
}
```

> peer 当中加载缓存

```cgo
func (g *Group) getFromPeer(ctx context.Context, peer ProtoGetter, key string) (ByteView, error) {
	req := &pb.GetRequest{
		Group: &g.name,
		Key:   &key,
	}
	res := &pb.GetResponse{}
	err := peer.Get(ctx, req, res)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: res.Value}
	
	// 使用 res.MinuteQps 或一些灵巧的东西有条件地填充hotCache. 现在, 只需要一定比例的时间即可.
	if rand.Intn(10) == 0 {
		g.populateCache(key, value, &g.hotCache)
	}
	return value, nil
}
```

> local Peer 获取缓存

```cgo
func (g *Group) getLocally(ctx context.Context, key string, dest Sink) (ByteView, error) {
	err := g.getter.Get(ctx, key, dest)
	if err != nil {
		return ByteView{}, err
	}
	return dest.view()
}
```

> 缓存填充与淘汰

```cgo
func (g *Group) populateCache(key string, value ByteView, cache *cache) {
	if g.cacheBytes <= 0 {
		return
	}
	// 在 "cache" 当中填充缓存
	cache.add(key, value)
    
    // 缓存淘汰
	for {
		mainBytes := g.mainCache.bytes()
		hotBytes := g.hotCache.bytes()
		// 空间足够, 没有超过上限
		if mainBytes+hotBytes <= g.cacheBytes {
			return
		}

        // 空间不足, 淘汰策略, 热点缓存超过主存的 1/8, 则淘汰热点缓存, 否则淘汰主存
		victim := &g.mainCache
		if hotBytes > mainBytes/8 {
			victim = &g.hotCache
		}
		victim.removeOldest()
	}
}
```


## cache 的实现

Lock + lru 队列的实现

```cgo
type cache struct {
	mu         sync.RWMutex
	nbytes     int64 // 缓存大小, k-v的总和
	lru        *lru.Cache // lru 存储
	nhit, nget int64 // 缓存命中次数, 缓存读取次数
	nevict     int64 // 缓存淘汰次数
}
```

- 插入 和 读取

```cgo
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
	    // 初始化 lru, 指定淘汰的回调函数
		c.lru = &lru.Cache{
			OnEvicted: func(key lru.Key, value interface{}) {
				val := value.(ByteView)
				c.nbytes -= int64(len(key.(string))) + int64(val.Len())
				c.nevict++
			},
		}
	}
	c.lru.Add(key, value)
	c.nbytes += int64(len(key)) + int64(value.Len())
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nget++
	if c.lru == nil {
		return
	}
	vi, ok := c.lru.Get(key)
	if !ok {
		return
	}
	c.nhit++
	return vi.(ByteView), true
}
```