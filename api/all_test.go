package api

import (
	"testing"
	"sync"
	"time"
)

func TestUploadFlow(t *testing.T) {
	err := UploadFlow("/home/user/Downloads/china.mp3")
	t.Log("err", err)
}

func TestSocket(t *testing.T) {
	s := Socket{sid: "XtRcAa1_UKMr4579BJIi"}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := s.polling1()
		if err != nil {
			t.Fatalf("polling1: %v", err)
		}
	}()

	go func() {
		time.Sleep(10000*time.Microsecond)
		defer wg.Done()
		err := s.polling2()
		if err != nil {
			t.Fatalf("polling2: %v", err)
		}
	}()

	wg.Wait()

	err := s.socket()
	if err != nil {
		t.Fatalf("socket: %v", err)
	}
}
