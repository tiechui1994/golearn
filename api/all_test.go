package api

import (
	"testing"
	"time"
)

func TestUploadFlow(t *testing.T) {
	tmpfile, err := UploadFlow("/home/user/Downloads/china.mp3")
	t.Log("err", err, tmpfile)
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

	go func() {
		for {
			err := s.polling2()
			if err != nil {
				t.Logf("polling2: %v", err)
				break
			}

			err = s.polling1()
			if err != nil {
				t.Logf("polling1: %v", err)
				break
			}
		}
	}()

}

func TestSocket(t *testing.T) {
	s := Socket{}

	err := s.polling()
	if err != nil {
		t.Logf("polling1: %v", err)
		return
	}

	go func() {
		err = s.socket()
		if err != nil {
			t.Logf("socket: %v", err)
		}
	}()

	time.Sleep(5 * time.Second)

	s.socketJob("/home/user/Downloads/china.mp3", "ogg")
	select {

	}
}

func TestUnix(t *testing.T) {
	t.Logf("Now: %v", time.Now().UnixNano()/1e6)
	s := tencode()
	t.Logf("Unix: %v", s)
	ts := tdecode("NB5mBqc")
	t.Logf("Timestamp: %v", ts)
}

func TestZip(t *testing.T) {
	files := []string{"s112pfWR989w_amr_7RRRt6Bm.wav",
		"s112A9L9bowG_amr_rC92G4s7.wav",
		"s112rdXJSUDR_amr_i7w3uB5B.wav",
		"s112WItIrACq_amr_4HvtB3cy.wav",
		"s112bk0l1cM7_amr_9705KVc0.wav",
		"s1121GuQTxu6_amr_YadZucs0.wav",
		"s112mZ1BOK42_amr_w08VK131.wav",
		"s112yCrkzWD6_amr_xONnCUpk.wav",
		"s112xMTUX8XU_amr_w4p4w5XG.wav",
		"s1128ZYzbyPW_amr_6Ln0kidb.wav",
		"s112wJ1b2g7o_amr_sjX4XYrG.wav"}
	uri, err := Zip(files)
	t.Log(uri, err)
}
