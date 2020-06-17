package api

import (
	"testing"
)

func TestUploadFlow(t *testing.T) {
	err := UploadFlow("/home/user/Downloads/china.mp3")
	t.Log("err", err)
}

func TestSocket(t *testing.T) {
	s := Socket{}
	err := s.polling1()
	if err != nil {
		t.Fatalf("polling1: %v", err)
	}
	t.Logf("%v", Unix())
	err = s.polling2()
	if err != nil {
		t.Fatalf("polling2: %v", err)
	}

	err = s.socket()
	if err != nil {
		t.Fatalf("socket: %v", err)
	}
}
