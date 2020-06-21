package api

import (
	"testing"
	"time"
	"math/rand"
)

func TestUploadFlow(t *testing.T) {
	tmpfile, err := Flow("/home/user/Downloads/china.mp3")
	t.Log("err", err, tmpfile)
}

func TestPolling(t *testing.T) {
	s := Socket{}
	s.polling()
	s.polling2("1:2") // 请求 2
	s.polling1()      // 响应 3(使用polling)  1(未知) 6 (切换到websocket)
}

func TestPoll(t *testing.T) {
	s := Socket{}
	go func() {
		s.Poll()
	}()

	done := make(chan struct{})

	go func() {
		musics := []string{"11.mp3", "22.mp3", "33.mp3", "44.mp3", "55.mp3"}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		timer := time.NewTimer(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
		for {
			select {
			case <-timer.C:
				s.PollJob("/home/user/Downloads/"+musics[int(rnd.Int31n(5))], "amr")
				timer.Reset(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
			case <-done:
				return
			}
		}
	}()

	time.Sleep(10 * time.Minute)
	close(done)
}

func TestSocket(t *testing.T) {
	s := Socket{}

	go func() {
		err := s.Socket()
		if err != nil {
			t.Logf("socket: %v", err)
		}
	}()

	done := make(chan struct{})

	go func() {
		musics := []string{"11.mp3", "22.mp3", "33.mp3", "44.mp3", "55.mp3"}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		timer := time.NewTimer(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
		for {
			select {
			case <-timer.C:
				s.SocketJob("/home/quinn/Downloads/"+musics[int(rnd.Int31n(5))], "amr")
				timer.Reset(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
			case <-done:
				return
			}
		}
	}()

	time.Sleep(3 * time.Minute)
	close(done)
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

func TestConfig(t *testing.T) {
	config := New(
		"ya29.a0AfH6SMCz2ifEl8c5eTE7zgJr_jT8cJ0Udf4BLuwGNls8lbu0uRt4orhtlVH3Dwu1Z4WXrcoXHIoJ5w8y0XZb-5QWbfMdc7Kl_UXZ8aoba12qYFTwmzstkTjmqYZf0Yv93ZJs-d6N6AgtMFDU3e6VL2deZpgRGkPRoVY",
		"1//04yyh3J4Xm7AuCgYIARAAGAQSNwF-L9IrQbs2CS9iDooSSkeO_b7IBHoRyDoWFb0u1fudz_6jEWydFSqFJXMfHdr00OCN5vTRNpM",
	)

	err := config.refreshToken()
	t.Log("err", err)
}
