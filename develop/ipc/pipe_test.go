package ipc

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

var fifo *Fifo

func initFIFO(mode uint32, t *testing.T) {
	var err error
	fifo, err = FIFO("/tmp/fifo", mode)
	if err != nil {
		t.Errorf("FIFO: %v", err)
		os.Exit(1)
	}
	t.Logf("init success")
}

func TestFifo_Write(t *testing.T) {
	initFIFO(syscall.O_WRONLY, t)

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				data := fmt.Sprintf("%v - %v - %v", time.Now(), time.Now().Unix(), time.Now().Format(time.RFC3339))
				n, err := fifo.Write([]byte(data))
				if err != nil {
					t.Errorf("FIFO Write: %v", err)
					return
				}

				t.Logf("FIFO Write Success, length: %v", n)
				time.Sleep(10 * time.Millisecond * time.Duration(rand.Int31n(300)+10))
			}
		}()
	}

	wg.Wait()
}

func TestFifo_Read(t *testing.T) {
	initFIFO(syscall.O_RDWR, t)

	buf := make([]byte, 1024)
	for {
		n, err := fifo.Read(buf)
		if err != nil {
			t.Errorf("FIFO Read: %v", err)
			continue
		}

		t.Logf("FIFO Read Success, data: %s", buf[:n])
		time.Sleep(2 * time.Second)
	}
}
