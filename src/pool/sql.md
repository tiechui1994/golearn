## sql Pool

### 数据结构介绍

`DB` 是一个数据库句柄, 包含了零个或多个基础连接的池. 对于多个goroutine并发使用是安全的.

`sql package` 自动 `create` 和 `release` 连接;它还维护空闲连接的空闲池.

如果数据库具有 `连接状态` 的概念, 则可以在事务(Tx)或连接(Conn)中可靠地观察到这种状态.

调用 `DB.Begin()` 之后, 返回的 `Tx` 将绑定到单个连接. 一旦在事务上调用了 `Commit` 或 `Rollback`, 该事务的连接将返回到
`DB`的空闲连接池.

池大小可以通过 `SetMaxIdleConns` 控制.

```cgo
type DB struct {
	// 仅限原子访问. 放置在首部, 是为了防止在32位平台上出现未对齐问题. 类型为time.Duration.
	waitDuration int64 // 等待新连接的总时间

	connector driver.Connector
	
	// numClosed是一个原子计数器, 表示关闭的连接总数. 
	// 在清除已关闭连接之前(Stmt.css方法中), Stmt.openStmt会对其进行检查.
	numClosed uint64

	mu           sync.Mutex // protects following fields
	freeConn     []*driverConn
	connRequests map[uint64]chan connRequest
	nextRequest  uint64 // 用于 connRequests 当中
	numOpen      int    // 已经连接的和正在连接的总数 
	
	// Used to signal the need for new connections
	// a goroutine running connectionOpener() reads on this chan and
	// maybeOpenNewConnections sends on the chan (one send per needed connection)
	// It is closed during db.Close(). The close tells the connectionOpener
	// goroutine to exit.
	openerCh          chan struct{}
	resetterCh        chan *driverConn
	closed            bool
	dep               map[finalCloser]depSet
	lastPut           map[*driverConn]string // 最终最新put的连接, debug only
	maxIdle           int                    // zero means defaultMaxIdleConns; negative means 0
	maxOpen           int                    // <= 0 means unlimited
	maxLifetime       time.Duration          // connection 最大重用的时间
	cleanerCh         chan struct{}          // 清理信号
	
	waitCount         int64 // 处于等待中的connections的数量
	maxIdleClosed     int64 // 由于idle而关闭的connections的数量
	maxLifetimeClosed int64 // 由于最大重用时间限制而关闭connections的数量

	stop func() // stop cancels the connection opener and the session resetter.
}
```


`driverConn` 使用 `mutex` 包装 `driver.Conn`, 目的是为了 `hold` 住所有对 `Conn` 的 `calls` (
包括对通过该 `Conn` 返回的接口的任何调用, 例如Tx, Stmt, Result, Row的调用)

```cgo
type driverConn struct {
	db        *DB
	createdAt time.Time

	sync.Mutex  // guards following
	ci          driver.Conn
	closed      bool
	finalClosed bool // ci.Close has been called
	openStmt    map[*driverStmt]bool
	lastErr     error // lastError captures the result of the session resetter.

	// guarded by db.mu
	inUse      bool
	onPut      []func() // code (with db.mu held) run when conn is next returned
	dbmuClosed bool     // same as closed, but guarded by db.mu, for removeClosedStmtLocked
}
```


## method

```cgo
// 运行在单独的goroutine当中, 用于创建新的connections
func (db *DB) connectionOpener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-db.openerCh:
			db.openNewConnection(ctx)
		}
	}
}

// 运行在单独的goroutine当中, rest connections(异步方式)
func (db *DB) connectionResetter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(db.resetterCh)
			for dc := range db.resetterCh {
				dc.Unlock()
			}
			return
		case dc := <-db.resetterCh:
			dc.resetSession(ctx)
		}
	}
}
```



```cgo
// 创建新的 connection 
func (db *DB) openNewConnection(ctx context.Context) {
	// maybeOpenNewConnctions has already executed db.numOpen++ before it sent
	// on db.openerCh. This function must execute db.numOpen-- if the
	// connection fails or is closed before returning.
	ci, err := db.connector.Connect(ctx)
	db.mu.Lock()
	defer db.mu.Unlock()
	
	// db 已关闭
	if db.closed {
		if err == nil {
			ci.Close()
		}
		db.numOpen-- // opened and opening 的数量
		return
	}
	
	// open failed
	if err != nil {
		db.numOpen--
		db.putConnDBLocked(nil, err) // 释放 connRequest 
		db.maybeOpenNewConnections() // 重新打开的操作
		return
	}
	
	// wrap
	dc := &driverConn{
		db:        db,
		createdAt: nowFunc(),
		ci:        ci,
	}
	
	// err为nil
	if db.putConnDBLocked(dc, err) {
		db.addDepLocked(dc, dc)
	} else {
		db.numOpen--
		ci.Close()
	}
}
```


```cgo
// 假设db.mu已锁定.
// 如果有 connRequests 并且尚未达到连接限制, 则告诉 `connectionOpener` 打开新的连接.
func (db *DB) maybeOpenNewConnections() {
	numRequests := len(db.connRequests) // 需要的连接数量
	if db.maxOpen > 0 {
		numCanOpen := db.maxOpen - db.numOpen
		if numRequests > numCanOpen {
			numRequests = numCanOpen
		}
	}
	
	// 打开 numRequests 个连接
	for numRequests > 0 {
		db.numOpen++ // optimistically
		numRequests--
		if db.closed {
			return
		}
		db.openerCh <- struct{}{} // 打开新的连接
	}
}
```


```cgo
// true 表示当前的 driverConn 可以给 connRequest 或 可以进入 freeConn.
// flase 表示当前的 driverConn 没有用, 需要释放.
// 
// 逻辑如下: 
// 如果存在一个 connRequest, 则putConnDBLocked将满足connRequest; 
// 如果err == nil并且不超过空闲连接限制, 它将把 driverConn返回到freeConn列表.
// 
// err 与 dc 的条件限定:
// 如果err != nil, 则忽略dc的值.
// 如果err == nil, 则dc不能等于nil.
func (db *DB) putConnDBLocked(dc *driverConn, err error) bool {
	// 关闭
	if db.closed {
		return false
	}
	// 超出配置的连接的限制
	if db.maxOpen > 0 && db.numOpen > db.maxOpen {
		return false
	}
	
	if c := len(db.connRequests); c > 0 {
		// 存在连接请求, 连接重用
		var req chan connRequest
		var reqKey uint64
		for reqKey, req = range db.connRequests {
			break
		}
		delete(db.connRequests, reqKey) // Remove from pending requests.
		if err == nil {
			dc.inUse = true 
		}
		req <- connRequest{
			conn: dc,
			err:  err,
		}
		return true
	} else if err == nil && !db.closed {
	    // 不存在连接请求, 连接空闲
	    
	    // db.maxIdleConnsLocked() 配置最大是空闲数量
	    // db.freeConn, 空闲连接
		if db.maxIdleConnsLocked() > len(db.freeConn) {
			db.freeConn = append(db.freeConn, dc)
			db.startCleanerLocked() // 开启清理模式
			return true
		}
		
		// 超出空闲上限, 只能关闭了
		db.maxIdleClosed++
	}
	
	return false
}
```


```cgo
// db.mu.Lock() 下开启清理模式:
func (db *DB) startCleanerLocked() {
    // 条件: 存在最大reuse时间, 存在最大连接限制, 并且当前的 cleanerCh 为 nil
	if db.maxLifetime > 0 && db.numOpen > 0 && db.cleanerCh == nil {
		db.cleanerCh = make(chan struct{}, 1)
		go db.connectionCleaner(db.maxLifetime)
	}
}

func (db *DB) connectionCleaner(d time.Duration) {
	const minInterval = time.Second

	if d < minInterval {
		d = minInterval
	}
	t := time.NewTimer(d)

	for {
		select {
		case <-t.C:
		case <-db.cleanerCh: // maxLifetime was changed or db was closed.
		}

		db.mu.Lock()
		d = db.maxLifetime
		if db.closed || db.numOpen == 0 || d <= 0 {
			db.cleanerCh = nil
			db.mu.Unlock()
			return
		}

		expiredSince := nowFunc().Add(-d)
		var closing []*driverConn
		for i := 0; i < len(db.freeConn); i++ {
			c := db.freeConn[i]
			if c.createdAt.Before(expiredSince) {
				closing = append(closing, c)
				last := len(db.freeConn) - 1
				db.freeConn[i] = db.freeConn[last]
				db.freeConn[last] = nil
				db.freeConn = db.freeConn[:last]
				i--
			}
		}
		db.maxLifetimeClosed += int64(len(closing))
		db.mu.Unlock()

		for _, c := range closing {
			c.Close()
		}

		if d < minInterval {
			d = minInterval
		}
		t.Reset(d)
	}
}
```
