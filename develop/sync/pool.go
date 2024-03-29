package main

import (
	"sync/atomic"
	"unsafe"
)

type CasQueue struct {
	head *node
	tail *node
}

type node struct {
	value interface{}
	next *node
}

func (q *CasQueue) EnQueue(val interface{})   {
	node := unsafe.Pointer(&node{value: val})
	for {
		tail := q.tail
		if tail == q.tail {
			// q.tail.next 为空, 可添加
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail.next)), nil, node) {
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(tail), node)
				return
			} else  {
				// q.tail.next 非空.  需要修复
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)), unsafe.Pointer(tail), unsafe.Pointer(tail.next))
			}
		}
	}
}


func (q *CasQueue) DeQueue()  interface{} {
	for {
		head := q.head
		tail := q.tail
		first := q.head.next // 首个节点
		if head == q.head {
			if head == tail {
				// queue empty
				if first == nil {
					return nil
				}
				// 只有一个节点, 则都(head, tail)需指向该节点
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.tail)),
					unsafe.Pointer(tail), unsafe.Pointer(first))
			} else {
				// 可以获取
				value := first.value
				if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.head)),
					unsafe.Pointer(head), unsafe.Pointer(first)) {
					return value
				}
			}
		}
	}
}