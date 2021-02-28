package category

func max(args ...int) int {
	max := args[0]
	for i := 1; i < len(args); i++ {
		if max < args[i] {
			max = args[i]
		}
	}
	return max
}

func min(args ...int) int {
	min := args[0]
	for i := 1; i < len(args); i++ {
		if min > args[i] {
			min = args[i]
		}
	}
	return min
}

type Deque struct {
	data []int
}

func (d *Deque) isEmpty() bool {
	return len(d.data) == 0
}

func (d *Deque) peekFirst() int {
	if !d.isEmpty() {
		return d.data[0]
	}
	return 0
}

func (d *Deque) pollFirst() int {
	if !d.isEmpty() {
		ans := d.data[0]
		d.data = d.data[1:]
		return ans
	}
	return 0
}

func (d *Deque) offerFirst(data int) {
	d.data = append(d.data, 0)
	copy(d.data[1:], d.data[0:])
	d.data[0] = data
}

func (d *Deque) peekLast() int {
	if !d.isEmpty() {
		return d.data[len(d.data)-1]
	}
	return 0
}

func (d *Deque) pollLast() int {
	if !d.isEmpty() {
		n := len(d.data)
		ans := d.data[n-1]
		d.data = d.data[:n-1]
		return ans
	}
	return 0
}
func (d *Deque) offerLast(data int) {
	d.data = append(d.data, data)
}
