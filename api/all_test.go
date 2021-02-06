package api

import (
	"bytes"
	"context"
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

func TestSpeechify(t *testing.T) {
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
	ctx, cancel := context.WithCancel(context.Background())
	s := speechtotext{}
	go func() {
		format := "simple"
		err := s.Socket(ctx, lang_zh_cn, format)
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
				cancel()
				close(done)
				break
			}

			select {
			case <-timer.C:
				count++
				msg, _ := s.SendSpeech("./data/" + musics[int(rnd.Int31n(int32(len(musics))))])
				log.Println("msg", msg)
				timer.Reset(time.Duration(rnd.Int63n(int64(time.Minute))) + time.Second)
			}
		}
	}()

	<-done
}

func TestLongSpeechOnce(t *testing.T) {
	s := speechtotext{}
	ctx := context.Background()
	go func() {
		format := "simple"
		err := s.Socket(ctx, lang_zh_cn, format)
		if err != nil {
			t.Logf("socket:%v", err)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(3 * time.Second)
		msg, _ := s.SendSpeech("./data/11.wav")
		log.Println("msg", msg)
	}()
	<-done
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

func TestLogin(t *testing.T) {
	pri, pub := GenKey()
	oeap := OAEP{prikey: pri, pubkey: pub, alg: "sha1"}
	oeap.init()

	plain := "hello"

	ciphertxt := oeap.Encrypt([]byte(plain))
	t.Log("ciphertxt", ciphertxt)

	pliantext := oeap.Decrypt(ciphertxt)
	t.Log("pliantext", string(pliantext), len(pliantext))

	t.Log("equal", string(plain) == string(pliantext))
}
