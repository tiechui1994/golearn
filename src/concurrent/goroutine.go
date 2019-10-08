package concurrent

import "fmt"

func DoWork() {
	dowork := func(strings <-chan string) <-chan interface{} {
		complete := make(chan interface{})
		go func() {
			defer fmt.Println("dowork done")
			defer close(complete)
			for s := range strings {
				fmt.Println(s)
			}
		}()
		return complete
	}

	<-dowork(nil)
}
