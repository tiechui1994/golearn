package concurrent

//FanIn扇入, 多个合并成一个
import (
	"reflect"
	"sync"
)

// 使用WaitGroup的方式将goroutine(与chan)绑定, 在goroutine当中进队列
func FaninSimple(chs ...<-chan interface{}) <-chan interface{} {
	out := make(chan interface{})

	go func() {
		defer close(out)
		var wg sync.WaitGroup
		wg.Add(len(chs))

		for i := range chs {
			go func() {
				defer wg.Done()
				for c := range chs[i] {
					out <- c
				}
			}()
		}

		wg.Wait()
	}()
	return out
}

// 使用归并的方式
func FaninMerge(chs ...<-chan interface{}) <-chan interface{} {
	switch len(chs) {
	case 0:
		return nil
	case 1:
		return chs[0]
	case 2:
		return mergeTwo(chs[0], chs[1])
	default:
		mid := (len(chs) - 1) / 2
		return mergeTwo(FaninMerge(chs[:mid]...), FaninMerge(chs[mid:]...))
	}
}

func mergeTwo(chanA, chanB <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for chanA != nil || chanB != nil {
			select {
			case val, ok := <-chanA:
				if ok == false {
					chanA = nil
					continue
				}
				out <- val

			case val, ok := <-chanB:
				if ok == false {
					chanB = nil
					continue
				}
				out <- val
			}
		}
	}()
	return out
}

// 构建反射的接收类型的Select, 使用reflect.Select()的方式多选1, 类似select多个case
func FaninReflect(chs ...<-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)

		cases := make([]reflect.SelectCase, len(chs))
		for i := range chs {
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,      // 类型: 发送, 接收, 默认
				Chan: reflect.ValueOf(chs[i]), // 用于接收或者发送的channel
			}
		}

		for len(cases) != 0 {
			i, r, ok := reflect.Select(cases)
			if !ok {
				cases = append(cases[:i], cases[i+1:]...)
			}
			out <- r.Interface()
		}

	}()
	return out
}
