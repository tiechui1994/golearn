package concurrent

import (
	"reflect"
	"sync"
)

// 核心: 只要有一个完成了, 就全部完成了

// 递归版本 channel 做or运算, 只要其中一个完成就返回
func Or(chs ...<-chan interface{}) <-chan interface{} {
	switch len(chs) {
	case 0:
		return nil
	case 1:
		return chs[0]
	}

	ordone := make(chan interface{})
	go func() {
		defer close(ordone)
		switch len(chs) {
		case 2:
			select {
			case <-chs[0]:
			case <-chs[1]:
			}
		default:
			m := (len(chs) - 1) / 2
			select {
			case <-Or(chs[:m]...):
			case <-Or(chs[m+1:]...):
			}
		}
	}()

	return ordone
}

func OrReflect(chs ...<-chan interface{}) <-chan interface{} {
	switch len(chs) {
	case 0:
		return nil
	case 1:
		return chs[0]
	}

	ordone := make(chan interface{})
	go func() {
		defer close(ordone)

		cases := make([]reflect.SelectCase, len(chs))
		for i := range chs {
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(chs[i]),
			}
		}

		reflect.Select(cases)
	}()

	return ordone
}

// 利用了once只能执行一次的特点, 一次执行, 其他全部的goroutine全部返回
func OrOnce(chs ...<-chan interface{}) <-chan interface{} {
	ordone := make(chan interface{})

	go func() {
		var once sync.Once

		for _, ch := range chs {

			go func(ch <-chan interface{}) {
				select {
				case <-ch:
					once.Do(func() {
						close(ordone)
					})
				case <-ordone:
				}
			}(ch)

		}

	}()

	return ordone
}

func OrMege(chs ...<-chan interface{}) <-chan interface{} {
	switch len(chs) {
	case 0:
		return nil
	case 1:
		return chs[0]
	}

	ordone := make(chan interface{})

	go func() {
		defer close(ordone)

		switch len(ordone) {
		case 2:
			select {
			case <-chs[0]:
			case <-chs[1]:
			}
		default:
			select {
			case <-chs[0]:
			case <-chs[1]:
			case <-chs[2]:
			case <-OrMege(append(chs[3:], ordone)...):
			}
		}

	}()

	return ordone
}
