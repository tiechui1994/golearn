package ipc

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"
	"unsafe"
)

var shm *ShareMemory
var sema *Sema

func InitShareM(t *testing.T) {
	var err error
	os.Create("/tmp/shm")
	shm, err = ShareM("/tmp/shm", 10, 4096, 0)
	if err != nil {
		t.Errorf("ShareM: %v", err)
		os.Exit(1)
	}
}

func InitSem(t *testing.T) {
	var err error
	os.Create("/tmp/shm")
	sema, err = Semaphore("/tmp/shm", 10, 1, 0666)
	if err != nil {
		t.Errorf("Semaphore: %v", err)
		os.Exit(1)
	}

	t.Logf("semid: %v", sema.semaid)

	err = sema.Init(1)
	if err != nil {
		t.Errorf("Init: %v", err)
		os.Exit(1)
	}
}

func TestShareM(t *testing.T) {
	InitShareM(t)
	defer func() {
		shm.Remove()
	}()

	err := shm.Link()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	var buf bytes.Buffer
	buf.WriteString(time.Now().Format(time.RFC3339Nano))

	data := (*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:]
	fmt.Fscanf(&buf, "%v", &data)

	t.Logf("read data: %v", string(data))

	err = shm.UnLink()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

func TestShareMemoryWrite(t *testing.T) {
	InitShareM(t)
	InitSem(t)
	err := shm.Link()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	var buf bytes.Buffer
	data := (*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:]

	for i := 0; i < 10; i++ {
		buf.Reset()
		buf.WriteString(time.Now().Format(time.RFC3339Nano))
		t.Logf("write: %s", buf.String())
		t.Logf("P: %v", sema.P())
		fmt.Fscanf(&buf, "%v", &data)
		t.Logf("V: %v", sema.V())
		time.Sleep(10 * time.Second)
		t.Log("-------------")
	}

	err = shm.UnLink()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

func TestShareMemoryRead(t *testing.T) {
	InitShareM(t)
	InitSem(t)
	err := shm.Link()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	data := (*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:]
	for i := 0; i < 10; i++ {
		t.Logf("P: %v", sema.P())
		t.Logf("read: %v", string(data))
		t.Logf("V: %v", sema.V())
		t.Log("-------------")
	}

	err = shm.UnLink()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}
