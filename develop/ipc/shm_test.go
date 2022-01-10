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

func InitShareM(t *testing.T, size, flag uint) {
	var err error
	os.Create("/tmp/shm")
	shm, err = ShareM("/tmp/shm", 10, size, flag)
	if err != nil {
		t.Errorf("ShareM: %v", err)
		os.Exit(1)
	}
	t.Logf("shmid: %v", shm.shid)
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
}

func TestShareM(t *testing.T) {
	InitShareM(t, 4096, IPC_CREAT)
	defer func() {
		shm.Remove()
	}()

	err := shm.Attach()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	var buf bytes.Buffer
	buf.WriteString(time.Now().Format(time.RFC3339Nano))

	data := (*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:]
	fmt.Fscanf(&buf, "%v", &data)

	t.Logf("read data: %v", string(data))

	err = shm.UnAttach()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

func TestShareMemoryWrite(t *testing.T) {
	InitShareM(t, 4096, 0600)
	InitSem(t)
	err := sema.Init(1)
	if err != nil {
		t.Errorf("Init: %v", err)
		os.Exit(1)
	}

	err = shm.Attach()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	// TODO: 每次都是直接写入到 address 当中.
	for i := 0; i < 10; i++ {
		t.Logf("P: %v", sema.P())
		content := time.Now().Format(time.RFC3339Nano)
		copy((*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:], content)
		t.Logf("write: %s", content)
	}

	err = shm.UnAttach()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

func TestShareMemoryRead(t *testing.T) {
	InitShareM(t, 0, 0600)

	InitSem(t)
	err := shm.Attach()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	// TODO: 每次都是直接从 address 当中读取内容.
	for i := 0; i < 10; i++ {
		t.Logf("read: %s", (*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:])
		t.Logf("V: %v", sema.V())
	}

	err = shm.UnAttach()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

