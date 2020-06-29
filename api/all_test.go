package api

import (
	"testing"
	"time"
	"math/rand"
	"log"
	"encoding/gob"
	"os"
	"io/ioutil"
	"sync"
)

func TestUploadFlow(t *testing.T) {
	tmpfile, err := Flow("/home/user/Downloads/china.mp3")
	t.Log("err", err, tmpfile)
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
				url, _ := s.PollJob("./data/"+musics[int(rnd.Int31n(5))], "amr")
				log.Println("success,", url)
				timer.Reset(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
			case <-done:
				return
			}
		}
	}()

	time.Sleep(5 * time.Minute)
	close(done)
}

func TestSocket(t *testing.T) {
	s := Socket{}

	go func() {
		err := s.Socket()
		if err != nil {
			t.Logf("socket:%v", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		musics := []string{"11.mp3", "22.mp3", "33.mp3", "44.mp3", "55.mp3"}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		timer := time.NewTimer(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
		var count int64
		for {
			if count > 200 {
				close(done)
				break
			}

			select {
			case <-timer.C:
				count++
				uri, _ := s.SocketJob("./data/"+musics[int(rnd.Int31n(5))], "amr")
				log.Println("download", uri)
				timer.Reset(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
			}
		}
	}()

	<-done
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
	config := New()
	u := config.GetGoogleCode()
	t.Log(u)
	err := config.getAccessToken("4/1AGkOZWoCf4leuSIhP9VDf50X4MD2OJkDJY5zdA5N0uDpqM8Hu5AUnqKTt7GfGJL6UnQa-FKhsLw3qXOX2yYGbI")
	if err == nil {
		config.Task()
	}
}

func init() {
	gob.Register(Speech{})
}

func TestOCR(t *testing.T) {
	_, err := OCR("./data/chn.png", Lan_zhs)
	t.Log("err", err)
}

func TestSpeech(t *testing.T) {
	speech := Speech{Region: "us-west-2"}
	fd, err := os.Open("./data/speech.data")
	if err != nil || time.Now().Unix() > int64(speech.Expiration) {
		err := speech.Identity()
		if err != nil {
			t.Fatal("err", err)
		}
		fd, _ = os.Create("./data/speech.data")
		gob.NewEncoder(fd).Encode(&speech)
	} else {
		gob.NewDecoder(fd).Decode(&speech)
	}

	err = speech.TextSplit()
	t.Log("err", err)

	text := `我现在去自由工厂, 需要带东西的小伙伴们安排起来,公司2.10准时出发`
	err = speech.Speech(text, "./data/11.mp3")
	t.Log("err", err)

	text = `刚才是哪个小姐姐来市场部借三脚架来着,现在可以了`
	err = speech.Speech(text, "./data/22.mp3")
	t.Log("err", err)

	text = `等下16:00系统升级,需要重启任务系统, 请大家做好准备, 以免影响任务的提交和处理. 感谢大家的支持. `
	err = speech.Speech(text, "./data/33.mp3")
	t.Log("err", err)

	text = `f084的日志不会自动关的？每天日志大量上报还占用数据上报的带宽`
	err = speech.Speech(text, "./data/44.mp3")
	t.Log("err", err)

	text = `四号线浦沿地铁口新浦苑小区顶层单间出租,带独卫`
	err = speech.Speech(text, "./data/55.mp3")
	t.Log("err", err)
}

func TestSpeechToText(t *testing.T) {
	s := speechtotext{}
	// token 来源
	// https://azure.microsoft.com/en-us/services/cognitive-services/speech-to-text/
	go func() {
		region := "eastus"
		format := "simple"
		authorization := "eyJhbGciOiJodHRwOi8vd3d3LnczLm9yZy8yMDAxLzA0L3htbGRzaWctbW9yZSNobWFjLXNoYTI1NiIsInR5cCI6IkpXVCJ9.eyJyZWdpb24iOiJlYXN0dXMiLCJzdWJzY3JpcHRpb24taWQiOiIwOWIyNTQwZTg3Yjk0ZDgwYmUyMzAyYTc5OTcyNTIyOSIsInByb2R1Y3QtaWQiOiJTcGVlY2hTZXJ2aWNlcy5TMCIsImNvZ25pdGl2ZS1zZXJ2aWNlcy1lbmRwb2ludCI6Imh0dHBzOi8vYXBpLmNvZ25pdGl2ZS5taWNyb3NvZnQuY29tL2ludGVybmFsL3YxLjAvIiwiYXp1cmUtcmVzb3VyY2UtaWQiOiIvc3Vic2NyaXB0aW9ucy81NmI4ZjEwYS04M2NiLTQwYzYtYTU3ZS00OGQ2MWRlNjEzZjUvcmVzb3VyY2VHcm91cHMvY29nbml0aXZlLXNlcnZpY2VzLXByb2QvcHJvdmlkZXJzL01pY3Jvc29mdC5Db2duaXRpdmVTZXJ2aWNlcy9hY2NvdW50cy9hY29tLXByb2Qtc3BlZWNoLWVhc3R1cyIsInNjb3BlIjoic3BlZWNoc2VydmljZXMiLCJhdWQiOiJ1cm46bXMuc3BlZWNoc2VydmljZXMuZWFzdHVzIiwiZXhwIjoxNTkzNDMzMTk2LCJpc3MiOiJ1cm46bXMuY29nbml0aXZlc2VydmljZXMifQ.puBuflcMurQZlDJ_ShbAhkie1RAtdR038-XpLCUtdE0"
		err := s.Socket(region, lang_zh_cn, format, authorization)
		if err != nil {
			t.Logf("socket:%v", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		musics := []string{"11.wav", "22.wav", "33.wav", "44.wav", "55.wav"}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		timer := time.NewTimer(time.Duration(rnd.Int63n(int64(5*time.Second))) + time.Second)
		var count int64
		for {
			if count > 20 {
				close(done)
				break
			}

			select {
			case <-timer.C:
				count++
				msg, _ := s.SendSpeech("./data/" + musics[int(rnd.Int31n(5))])
				log.Println("msg", msg)
				timer.Reset(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
			}
		}
	}()

	<-done
}

func TestEscape(t *testing.T) {
	data, _ := ioutil.ReadFile("/home/user/Downloads/test.wav")
	// header: 42
	// body: 2976

	id := "28B0477EA6074BD9A3E075F6B8CE5FEE"
	var n int64
	header := data[0:44]
	cmd := getBinaryCmd(id, true, header)
	ioutil.WriteFile("./header", cmd, 0777)
	t.Log(len(cmd))
	n += 44

	var o sync.Once
	var step int64 = 3200
	for n < int64(len(data)) && n+step < int64(len(data)) {
		body := data[n:n+step]
		cmd = getBinaryCmd(id, false, body)
		o.Do(func() {
			ioutil.WriteFile("./body", cmd, 0777)
		})
		t.Log(len(cmd))
		n += 2976
	}

	cmd = getBinaryCmd(id, false, nil)
	ioutil.WriteFile("./tail", cmd, 0777)
	t.Log(len(cmd))
}
