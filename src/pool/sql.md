## sql Pool

### 数据结构介绍

- DB

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
	lastPut           map[*driverConn]string // debug 日志
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

- driverConn

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

- connectionOpener, connectionResetter (job)

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

- openNewConnection

create new connection

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
		db.addDepLocked(dc, dc)  // 新创建的依赖会添加依赖
	} else {
		db.numOpen--
		ci.Close()
	}
}
```

> **调用 maybeOpenNewConnections 的地方**

> 1.DB.openNewConnection() 连接创建失败了, 当然需要重新创建了

> 2.DB.conn() 连接创建失败了

> 3.DB.putConn() 连接放回到freeConn当中, 但是连接最新错误是ErrBadConn

> 4.driverConn.finalClose() 当前的连接已经关闭, 连接可能缺少了


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
		db.openerCh <- struct{}{} // 该函数的调用导致补充新的连接, 异步调用了 openNewConnection()
	}
}
```

- putConnDBLocked 

put driverConn to connRequest or idle pool.

> **调用 putConnDBLocked 的地方** 

> 1.DB.putConn() 将连接放入 freeConn 的时候;

> 2.DB.openNewConnection() 的时候, 接收到创建的信号, 创建了新的连接, 当然需要给需要的地方

```cgo
// 在 db.mu.Lock 的状况下调用
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

- putConn

put driverConn to freeConn pool
 
dc: *driverConn
err: current error
resetSession: reset driverConn Session

> **调用 putConn 的地方** 

> 1.DB.conn() 取消的时候.
 
> 2.driverConn 释放连接的时候 finalColse()

```cgo
// adds a connection to the db's free pool.
func (db *DB) putConn(dc *driverConn, err error, resetSession bool) {
	db.mu.Lock()
	
	// 当前的 dc.inUse 为 true
	if !dc.inUse {
		if debugGetPut {
			fmt.Printf("putConn(%v) DUPLICATE was: %s\n\nPREVIOUS was: %s", dc, stack(), db.lastPut[dc])
		}
		panic("sql: connection returned that was never out")
	}
	if debugGetPut {
		db.lastPut[dc] = stack()
	}
	dc.inUse = false

  // callback
	for _, fn := range dc.onPut {
		fn()
	}
	dc.onPut = nil

	if err == driver.ErrBadConn {
		// Don't reuse bad connections.
		// Since the conn is considered bad and is being discarded, treat it as closed. 
		// Don't decrement the open count here, finalClose will take care of that.
		db.maybeOpenNewConnections()
		db.mu.Unlock()
		dc.Close()
		return
	}
	
	// Hook
	if putConnHook != nil {
		putConnHook(db, dc)
	}
	 
	if db.closed {
		// Connections do not need to be reset if they will be closed.
		// Prevents writing to resetterCh after the DB has closed.
		resetSession = false
	}
	if resetSession {
		if _, resetSession = dc.ci.(driver.SessionResetter); resetSession {
			// Lock the driverConn here so it isn't released until the connection is reset.
			// The lock must be taken before the connection is put into the pool to prevent it from 
			// being taken out before it is reset.
			dc.Lock()
		}
	}
	added := db.putConnDBLocked(dc, nil)
	db.mu.Unlock()

  // need to release driverConn
	if !added {
		if resetSession {
			dc.Unlock()
		}
		dc.Close()
		return
	}
	
	// driverConn 被重用 或者 放到了 freeConn 当中
	if !resetSession {
		return
	}
	
	// 发送信号重置 driverConn 的 Session
	select {
	default:
		// If the resetterCh is blocking then mark the connection as bad and continue on.
		dc.lastErr = driver.ErrBadConn
		dc.Unlock()
	case db.resetterCh <- dc: // 直接导致 resetSession 调用
	}
}
```


- startCleanerLocked

```cgo
// db.mu.Lock() 下开启清理模式:
func (db *DB) startCleanerLocked() {
    // 条件: 存在最大reuse时间, 存在最大连接限制, 并且当前的 cleanerCh 为 nil
	if db.maxLifetime > 0 && db.numOpen > 0 && db.cleanerCh == nil {
		db.cleanerCh = make(chan struct{}, 1)
		go db.connectionCleaner(db.maxLifetime)
	}
}

// 连接清理工作Job, d 是清理间隔
func (db *DB) connectionCleaner(d time.Duration) {
	const minInterval = time.Second
  
  // 清理的时间最小是1s
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
		d = db.maxLifetime // 这里 d 发生了改变, 也就是说maxLifetime可以在1s以下
		// db closed, or connections no limit, or maxLifetime lte 0
		if db.closed || db.numOpen == 0 || d <= 0 {
			db.cleanerCh = nil
			db.mu.Unlock()
			return
		}
    
		expiredSince := nowFunc().Add(-d) // 已经过期的时间点
		var closing []*driverConn
		// 从 freeConn 当中进行清理
		for i := 0; i < len(db.freeConn); i++ {
			c := db.freeConn[i]
			if c.createdAt.Before(expiredSince) {
				closing = append(closing, c) // 当前 i 加入到closeing队列
				last := len(db.freeConn) - 1 
				db.freeConn[i] = db.freeConn[last] // 最后一位复制到当前的位置
				db.freeConn[last] = nil 
				db.freeConn = db.freeConn[:last] // 更新freConn数组
				i--
			}
		}
		db.maxLifetimeClosed += int64(len(closing))
		db.mu.Unlock()

    // 清理
		for _, c := range closing {
			c.Close()
		}
    
    // 重新设置清理周期, 最少1s
		if d < minInterval {
			d = minInterval
		}
		t.Reset(d)
	}
}
```

- conn

获取数据库连接. newly or cached *driverConn

ctx: context.Context
strategy: 更新策略, alwaysNewConn(0), cachedOrNewConn(1)

```cgo
func (db *DB) conn(ctx context.Context, strategy connReuseStrategy) (*driverConn, error) {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil, errDBClosed
	}
	// Check if the context is expired.
	select {
	default:
	case <-ctx.Done():
		db.mu.Unlock()
		return nil, ctx.Err()
	}
	lifetime := db.maxLifetime
	
	//  使用Cached
	numFree := len(db.freeConn) // free数量
	if strategy == cachedOrNewConn && numFree > 0 {
		conn := db.freeConn[0]
		copy(db.freeConn, db.freeConn[1:]) // 更新freeConn
		db.freeConn = db.freeConn[:numFree-1]
		conn.inUse = true
		db.mu.Unlock()
		// 过期
		if conn.expired(lifetime) {
			conn.Close()
			return nil, driver.ErrBadConn
		}
		
		// 锁定当前的conn, 读取lastErr(最新的error), 确定是否重置该 conn
		conn.Lock()
		err := conn.lastErr
		conn.Unlock()
		if err == driver.ErrBadConn {
			conn.Close()
			return nil, driver.ErrBadConn
		}
		return conn, nil
	}
  
  // 限制条件判别, 当前已经达到最大连接数量. 需要connRequest进行等待, 可能是创建, 也可能是缓存当中获取
	if db.maxOpen > 0 && db.numOpen >= db.maxOpen {
		// Make the connRequest channel. It's buffered so that the
		// connectionOpener doesn't block while waiting for the req to be read.
		req := make(chan connRequest, 1)
		reqKey := db.nextRequestKeyLocked() // 获取reqKey, Lock 情况下获取
		db.connRequests[reqKey] = req
		db.mu.Unlock()

		// Timeout the connection request with the context.
		select {
		case <-ctx.Done():
			// Remove the connection request and ensure no value has been sent
			// on it after removing.
			db.mu.Lock()
			delete(db.connRequests, reqKey)
			db.mu.Unlock()
			
			// 请求已经获取到了 conn, 但是此时已经取消了任务
			select {
			default:
			case ret, ok := <-req:
				if ok && ret.conn != nil {
					db.putConn(ret.conn, ret.err, false) // 任务取消, 需要释放连接到 free pool
				}
			}
			return nil, ctx.Err()
			
		case ret, ok := <-req:
			if !ok {
				return nil, errDBClosed
			}
			
			// 过期了
			if ret.err == nil && ret.conn.expired(lifetime) {
				ret.conn.Close()
				return nil, driver.ErrBadConn
			}
			
			// 不存在的conn
			if ret.conn == nil {
				return nil, ret.err
			}
			
			// 读取 lastErr, 进行判别
			ret.conn.Lock()
			err := ret.conn.lastErr
			ret.conn.Unlock()
			if err == driver.ErrBadConn {
				ret.conn.Close()
				return nil, driver.ErrBadConn
			}
			
			return ret.conn, ret.err
		}
	}
 
  // 创建新的请求
	db.numOpen++ // optimistically
	db.mu.Unlock()
	ci, err := db.connector.Connect(ctx)
	if err != nil {
		db.mu.Lock()
		db.numOpen-- // correct for earlier optimism
		db.maybeOpenNewConnections()
		db.mu.Unlock()
		return nil, err
	}
	
	// 获取到了
	db.mu.Lock()
	dc := &driverConn{
		db:        db,
		createdAt: nowFunc(),
		ci:        ci,
		inUse:     true,
	}
	db.addDepLocked(dc, dc) // 新创建的连接, 会添加依赖
	db.mu.Unlock()
	return dc, nil
}
```


- Close

关闭数据库, 并 "阻止启动新查询".

关闭, 然后等待在服务器上已经开始处理的所有查询完成.

```cgo
// 关闭数据库, 并 "阻止启动新查询".
// 关闭, 然后等待在服务器上已经开始处理的所有查询完成.
//
// 关闭数据库很少, 因为该数据库句柄是长期存在的并且在许多goroutine之间共享.
func (db *DB) Close() error {
	db.mu.Lock()
	if db.closed { // Make DB.Close idempotent
		db.mu.Unlock()
		return nil
	}
	
	// 关闭清理任务
	if db.cleanerCh != nil {
		close(db.cleanerCh)
	}
	
	// 清理freeConn列表(添加回调函数), closed状态, 关闭所有的ConnRequests(申请连接请求)
	var err error
	fns := make([]func() error, 0, len(db.freeConn))
	for _, dc := range db.freeConn {
		fns = append(fns, dc.closeDBLocked())
	}
	db.freeConn = nil
	db.closed = true
	for _, req := range db.connRequests {
		close(req)
	}
	db.mu.Unlock()
	
	// 对于 freeConn 的每一个 driverConn 执行`解绑`操作. 减少引用的操作
	for _, fn := range fns {
		err1 := fn()
		if err1 != nil {
			err = err1
		}
	}
	
	// db 执行 context 的 cancel() 的退出函数
	db.stop()
	return err
}
```


```cgo
func (dc *driverConn) closeDBLocked() func() error {
	dc.Lock()
	defer dc.Unlock()
	if dc.closed {
		return func() error { return errors.New("sql: duplicate driverConn close") }
	}
	dc.closed = true
	return dc.db.removeDepLocked(dc, dc)
}
```

- Dep 相关的函数


```cgo
// 依赖关联计数
type finalCloser interface {
	// 当所有的引用object的数量变为0, 会调用此函数. 当调用此函数的时候, (*DB).mu 不加锁
	finalClose() error
}
```

```cgo
// addDep, 将 x 和 dep 进行依赖绑定. 一般都是 `self bind self`
func (db *DB) addDep(x finalCloser, dep interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.addDepLocked(x, dep)
}

func (db *DB) addDepLocked(x finalCloser, dep interface{}) {
	if db.dep == nil {
		db.dep = make(map[finalCloser]depSet)
	}
	xdep := db.dep[x]
	if xdep == nil {
		xdep = make(depSet)
		db.dep[x] = xdep
	}
	xdep[dep] = true
}

// removeDep notes that x no longer depends on dep.
// 如果 x 有 dependencies, 返回 nil
// 如果 x 没有任何的 dependencies, x.finalClose() 会被调用, 返回调用的结果
func (db *DB) removeDep(x finalCloser, dep interface{}) error {
	db.mu.Lock()
	fn := db.removeDepLocked(x, dep)
	db.mu.Unlock()
	return fn()
}

func (db *DB) removeDepLocked(x finalCloser, dep interface{}) func() error {
	xdep, ok := db.dep[x]
	if !ok {
		panic(fmt.Sprintf("unpaired removeDep: no deps for %T", x))
	}

	l0 := len(xdep)
	delete(xdep, dep)

	switch len(xdep) {
	case l0:
		// Nothing removed. Shouldn't happen.
		panic(fmt.Sprintf("unpaired removeDep: no %T dep on %T", dep, x))
	case 0:
		// No more dependencies.
		delete(db.dep, x)
		return x.finalClose
	default:
		// Dependencies remain.
		return func() error { return nil }
	}
}
```


- driverConn 相关的方法

```cgo
func (dc *driverConn) releaseConn(err error) {
	dc.db.putConn(dc, err, true) // 将 dc 放入到 free pool当中
}

// the dc.db's Mutex is held.
func (dc *driverConn) closeDBLocked() func() error {
	dc.Lock()
	defer dc.Unlock()
	if dc.closed {
		return func() error { return errors.New("sql: duplicate driverConn close") }
	}
	dc.closed = true
	return dc.db.removeDepLocked(dc, dc) // 移除依赖
}

// Close 
func (dc *driverConn) Close() error {
	dc.Lock()
	if dc.closed {
		dc.Unlock()
		return errors.New("sql: duplicate driverConn close")
	}
	dc.closed = true
	dc.Unlock() // not defer; removeDep finalClose calls may need to lock

	// And now updates that require holding dc.mu.Lock.
	dc.db.mu.Lock()
	dc.dbmuClosed = true
	fn := dc.db.removeDepLocked(dc, dc) // 移除依赖
	dc.db.mu.Unlock()
	return fn() // 可能是执行 finalClose() 函数
}

// 当 driverConn 解除依赖, Closed 的函数
func (dc *driverConn) finalClose() error {
	var err error

	// Each *driverStmt has a lock to the dc. Copy the list out of the dc
	// before calling close on each stmt.
	var openStmt []*driverStmt
	withLock(dc, func() {
		openStmt = make([]*driverStmt, 0, len(dc.openStmt))
		for ds := range dc.openStmt {
			openStmt = append(openStmt, ds)
		}
		dc.openStmt = nil
	})
	for _, ds := range openStmt {
		ds.Close()
	}
	withLock(dc, func() {
		dc.finalClosed = true
		err = dc.ci.Close()
		dc.ci = nil
	})

	dc.db.mu.Lock()
	dc.db.numOpen--
	dc.db.maybeOpenNewConnections()
	dc.db.mu.Unlock()

	atomic.AddUint64(&dc.db.numClosed, 1)
	return err
}
```


- Query, Ping, Exec, Tx

XXX -> XXXContext -> xxx -> xxxDC

XXXContext, 使用 Context 进行包装, 可以随时取消

xxx, 获取 driverConn 连接

xxxDC, 执行具体的操作, 会释放连接(error)或者由返回的结果释放连接

```cgo
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	var rows *Rows
	var err error
	for i := 0; i < maxBadConnRetries; i++ {
		rows, err = db.query(ctx, query, args, cachedOrNewConn)
		if err != driver.ErrBadConn {
			break
		}
	}
	if err == driver.ErrBadConn {
		return db.query(ctx, query, args, alwaysNewConn)
	}
	return rows, err
}

func (db *DB) query(ctx context.Context, query string, args []interface{}, strategy connReuseStrategy) (*Rows, error) {
	dc, err := db.conn(ctx, strategy)
	if err != nil {
		return nil, err
	}

	return db.queryDC(ctx, nil, dc, dc.releaseConn, query, args)
}

func (db *DB) queryDC(ctx, txctx context.Context, dc *driverConn, releaseConn func(error), query string, args []interface{}) (*Rows, error) {
	queryerCtx, ok := dc.ci.(driver.QueryerContext)
	var queryer driver.Queryer
	if !ok {
		queryer, ok = dc.ci.(driver.Queryer)
	}
	if ok {
		var nvdargs []driver.NamedValue
		var rowsi driver.Rows
		var err error
		withLock(dc, func() {
			nvdargs, err = driverArgsConnLocked(dc.ci, nil, args)
			if err != nil {
				return
			}
			rowsi, err = ctxDriverQuery(ctx, queryerCtx, queryer, query, nvdargs)
		})
		if err != driver.ErrSkip {
			if err != nil {
				releaseConn(err) // release 1
				return nil, err
			}
			// Note: ownership of dc passes to the *Rows, to be freed
			// with releaseConn.
			rows := &Rows{
				dc:          dc,
				releaseConn: releaseConn, // release 2
				rowsi:       rowsi,
			}
			rows.initContextClose(ctx, txctx)
			return rows, nil
		}
	}

	var si driver.Stmt
	var err error
	withLock(dc, func() {
		si, err = ctxDriverPrepare(ctx, dc.ci, query)
	})
	if err != nil {
		releaseConn(err) // release 3
		return nil, err
	}

	ds := &driverStmt{Locker: dc, si: si}
	rowsi, err := rowsiFromStatement(ctx, dc.ci, ds, args...)
	if err != nil {
		ds.Close()
		releaseConn(err) // release 4
		return nil, err
	}

	// Note: ownership of ci passes to the *Rows, to be freed
	// with releaseConn.
	rows := &Rows{
		dc:          dc,
		releaseConn: releaseConn, // release 5
		rowsi:       rowsi,
		closeStmt:   ds,
	}
	rows.initContextClose(ctx, txctx)
	return rows, nil
}
```