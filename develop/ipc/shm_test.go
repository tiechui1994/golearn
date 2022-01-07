package ipc

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
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
	InitShareM(t, 4096, 0600)
	InitSem(t)
	err := sema.Init(1)
	if err != nil {
		t.Errorf("Init: %v", err)
		os.Exit(1)
	}

	err = shm.Link()
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

	err = shm.UnLink()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

func TestShareMemoryRead(t *testing.T) {
	InitShareM(t, 0, 0600)

	InitSem(t)
	err := shm.Link()
	if err != nil {
		t.Errorf("Link: %v", err)
		return
	}

	// TODO: 每次都是直接从 address 当中读取内容.
	for i := 0; i < 10; i++ {
		t.Logf("read: %s", (*(*[4096]byte)(unsafe.Pointer(shm.Address())))[:])
		t.Logf("V: %v", sema.V())
	}

	err = shm.UnLink()
	if err != nil {
		t.Errorf("UnLink: %v", err)
		return
	}
}

func TestOpenFileMapping(t *testing.T) {
	fm, err := OpenFileMapping("/tmp/fm", 1024, WithWrite())
	if err != nil {
		t.Errorf("OpenFileMapping: %v", err)
		return
	}

	t.Logf("len: %v", len(fm.data))

	data := (*(*[1024]byte)(unsafe.Pointer(&fm.data)))[:]
	//t.Logf("copy data: %v", copy(data, "Hello world"))
	t.Logf("data: %s", data)

	err = fm.Close()
	if err != nil {
		t.Errorf("Close: %v", err)
		return
	}
}

func memory(t *testing.T, mode int) {
	shmid, _, err := syscall.Syscall(syscall.SYS_SHMGET, 5, 1024, 01000)
	if err != 0 {
		t.Errorf("syscall error, err: %v", err)
		os.Exit(-1)
	}
	t.Logf("shmid: %v", shmid)

	shmaddr, _, err := syscall.Syscall(syscall.SYS_SHMAT, shmid, 0, 0)
	if err != 0 {
		t.Errorf("syscall error, err: %v", err)
		os.Exit(-2)
	}
	t.Logf("shmaddr: %v", shmaddr)

	defer syscall.Syscall(syscall.SYS_SHMDT, shmaddr, 0, 0)

	if mode == 0 {
		t.Log("write mode")
		i := 1000
		for {
			sema.P()
			content := time.Now().Format(time.RFC3339Nano)
			//*(*int)(unsafe.Pointer(uintptr(shmaddr))) = i
			copy((*(*[1024]uint8)(unsafe.Pointer(uintptr(shmaddr))))[:], content)
			t.Logf("write: %s", content)
			i++
			time.Sleep(1 * time.Second)
		}
	} else {
		t.Log("read mode")
		for {
			sema.V()
			//t.Logf("read: %d", *(*int)(unsafe.Pointer(uintptr(shmaddr))))
			t.Logf("read: %s", (*(*[1024]uint8)(unsafe.Pointer(uintptr(shmaddr))))[:])
			time.Sleep(1 * time.Second)
		}
	}
}

func TestMemoryWrite(t *testing.T) {
	InitSem(t)
	err := sema.Init(1)
	if err != nil {
		t.Errorf("Init: %v", err)
		os.Exit(1)
	}
	memory(t, 0)
}

func TestMemoryRead(t *testing.T) {
	InitSem(t)
	memory(t, 1)
}
