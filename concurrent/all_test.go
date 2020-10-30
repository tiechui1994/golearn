package concurrent

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestDeadlock(t *testing.T) {
	Deadlock()
}

func TestLivelock(t *testing.T) {
	Livelock()
}

func TestHungryLock(t *testing.T) {
	HungryLock()
}

func TestDoWork(t *testing.T) {
	DoWork()
}

func TestBridge(t *testing.T) {
	getvals := func() <-chan <-chan interface{} {
		chans := make(chan (<-chan interface{}))
		go func() {
			defer close(chans)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 2)
				stream <- i
				close(stream)
				chans <- stream
			}
		}()
		return chans
	}

	for val := range Bridge(nil, getvals()) {
		fmt.Printf("%v ", val)
	}

	val := make(chan int, 2)
	val <- 100
	val <- 200
	close(val)
	fmt.Println(<-val, <-val)
}

func TestOrDone(t *testing.T) {
	stream := make(chan interface{}, 4)
	done := make(chan interface{})
	go func() {
		var count int
		for {
			stream <- count + 1
			count = count + 1
			if count == 12 {
				close(done)
				close(stream)
				return
			}
		}
	}()
	for range OrDone(done, stream) {
	}
}

func TestHeartInterval(t *testing.T) {
	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() {
		close(done)
	})

	const timeout = 2 * time.Second

	heartbeat, results := HeartInterval(done, timeout/2)
	for {
		select {
		case _, ok := <-heartbeat:
			if ok == false {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if ok == false {
				return
			}
			fmt.Printf("results %v\n", r.Second())
		case <-time.After(timeout):
			return
		}
	}
}

func TestHeartWork(b *testing.T) {
	done := make(chan interface{})
	defer close(done)

	vals := []int{0, 1, 2, 3, 5}
	_, result := HeartWork(done, vals...)
	for i, expect := range vals {
		select {
		case r := <-result:
			if r != expect {
				b.Errorf("index %v : expect %v, but received %v,", i, expect, r)
			}
		case <-time.After(time.Second):
			b.Fatalf("test timed out")
		}
	}
}

func TestHeartWork2(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	vals := []int{0, 1, 2, 3, 5}
	heartbeat, result := HeartWork(done, vals...)
	<-heartbeat

	i := 0
	for r := range result {
		if r != vals[i] {
			t.Errorf("index %v : expect %v, but received %v,", i, vals[i], r)
		}
		i++
	}
}

func TestSecurtyHeartInterval(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	nums := []int{0, 1, 2, 3, 5}
	timeout := 2 * time.Second

	heartbeat, results := SecurtyHeartInterval(done, timeout/2, nums...)
	<-heartbeat

	i := 0
	for {
		select {
		case r, ok := <-results:
			if !ok {
				return
			} else if r != nums[i] {
				t.Errorf("index %v expect %v, but result %v", i, nums[i], r)
			}
			i++
		case <-heartbeat:
		case <-time.After(timeout):
			t.Fatalf("time out")
		}
	}
}

func TestNewSteward(t *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	doWork := func(done <-chan interface{}, t time.Duration) <-chan interface{} {
		log.Println("ward: hello, I'm irresponsible")
		go func() {
			<-done
			log.Println("ward: I am halting")
		}()
		return nil
	}

	doWorkWithSteward := NewSteward(4*time.Second, doWork)

	done := make(chan interface{})
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward.")
		close(done)
	})

	for range doWorkWithSteward(done, 4*time.Second) {
	}

	log.Println("done")
}

func TestNewSteward2(t *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)
	// 信号流 ->
	done := make(chan interface{})
	defer close(done)
	dowork, intstream := doworkFn(done, 1, 2, -1, 3, 4, 5) // 模拟通信流(goroutine的状况)
	doWorkWithSteward := NewSteward(time.Millisecond, dowork)
	doWorkWithSteward(done, time.Hour)
	for val := range Take(done, intstream, 6) {
		log.Printf("received:%v\n", val)
	}

	log.Println("done")
}

func TestErrGroup(t *testing.T) {
	ErrGroup()
}

func TestWaitGroup(t *testing.T) {
	WaitGroup()
}
