package api

import (
	"bytes"
	"encoding/gob"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)


func TestPoll(t *testing.T) {
	s := Socket{}
	go func() {
		s.Poll()
	}()

	done := make(chan struct{})

	go func() {
		musics := []string{"11.mp3", "22.mp3"}
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		timer := time.NewTimer(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
		for {
			select {
			case <-timer.C:
				url, _ := s.PollJob("./data/"+musics[int(rnd.Int31n(2))], "amr")
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
		musics := []string{"11.mp3", "22.mp3"}
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
}

func TestSpeechToText(t *testing.T) {
	s := speechtotext{}
	// token 来源
	// https://azure.microsoft.com/en-us/services/cognitive-services/speech-to-text/
	go func() {
		region := "eastus"
		format := "simple"
		authorization := "eyJhbGciOiJodHRwOi8vd3d3LnczLm9yZy8yMDAxLzA0L3htbGRzaWctbW9yZSNobWFjLXNoYTI1NiIsInR5cCI6IkpXVCJ9.eyJyZWdpb24iOiJlYXN0dXMiLCJzdWJzY3JpcHRpb24taWQiOiIwOWIyNTQwZTg3Yjk0ZDgwYmUyMzAyYTc5OTcyNTIyOSIsInByb2R1Y3QtaWQiOiJTcGVlY2hTZXJ2aWNlcy5TMCIsImNvZ25pdGl2ZS1zZXJ2aWNlcy1lbmRwb2ludCI6Imh0dHBzOi8vYXBpLmNvZ25pdGl2ZS5taWNyb3NvZnQuY29tL2ludGVybmFsL3YxLjAvIiwiYXp1cmUtcmVzb3VyY2UtaWQiOiIvc3Vic2NyaXB0aW9ucy81NmI4ZjEwYS04M2NiLTQwYzYtYTU3ZS00OGQ2MWRlNjEzZjUvcmVzb3VyY2VHcm91cHMvY29nbml0aXZlLXNlcnZpY2VzLXByb2QvcHJvdmlkZXJzL01pY3Jvc29mdC5Db2duaXRpdmVTZXJ2aWNlcy9hY2NvdW50cy9hY29tLXByb2Qtc3BlZWNoLWVhc3R1cyIsInNjb3BlIjoic3BlZWNoc2VydmljZXMiLCJhdWQiOiJ1cm46bXMuc3BlZWNoc2VydmljZXMuZWFzdHVzIiwiZXhwIjoxNTkzNDM5NTA4LCJpc3MiOiJ1cm46bXMuY29nbml0aXZlc2VydmljZXMifQ.AeRaTLycanPaUAtBizcggHR8deerGtXGgISQlp4wcL4"
		err := s.Socket(region, lang_zh_cn, format, authorization)
		if err != nil {
			t.Logf("socket:%v", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		musics := []string{"11.wav", "22.wav"}
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

func TestLongSpeech(t *testing.T) {
	s := speechtotext{}
	go func() {
		region := "eastus"
		format := "simple"
		authorization := "eyJhbGciOiJodHRwOi8vd3d3LnczLm9yZy8yMDAxLzA0L3htbGRzaWctbW9yZSNobWFjLXNoYTI1NiIsInR5cCI6IkpXVCJ9.eyJyZWdpb24iOiJlYXN0dXMiLCJzdWJzY3JpcHRpb24taWQiOiIwOWIyNTQwZTg3Yjk0ZDgwYmUyMzAyYTc5OTcyNTIyOSIsInByb2R1Y3QtaWQiOiJTcGVlY2hTZXJ2aWNlcy5TMCIsImNvZ25pdGl2ZS1zZXJ2aWNlcy1lbmRwb2ludCI6Imh0dHBzOi8vYXBpLmNvZ25pdGl2ZS5taWNyb3NvZnQuY29tL2ludGVybmFsL3YxLjAvIiwiYXp1cmUtcmVzb3VyY2UtaWQiOiIvc3Vic2NyaXB0aW9ucy81NmI4ZjEwYS04M2NiLTQwYzYtYTU3ZS00OGQ2MWRlNjEzZjUvcmVzb3VyY2VHcm91cHMvY29nbml0aXZlLXNlcnZpY2VzLXByb2QvcHJvdmlkZXJzL01pY3Jvc29mdC5Db2duaXRpdmVTZXJ2aWNlcy9hY2NvdW50cy9hY29tLXByb2Qtc3BlZWNoLWVhc3R1cyIsInNjb3BlIjoic3BlZWNoc2VydmljZXMiLCJhdWQiOiJ1cm46bXMuc3BlZWNoc2VydmljZXMuZWFzdHVzIiwiZXhwIjoxNTkzNDQ1NjExLCJpc3MiOiJ1cm46bXMuY29nbml0aXZlc2VydmljZXMifQ.ecMRtMzXn0NJg5K-FEYqexDXB3_eRAOKWX8-WkSvYAM"
		err := s.Socket(region, lang_zh_cn, format, authorization)
		if err != nil {
			t.Logf("socket:%v", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(3 * time.Second)
		msg, _ := s.SendSpeech("./data/66.wav")
		log.Println("msg", msg)
	}()
	<-done
}

const (
	dir  = "/home/user/Downloads/speaker/mp3/"
	data = `
Alexa,open the door; 10039_rmmin3_Alexa_001_OpenDoor_en.amr
Alexa,close the door; 10039_rmmin3_Alexa_002_CloseDoor_en.amr
Alexa,lock the door; 10039_rmmin3_Alexa_001_LockDoor_en.amr
Alexa,unlock the door; 10039_rmmin3_Alexa_002_UnlockDoor_en.amr
`
)

func TestSpeechs(t *testing.T) {
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

	tokens := strings.Split(data, "\n")
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		token = strings.TrimSuffix(token, ";")
		if len(token) == 0 {
			continue
		}

		ts := strings.Split(token, ";")
		if len(ts) == 2 {
			text := strings.TrimSpace(ts[0])
			filename := strings.Replace(strings.TrimSpace(ts[1]), ".amr", ".mp3", 1)
			err = speech.Speech(text, dir+filename)
			t.Log("err", err)
			time.Sleep(200 * time.Millisecond)
		}

	}
}

func TestBaiduAudio(t *testing.T) {
	f := flag.String("input", "", "input voice text")
	o := flag.String("output", "", "output voice dir")
	flag.Parse()
	if f == nil || *f == "" {
		log.Println("input file not exist")
		return
	}
	if o == nil || *o == "" {
		log.Println("output file not exist")
		return
	}

	fd, err := os.Stat(*f)
	if err != nil || fd.IsDir() {
		log.Println("input file invalid")
		return
	}

	fd, err = os.Stat(*o)
	if err != nil && os.IsNotExist(err) {
		os.MkdirAll(*o, 0666)
	}

	data, _ := ioutil.ReadFile(*f)

loop:
	var buf bytes.Buffer
	tokens := strings.Split(string(data), "\n")
	for _, token := range tokens {
		t := strings.TrimSpace(token)
		t = strings.TrimSuffix(t, ";")
		if len(t) == 0 {
			continue
		}

		ts := strings.Split(t, ";")
		if len(ts) == 2 {
			text := strings.TrimSpace(ts[0])
			filename := strings.Replace(strings.TrimSpace(ts[1]), ".amr", ".mp3", 1)
			err := ConvertText(text, filepath.Join(*o, filename))
			if err != nil {
				buf.WriteString(t + "\n")
			}
			time.Sleep(5000 * time.Millisecond)
		}
	}

	if buf.Len() > 0 {
		data = buf.Bytes()
		goto loop
	}
}

func TestDNS(t *testing.T) {
	ips, err := DNS("avs-alexa-12-na.amazon.com", DNS_A)
	t.Log(err)
	t.Log(ips)
}
