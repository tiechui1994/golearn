## List 实现

golang 里的 List (在`container/list`包当中) 是一个特殊的双向链表. 

每一个元素会直接指向链表的"哨兵节点". 

数据结构:

```cgo
// 链表元素 
type Element struct {
	// 链表的元素中的下一个和上一个指针.
    // 为了简化实现, 内部将列表 l 实现为环, 这样 &l.root 既是最后一个列表元素( l.Back() )的下一个元素,
    // 也是第一个列表元素( l.Front() ).
	next, prev *Element

	// 该元素所属的链表
	list *List

	// 元素存储的内容
	Value interface{}
}

// List代表一个双向链表.
// List的零值是可以使用的空列表.
type List struct {
	root Element // 链表的哨兵元素, 只有 &root, root.prev, root.next 被使用
	len  int     // 链表的长度, 不包括哨兵元素
}
```

### 链表初始化 

```cgo
func (l *List) Init() *List {
	// 双向链表, 都指向自己
	l.root.next = &l.root
	l.root.prev = &l.root
	
	l.len = 0
	return l
}
```

```cgo
func (l *List) lazyInit() {
	// l.root.next 为 nil, 则说明链表没有初始化, 因为初始化之后, l.root.next 为 &l.root.
	if l.root.next == nil {
		l.Init()
	}
}
```


### 插入数据

```cgo
// 在 at 后面插入 e
/**
before:

       ->      ->
    X      at      Y
       <-      <- 
 
------------------------------
 
after:
       ->      ->      ->
    X      at       e       Y
       <-      <-      <-
**/
func (l *List) insert(e, at *Element) *Element {
	// 保存 at 的下一个元素 Y
	n := at.next
	
	// 调整 at 的 next,  e 的 prev
	at.next = e
	e.prev = at
	
	// 调整 e 的 next, n 的 prev
	e.next = n
	n.prev = e
	
	// 设置 e 所属的列表, 增加列表长度
	e.list = l
	l.len++
	return e
}
```


### 删除数据

```cgo
// 删除元素 e
/**
before:
        
       ->      ->
    X      e       Y
       <-      <- 
 
---------------------

after:
       ->      
    X      Y   
       <-    
**/

func (l *List) remove(e *Element) *Element {
    // 调整 e 前面元素的 next 和 e 后面元素的 prev
	e.prev.next = e.next
	e.next.prev = e.prev
	
	// 调整 e 元素本身的前后指向
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	
	// 移除 e 的 list, 减少元素数量
	e.list = nil
	l.len--
	return e
}
```


### 移动元素

```cgo
// 将元素 e 移动到 at 后面, 然后返回 e
/**
before:
        
       ->      ->            ->      ->
    X      e       Y .... U      at      V
       <-      <-            <-      <-
 
-------------------------------------------

after:
       ->     ->      ->             ->
    X      e      at       Y .... U      V
       <-     <-      <-             <-
**/

// 相当于先删除 e, 然后将 e 添加到 at 后面
func (l *List) move(e, at *Element) *Element {
	if e == at {
		return e
	}
	// 删除元素 e
	e.prev.next = e.next
	e.next.prev = e.prev

    // 添加元素 e 
	n := at.next
	
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e

	return e
}
```

### 合并 list  

```cgo
func (l *List) PushBackList(other *List) {
	l.lazyInit()
	// 初始化的时候, e 是 other 的第一个元素
	// 然后每次将 e 的下一个元素, 插入到 l 的后面
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}
```


### 移动元素位置

```cgo
// 将 e 移动到 mark 的后面
func (l *List) MoveAfter(e, mark *Element) {
	// e.list != l 或者 mark.list != l, 说明 e 或 mark 不属于当前的 list
	// e == mark 说明 e 和 mark 是同一个元素.
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}
```


```cgo
// 将 e 移动到最前面.  <=> 将 e 移动 l.root 的后面
func (l *List) MoveToFront(e *Element) {
	// l.root.next == e, 说明 e 已经是第一个元素了.
	if e.list != l || l.root.next == e {
		return
	}
	l.move(e, &l.root)
}
```