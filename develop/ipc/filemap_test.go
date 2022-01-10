package ipc

import (
	"math/rand"
	"testing"
	"time"
)

func fileMapWrite(t *testing.T, call func(om *FileMap)) {
	fm, err := OpenFileMap("/tmp/fm", WithWrite())
	if err != nil {
		t.Errorf("OpenFileMapping: %v", err)
		return
	}

	t.Logf("len: %v", len(fm.data))

	for i := 0; i < 100; i++ {
		call(fm)
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("Close: %v", err)
		return
	}
}

func TestFileMapWrite1(t *testing.T) {
	fileMapWrite(t, func(om *FileMap) {

		t.Logf("Write1 Read: %s", om.mapping)
		copy(om.mapping, "This is test data case")
		time.Sleep(time.Duration(rand.Int31n(822)) * time.Millisecond)
	})
}

func TestFileMapWrite2(t *testing.T) {
	fileMapWrite(t, func(om *FileMap) {
		t.Logf("Write2 Read: %s", om.mapping)
		copy(om.mapping, "I am TestFileMapWrite2")
		time.Sleep(time.Duration(rand.Int31n(945)) * time.Millisecond)
	})
}
