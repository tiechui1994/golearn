package api

import (
	"testing"
	"time"
)

func TestUploadFlow(t *testing.T) {
	err := UploadFlow("/home/user/Downloads/china.mp3")
	t.Log("err", err)
}

func TestPoll(t *testing.T) {
	s := Socket{}

	err := s.polling()
	if err != nil {
		t.Fatalf("polling1: %v", err)
	}

	for {
		err := s.polling1()
		if err != nil {
			t.Logf("polling1: %v", err)
			break
		}

		err = s.polling()
		if err != nil {
			t.Logf("polling: %v", err)
			break
		}
	}

}

func TestPolling(t *testing.T) {
	s := Socket{}

	err := s.polling()
	if err != nil {
		t.Fatalf("polling1: %v", err)
	}

	var i = 0

	for {
		i++
		err := s.polling2()
		if err != nil {
			t.Logf("polling2: %v", err)
			break
		}

		i++
		err = s.polling1()
		if err != nil {
			t.Logf("polling1: %v", err)
			break
		}
	}

	t.Logf("total: %v", i)
}

func TestSocket(t *testing.T) {
	s := Socket{}

	err := s.polling()
	if err != nil {
		t.Logf("polling1: %v", err)
		return
	}

	err = s.socket()
	if err != nil {
		t.Logf("socket: %v", err)
		return
	}
}

func TestUnix(t *testing.T) {
	t.Logf("Now: %v", time.Now().UnixNano()/1e6)
	s := encode()
	t.Logf("Unix: %v", s)
	ts := decode("NB5mBqc")
	t.Logf("Timestamp: %v", ts)
}
