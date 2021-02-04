## raft 分布式协商协议

- node

// Node 表示raft集群当中的节点.

```cgo
type Node interface {
	Tick() // 时钟的实现, 选举超时和心跳基于此实现
	Campaign(ctx context.Context) error // 参与 leader 竞争
	
	// 在日志中追加数据, 需要实现方保证数据追加的成功
	Propose(ctx context.Context, data []byte) error 
	// ProposeConfChange proposes a configuration change. Like any proposal, the
	// configuration change may be dropped with or without an error being
	// returned. In particular, configuration changes are dropped unless the
	// leader has certainty that there is no prior unapplied configuration
	// change in its log.
	//
	// The method accepts either a pb.ConfChange (deprecated) or pb.ConfChangeV2
	// message. The latter allows arbitrary configuration changes via joint
	// consensus, notably including replacing a voter. Passing a ConfChangeV2
	// message is only allowed if all Nodes participating in the cluster run a
	// version of this library aware of the V2 API. See pb.ConfChangeV2 for
	// usage details and semantics.
	// 集群配置变更
	ProposeConfChange(ctx context.Context, cc pb.ConfChangeI) error 

    // 根据消息变更状态机的状态
	Step(ctx context.Context, msg pb.Message) error

	// 返回当前 point-in-time 状态的 chan. 
	// 节点的用户必须在检索Ready返回状态后调用 Advance.
	// 
	// 注: 直到下一个 Ready 的所有提交的 "entries和snapshots" 都已经完成后, 才能 apply 下一个 Ready 的提交 
	// entries
	// 标志某一状态的完成, 收到状态变化的节点必须提交变更
	Ready() <-chan Ready

	
	// 通知Node该应用程序已保存进度到最近的Ready. 它准备返回下一个可用的Ready.
	//
	// 在 apply 最近 Ready 的 entries 后, 应用程序通常会调用 Advance.
	//
	// 但是, 作为一种优化, 应用程序可以在 apply 命令时调用 Advance. 例如, 当最后一个 Ready 包含快照时, 应用程序
	// 可能需要很长时间才能 apply 快照数据. 要继续接收 Ready 消息而不是阻塞 raft 处理, 可以在完成 apply 最后一个 
	// Ready 消息之前调用 Advance.
    // 	
	// 进行状态的提交, 收到完成标志后, 必须提交过后节点才会实际进行状态机的更新. 在包含快照的场景, 为了避免快照落地
	// 带来的长时间阻塞, 允许继续接受和提交其他状态, 即使之前的快照状态变更并没有完成.
	Advance()
	
	// 将集群配置变更(先传递给 ProposeConfChange)应用于节点. 每当在 Ready.CommittedEntries 中检测到配置更改时,
	// 都必须调用此方法, 除非应用决定拒绝配置更改(即, 将其视为no-op), 在这种状况下, 不要调用此函数.
	ApplyConfChange(cc pb.ConfChangeI) *pb.ConfState

	// 变更leader
	TransferLeadership(ctx context.Context, lead, transferee uint64)

	
	// 保证线性一致性读
	// 请求 read 状态. read 状态将被设置为 ready
	// read 状态具有 read index. 一旦应用程序比 read index更新, 在 read 请求之前的所有线性的 read 请求可以被安全
	// 地处理. read 状态也将将会复制到 rctx 当中.
	ReadIndex(ctx context.Context, rctx []byte) error

	// 状态机当前的状态
	Status() Status
	
	// 上报节点的不可达
	ReportUnreachable(id uint64)
	
	// 上报已发送快照的状态. 
	// id 是用户接受快照的 follower 的 raft ID, status 是 SnapshotFinish 或 SnapshotFailure 使用 
	// SnapshotFinish 调用是没有操作的. 但是, 在 apply 快照当中的任何 failure (eg, while streaming it from 
	// leader to follower) 都应通过 SnapshotFailure 上报给 leader.
	//
	// 当 leader 将快照发送给 follwer 时, 它将暂停所有 raft log probe, 直到 follower 可以 apply 快照并且 
	// advance 其状态. 如果 follower 不能这样做, 例如由于 crash, 它将永远不会从 leader 得到新的更新. 因此, 最重
	// 要的是, 应用程序必须确保捕获快照发送过程中的任何故障并将其上报给 leader. 
	ReportSnapshot(id uint64, status SnapshotStatus)
	
	// 停止节点
	Stop()
}
```
