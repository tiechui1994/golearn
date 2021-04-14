package redis

type Entry struct {
	hash  uint64
	key   string
	value interface{}
	next  *Entry

	before *Entry
	after  *Entry
}

type LinkedHashMap struct {
	head       *Entry
	hash       func(interface{}) uint64
	table      []*Entry
	size       int
	maxSize    int
	bucketSize uint64
	modCount   uint64
}

// redis lru cache
func (h *LinkedHashMap) set(key string, value interface{}) interface{} {
	hash := h.hash(key)
	i := hash & (h.bucketSize - 1)

	// search
	for e := h.table[i]; e != nil; e = e.next {
		if e.hash == hash && e.key == key {
			old := e.value
			e.value = value
			h.modCount += 1

			// update order(remove,add)
			e.before.after = e.after
			e.after.before = e.before

			after := h.head
			before := h.head.before
			before.after = e
			after.before = e

			return old
		}
	}

	// insert
	e := Entry{hash: hash, key: key, value: value, next: h.table[i]}
	h.table[i] = &e

	if h.head == nil {
		e := &Entry{}
		e.before = e
		e.after = e
		h.head = e
	}

	// update order(add)
	after := h.head
	before := h.head.before
	before.after = &e
	after.before = &e

	h.size++

	// overflow
	eldst := h.head.after
	if h.size > h.maxSize {
		// update order(remove)
		eldst.before.after = eldst.after
		eldst.after.before = eldst.before

		// delete node
		idx := e.hash & (h.bucketSize - 1)
		if eldst == h.table[idx] {
			h.table[idx] = nil
		} else {
			for e := h.table[idx]; e != nil; e = e.next {
				if e.next == eldst {
					e.next = eldst.next
				}
			}
		}
	}

	return nil
}

func (h *LinkedHashMap) get(key string) interface{} {
	hash := h.hash(key)
	i := hash & (h.bucketSize - 1)

	for e := h.table[i]; e != nil; e = e.next {
		if e.hash == hash && e.key == key {
			value := e.value
			h.modCount += 1

			// 删除e
			e.before.after = e.after
			e.after.before = e.before

			// 添加e
			after := h.head
			before := h.head.before
			before.after = e
			after.before = e

			return value
		}
	}

	return nil
}
