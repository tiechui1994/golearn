package ipc

import (
	"bytes"
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
	initFIFO(syscall.O_RDWR, t)

	CNT := 1
	var wg sync.WaitGroup
	wg.Add(CNT)

	for i := 0; i < CNT; i++ {
		go func(idx int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				data := fmt.Sprintf("[%v-%v - %v - %v]", idx, i, time.Now().Unix(), time.Now().Format(time.RFC3339))
				n, err := fifo.Write([]byte(data))
				if err != nil {
					t.Errorf("FIFO Write: %v", err)
					return
				}

				t.Logf("FIFO Write Success, length: %v", n)
				time.Sleep(5 * time.Millisecond * time.Duration(rand.Int31n(300)+10))
			}
		}(i)
	}

	wg.Wait()
}

func TestFifo_Read(t *testing.T) {
	initFIFO(syscall.O_RDWR, t)

	var temp []byte
	buf := make([]byte, 4096)
	for {
		n, err := fifo.Read(buf)
		if err != nil {
			t.Errorf("FIFO Read: %v", err)
			continue
		}

		// handle  sticky packget
		begin, end := 0, n
		nums := bytes.Count(buf[begin:end], []byte("]"))
		last := buf[n-1] == ']'
		for i := 0; i < nums; i++ {
			length := bytes.IndexByte(buf[begin:end], ']')

			if i == 0 && buf[0] != '[' {
				t.Logf("FIFO Read Success, data: %s", append(temp, buf[begin:begin+length+1]...))
				temp = nil
				begin += length + 1
				continue
			}

			t.Logf("FIFO Read Success, data: %s", buf[begin:begin+length+1])
			begin += length + 1
		}

		if !last {
			temp = make([]byte, len(buf[begin:end]))
			copy(temp, buf[begin:end])
		}

		time.Sleep(1 * time.Second)
	}
}
