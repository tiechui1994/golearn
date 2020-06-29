package api

import (
	"fmt"
	"time"
	"net/http"
	"log"
	"io/ioutil"
	"sync"
	"errors"
	"strings"
	"encoding/json"

	"github.com/gorilla/websocket"
)

const (
	lang_zh_hk = "zh-HK" // Chinese (Cantonese Traditional)
	lang_zh_cn = "zh-CN" // Chinese (Mandarin simplified)
	lang_zh_tw = "zh-TW" // Chinese (Taiwanese Mandarin)

	lang_en_us = "en-US" // English (United States)
	lang_en_au = "en-AU" // English (Australia)
	lang_en_ca = "en-CA" // English (Canada)
	lang_en_in = "en-IN" // English (India)
	lang_en_nz = "en-NZ" // English (New Zealand)
	lang_en_gb = "en-GB" // English (United Kingdom)

	lang_de_de = "de-DE" // German (Germany)

	lang_ja_jp = "ja-JP" // Japanese (Japan)

	lang_ko_kr = "ko-KR" // Korean (Korea)
)

type speechtotext struct {
	*websocket.Conn
	writeLock sync.Mutex
	jobLock   sync.Mutex
	closed    bool
	done      chan struct{}
	text      chan string
	phrase    map[string]time.Time

	ConnID string
	JobID  string
}

const (
	phrase_conn_listen = "conn.listen"
	phrase_conn_start  = "conn.start"
	phrase_conn_end    = "conn.end"

	phrase_turn_start         = "turn.start"
	phrase_turn_end           = "turn.end"
	phrase_speech_enddetected = "speech.endDetected"
	phrase_speech_phrase      = "speech.phrase"
)

func (s *speechtotext) Socket(region, language, format, authorization string) error {
	timeoutMs := 5000
	xConnectionId := strings.ToUpper(MD5(time.Now().String()))
	u := "wss://" + region + ".stt.speech.microsoft.com/speech/recognition/dictation/cognitiveservices/v1"
	u = fmt.Sprintf("%v?language=%v&format=%vinitialSilenceTimeoutMs=%v&endSilenceTimeoutMs=%v&Authorization=%v&X-ConnectionId=%v",
		u, language, format, timeoutMs, timeoutMs, authorization, xConnectionId)

	dailer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	s.ConnID = xConnectionId
	s.text = make(chan string)
	s.done = make(chan struct{})
	s.phrase = map[string]time.Time{
		phrase_conn_listen: time.Now(),
	}
	var retryTimeout = time.Second

again:
	header := make(http.Header)
	s.phrase[phrase_conn_start] = time.Now()
	conn, response, err := dailer.Dial(u, header)
	if err != nil && retryTimeout < 64*time.Second {
		log.Println("dailer.Dial", err)
		errmsg, _ := ioutil.ReadAll(response.Body)
		log.Println("message", string(errmsg))
		time.Sleep(retryTimeout)
		retryTimeout *= 2
		goto again
	}

	if err != nil {
		log.Println("dailer.Dial", err)
		return err
	}

	log.Println("success to connect ....")

	s.phrase[phrase_conn_end] = time.Now()
	s.Conn = conn
	s.JobID = strings.ToUpper(MD5(time.Now().String()))
	s.write(getConfigCmd(s.JobID), websocket.BinaryMessage)
	s.write(getContextCmd(s.JobID), websocket.BinaryMessage)

	log.Println("success to send config ....")

	for {
		data, ok := s.read()
		if ok {
			goto close
		}

		text := jsonCmdDecode(data)
		if text.RecognitionStatus == "Success" {
			select {
			case <-s.done:
				return nil
			case s.text <- text.DisplayText:
			}
		} else if text.RecognitionStatus == "InitialSilenceTimeout" {
			getTelemetryCmd(s.JobID, s.ConnID, s.phrase)
			s.JobID = strings.ToUpper(MD5(time.Now().String()))
			getContextCmd(s.JobID)
			close(s.text)
		} else {
			select {
			case <-s.done:
				return nil
			default:
				s.phrase[text.Phrase] = time.Now()
			}
		}
	}

close:
	return errors.New("exception cloded")
}

func (s *speechtotext) read() (data string, isclose bool) {
	if s.closed {
		return data, s.closed
	}

	_, message, err := s.Conn.ReadMessage()
	if s.socketError(err) {
		return data, true
	}

	return string(message), false
}

func (s *speechtotext) write(data []byte, ttype int) (isclose bool) {
	if s.closed {
		return s.closed
	}

	s.writeLock.Lock()
	defer s.writeLock.Unlock()
	err := s.Conn.WriteMessage(ttype, []byte(data))
	return s.socketError(err)
}

func (s *speechtotext) socketError(err error) (isclose bool) {
	if err == nil {
		return false
	}

	if _, ok := err.(*websocket.CloseError); ok {
		log.Printf("err: %v", err)
		s.Close()
		return true
	}

	log.Printf("failed, %v", err)
	return
}

const (
	len_header = 44   // header: 44
	len_body   = 3200 // body: 3200
)

func (s *speechtotext) SendSpeech(src string) (msg string, err error) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		log.Println("err", err)
		return msg, err
	}

	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	log.Println("data len", len(data))

retry:
	var (
		n      int64
		length = int64(len(data))
	)
	header := data[0:len_header]
	cmd := getBinaryCmd(s.JobID, true, header)
	s.write(cmd, websocket.BinaryMessage)
	n += len_header

	for n < length {
		var body []byte
		if n+len_body < length {
			body = data[n:n+len_body]
		} else {
			body = data[n:]
		}
		cmd = getBinaryCmd(s.JobID, false, body)
		s.write(cmd, websocket.BinaryMessage)
		n += len_body
	}

	cmd = getBinaryCmd(s.JobID, false, nil)
	s.write(cmd, websocket.BinaryMessage)
	log.Println("write success")
	msg, ok := <-s.text
	if !ok {
		log.Println("retury....")
		s.text = make(chan string)
		goto retry
	}
	s.JobID = strings.ToUpper(MD5(time.Now().String()))

	return msg, nil
}

func (s *speechtotext) Close() {
	if s.closed {
		return
	}

	s.closed = true

	if s.done != nil {
		close(s.done)
	}

	if s.Conn != nil {
		s.Conn.Close()
	}
}

func getConfigCmd(id string) []byte {
	path := "speech.config"
	var result struct {
		Context struct {
			System struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Build   string `json:"build"`
				Lang    string `json:"lang"`
			} `json:"system"`
			Os struct {
				Platform string `json:"platform"`
				Name     string `json:"name"`
				Version  string `json:"version"`
			} `json:"os"`
			Audio struct {
				Source struct {
					Bitspersample int    `json:"bitspersample"`
					Channelcount  int    `json:"channelcount"`
					Connectivity  string `json:"connectivity"`
					Manufacturer  string `json:"manufacturer"`
					Model         string `json:"model"`
					Samplerate    int    `json:"samplerate"`
					Type          string `json:"type"`
				} `json:"source"`
			} `json:"audio"`
		} `json:"context"`
		Recognition string `json:"recognition"`
	}

	result.Recognition = "conversation"

	result.Context.System.Name = "SpeechSDK"
	result.Context.System.Version = "1.12.1-rc.1"
	result.Context.System.Build = "JavaScript"
	result.Context.System.Lang = "JavaScript"

	result.Context.Os.Platform = "Browser/Linux x86_64"
	result.Context.Os.Name = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"
	result.Context.Os.Version = "5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"

	result.Context.Audio.Source.Bitspersample = 16
	result.Context.Audio.Source.Channelcount = 1
	result.Context.Audio.Source.Connectivity = "Unknown"
	result.Context.Audio.Source.Manufacturer = "Speech SDK"
	result.Context.Audio.Source.Model = "File"
	result.Context.Audio.Source.Samplerate = 16000
	result.Context.Audio.Source.Type = "File"

	data, _ := json.Marshal(result)
	return jsonCmdEncode(path, id, string(data))
}

func getContextCmd(id string) []byte {
	path := "speech.context"
	return jsonCmdEncode(path, id, "{}")
}

func getTelemetryCmd(id, connid string, times map[string]time.Time) []byte {
	listen := times[phrase_conn_listen].UTC().Format(timeformat)
	cstart := times[phrase_conn_start].UTC().Format(timeformat)
	cend := times[phrase_conn_end].UTC().Format(timeformat)

	tt := times[phrase_turn_start].UTC().Format(timeformat)
	td := times[phrase_turn_end].UTC().Format(timeformat)
	sd := times[phrase_speech_enddetected].UTC().Format(timeformat)
	se := times[phrase_speech_phrase].UTC().Format(timeformat)
	ms := times[phrase_turn_end].Sub(times[phrase_conn_listen]).Nanoseconds() / 1e6
	path := "telemetry"
	body := fmt.Sprintf(`{
		"Metrics":[
			{
				"End":"%v",
				"Name":"ListeningTrigger",
				"Start":"%v"
			},
  			{
				"End":"%v",
				"Id":"%v",
				"Name":"Connection",
				"Start":"%v"
			},
			{
				"PhraseLatencyMs":[%v]
			}
		],
		"ReceivedMessages": {
			"turn.start":["%v"],
			"speech.endDetected":["%v"],
			"speech.phrase":["%v"],
			"turn.end":["%v"]
		}
	}`, listen, listen, cend, connid, cstart, ms, tt, sd, se, td)
	return jsonCmdEncode(path, id, body)
}

const (
	timeformat = "2006-01-02T15:04:05.000Z"
	sep        = "\r\n"
)

func getBinaryCmd(id string, isHeader bool, data []byte) []byte {
	var arrs []string
	if isHeader {
		arrs = []string{
			"~Path: audio",
			"X-RequestId: " + id,
			"X-Timestamp: " + time.Now().UTC().Format(timeformat),
			"Content-Type: audio/x-wav",
			"",
		}
	} else {
		arrs = []string{
			"cPath: audio",
			"X-RequestId: " + id,
			"X-Timestamp: " + time.Now().UTC().Format(timeformat),
			"",
		}
	}

	header := strings.Join(arrs, sep)

	return append([]byte{0x00}, append([]byte(header), data...)...)
}

func jsonCmdEncode(path, id, data string) []byte {
	arrs := []string{
		"Path: " + path,
		"X-RequestId: " + id,
		"X-Timestamp: " + time.Now().UTC().Format(timeformat),
		"Content-Type: application/json",
		"",
		data,
	}

	return []byte(strings.Join(arrs, sep))
}

type textresponse struct {
	Requestid         string
	Phrase            string
	Id                string
	RecognitionStatus string
	DisplayText       string
	Offset            int
	Duration          int
}

func jsonCmdDecode(data string) (text textresponse) {
	tokens := strings.Split(data, "\n")
	if len(tokens) < 5 {
		log.Println("data", data)
		return
	}

	var requestid, pharse string
	for i := 0; i < len(tokens); i++ {
		log.Println(tokens[i])
		if strings.Contains(tokens[i], "X-RequestId") {
			k := strings.Index(tokens[i], ":")
			requestid = strings.TrimSpace(tokens[i][k+1:])
		}
		if strings.Contains(tokens[i], "Path") {
			k := strings.Index(tokens[i], ":")
			pharse = strings.TrimSpace(tokens[i][k+1:])
		}
	}

	json.Unmarshal([]byte(tokens[4]), &text)
	text.Requestid = requestid
	text.Phrase = pharse

	return text
}
