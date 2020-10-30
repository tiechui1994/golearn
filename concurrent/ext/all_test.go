package ext

import (
	"fmt"
	"testing"
	"time"
)

func TestSpinLock(t *testing.T) {
	var c int
	for k := 0; k < 100; k++ {
		num := 100
		l := SpinLock{}
		go func() {
			for i := 0; i < 10; i++ {
				go func() {
					l.Lock()
					num = num - 1
					l.Unlock()
				}()
			}
		}()
		go func() {
			for i := 0; i < 10; i++ {
				go func() {
					l.Lock()
					num = num - 1
					l.Unlock()
				}()
			}
		}()

		time.Sleep(1 * time.Millisecond)
		if num != 80 {
			c++
		}
	}

	fmt.Println(c)
}
