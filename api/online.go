package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var oclient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Minute)
		},
	},
	Timeout: time.Minute,
}

const (
	host          = "s113.123apps.com"
	uid           = "MzjycdJYT3eFIabPmFK5eeb1ee17ad8e"
	flowChunkSize = 52428800
)

type Audio struct {
	Index          int    `json:"index"`
	CodecName      string `json:"codec_name"`
	CodecLongName  string `json:"codec_long_name"`
	CodecType      string `json:"codec_type"`
	CodecTimeBase  string `json:"codec_time_base"`
	CodecTagString string `json:"codec_tag_string"`
	CodecTag       string `json:"codec_tag"`
	SampleFmt      string `json:"sample_fmt"`
	SampleRate     string `json:"sample_rate"`
	Channels       int    `json:"channels"`
	ChannelLayout  string `json:"channel_layout"`
	BitsPerSample  int    `json:"bits_per_sample"`
	RFrameRate     string `json:"r_frame_rate"`
	AvgFrameRate   string `json:"avg_frame_rate"`
	TimeBase       string `json:"time_base"`
	StartPts       int    `json:"start_pts"`
	StartTime      string `json:"start_time"`
	DurationTs     int64  `json:"duration_ts"`
	Duration       string `json:"duration"`
	BitRate        string `json:"bit_rate"`
	Disposition    struct {
		Default         int `json:"default"`
		Dub             int `json:"dub"`
		Original        int `json:"original"`
		Comment         int `json:"comment"`
		Lyrics          int `json:"lyrics"`
		Karaoke         int `json:"karaoke"`
		Forced          int `json:"forced"`
		HearingImpaired int `json:"hearing_impaired"`
		VisualImpaired  int `json:"visual_impaired"`
		CleanEffects    int `json:"clean_effects"`
		AttachedPic     int `json:"attached_pic"`
		TimedThumbnails int `json:"timed_thumbnails"`
	} `json:"disposition"`
}

// /aconv/upload/flow/?
// uid=2VlBGJGDkxWMXGYGHs65e5cba075f488&
// id3=true&
// ff=true&
// flowChunkNumber=1&
// flowChunkSize=52428800&
// flowCurrentChunkSize=6182&
// flowTotalSize=6182&
// flowIdentifier=6182-Light_Google_002_set50_enamr&
// flowFilename=Light_Google_002_set50_en.amr&
// flowRelativePath=Light_Google_002_set50_en.amr&
// flowTotalChunks=1
func Flow(filename string) (tmpfile string, err error) {
	fd, err := os.Open(filename)
	if err != nil {
		return tmpfile, err
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return tmpfile, err
	}

	n := stat.Size() / flowChunkSize
	if stat.Size()%flowChunkSize != 0 {
		n += 1
	}

	data := make([]byte, flowChunkSize)
	total := int64(0)
	for i := int64(1); i <= n; i++ {
		currentSize := int64(flowChunkSize)
		if currentSize >= stat.Size()-total {
			currentSize = stat.Size() - total
		}

		values := make(url.Values)
		values.Set("uid", uid)
		values.Set("id3", "true")
		values.Set("ff", "true")
		values.Set("flowChunkNumber", fmt.Sprintf("%v", i))
		values.Set("flowChunkSize", fmt.Sprintf("%v", flowChunkSize))
		values.Set("flowCurrentChunkSize", fmt.Sprintf("%v", currentSize))
		values.Set("flowTotalSize", fmt.Sprintf("%v", stat.Size()))
		values.Set("flowIdentifier", fmt.Sprintf("%v-%v", stat.Size(), strings.Replace(stat.Name(), ".", "", -1)))
		values.Set("flowFilename", stat.Name())
		values.Set("flowRelativePath", stat.Name())
		values.Set("flowTotalChunks", fmt.Sprintf("%v", n))

		n, _ := fd.ReadAt(data, total)
		data = data[:n]
		total += flowChunkSize

		u := "https://" + host + "/aconv/upload/flow/?" + values.Encode()
		request, _ := http.NewRequest("GET", u, nil)
		response, err := oclient.Do(request)
		if err != nil {
			return tmpfile, err
		}

		log.Println("Status:", response.StatusCode)

		tmpfile, err = flow(values, data)
		if err != nil {
			return tmpfile, err
		}

		log.Println("tmpfile", tmpfile)
	}

	return tmpfile, nil
}

func flow(vals url.Values, data []byte) (tmpfile string, err error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for k, v := range vals {
		writer.WriteField(k, v[0])
	}

	fd, err := writer.CreateFormFile("file", vals.Get("flowFilename"))
	if err != nil {
		return tmpfile, err
	}
	fd.Write(data)

	contentType := writer.FormDataContentType()
	writer.Close()

	u := "https://" + host + "/aconv/upload/flow/"
	request, _ := http.NewRequest("POST", u, &body)
	request.Header.Set("Content-Type", contentType)

	response, err := oclient.Do(request)
	if err != nil {
		return tmpfile, err
	}
	defer response.Body.Close()

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return tmpfile, err
	}

	log.Printf("%v", string(bs))
	var result struct {
		Error            int    `json:"error"`
		FileSize         int64  `json:"filesize"`
		TmpFilename      string `json:"tmp_filename"`
		OriginalFilename string `json:"original_filename"`
		FileType         string `json:"file_type"`
		Id3              struct {
			TagWasRead          bool   `json:"tag_was_read"`
			TrackTagGenre       string `json:"track_tag_genre"`
			TrackTagTracknumber string `json:"track_tag_tracknumber"`
			Bitrate             int    `json:"bitrate"`
			TrackBitrate        int    `json:"track_bitrate"`
			DurationInSeconds   int    `json:"duration_in_seconds"`
		} `json:"id3"`
		Ff struct {
			FfprobeSuccess         bool   `json:"ffprobe_success"`
			DurationInSeconds      int    `json:"duration_in_seconds"`
			DurationInMilliseconds int    `json:"duration_in_milliseconds"`
			Bitrate                int    `json:"bitrate"`
			HasAudioStreams        bool   `json:"has_audio_streams"`
			HasVedioStreams        bool   `json:"has_vedio_streams"`
			Filesize               string `json:"filesize"`
			Streams                struct {
				Audio []Audio `json:"audio"`
			} `json:"streams"`
			Format struct {
				NbStreams      int    `json:"nb_streams"`
				NbPrograms     int    `json:"nb_programs"`
				FormatName     string `json:"format_name"`
				FormatLongName string `json:"format_long_name"`
				Duration       string `json:"duration"`
				BitRate        string `json:"bit_rate"`
				ProbeScore     int    `json:"probe_score"`
			} `json:"format"`
		} `json:"ff"`
	}
	err = json.Unmarshal(bs, &result)
	if err != nil {
		log.Println("json", err)
		return tmpfile, err
	}

	return result.TmpFilename, nil
}

func Zip(files []string) (uri string, err error) {
	values := make(url.Values)
	values.Set("uid", uid)
	values.Set("files", strings.Join(files, ","))
	u := "https://" + host + "/aconv/zip/"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(values.Encode()))
	request.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	response, err := oclient.Do(request)
	if err != nil {
		return uri, err
	}
	defer response.Body.Close()

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return uri, err
	}

	log.Println("zip:", string(bs))

	var result struct {
		Error           int    `json:"error"`
		BrowserFilename string `json:"browser_filename"`
		PublicFilename  string `json:"public_filename"`
		DownloadUrl     string `json:"download_url"`
		FileSize        int    `json:"filesize"`
		Zipped          int    `json:"zipped"`
	}

	err = json.Unmarshal(bs, &result)
	if err != nil {
		return uri, err
	}

	return result.DownloadUrl, nil
}

const (
	probe_req  = "3probe"
	probe_resp = "3probe"

	start_req  = "5"
	heart_req  = "2"
	heart_resp = "3"

	sid_prefix = "0"
	cmd_prefix = "42"
)

const (
	mode_polling = "polling"
	mode_socket  = "socket"
)

type audio struct {
	BitrateType     string
	ConstantBitrate string
	VariableBitrate string
	SampleRate      string
	Channels        int
	Fadein          bool
	Fadeout         bool
	Reverse         bool
}

var config = map[string]audio{
	"mp3": {
		BitrateType:     "constant",
		ConstantBitrate: "128",
		SampleRate:      "44100",
		Channels:        2,
	},
	"wav": {
		BitrateType:     "constant", //
		ConstantBitrate: "160",      //
		SampleRate:      "44100",
		Channels:        2,
	},
	"m4r": {
		BitrateType:     "constant", //
		ConstantBitrate: "160",
		SampleRate:      "44100",
		Channels:        2,
	},
	"m4a": {
		BitrateType:     "constant", //
		ConstantBitrate: "160",
		SampleRate:      "44100",
		Channels:        2,
	},
	"flac": {
		BitrateType:     "constant", //
		ConstantBitrate: "160",      //
		SampleRate:      "48000",
		Channels:        2,
	},
	"ogg": {
		BitrateType:     "constant", //
		ConstantBitrate: "160",
		SampleRate:      "44100",
		Channels:        2,
	},
	"mp2": {
		BitrateType:     "constant", //
		ConstantBitrate: "160",
		SampleRate:      "44100",
		Channels:        2,
	},
	"amr": {
		BitrateType:     "constant", //
		ConstantBitrate: "12.20",
		SampleRate:      "8000", //
		Channels:        1,      //
	},
}

type Socket struct {
	*websocket.Conn
	sync.Mutex
	sid    string
	closed bool   // 是否已经关闭
	mode   string // 模式
	done   chan struct{}
	task   struct {
		job    chan string
		result chan struct {
			err error
			url string
		}
	}

	heart struct {
		ticker  *time.Ticker
		counter int64
	}
}

func (s *Socket) Close() {
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

// /socket.io/?
// EIO=3&
// transport=polling&
// t=Nlbxcde
func (s *Socket) polling() error {
	u := fmt.Sprintf("https://"+host+"/socket.io/?EIO=3&transport=polling&t=%v", tencode())
	request, _ := http.NewRequest("GET", u, nil)
	resonse, err := oclient.Do(request)
	if err != nil {
		return err
	}
	defer resonse.Body.Close()

	if resonse.StatusCode != http.StatusOK {
		return errors.New("invalid status code")
	}

	data, err := ioutil.ReadAll(resonse.Body)
	if err != nil {
		return err
	}

	data = regexp.MustCompile(`{.*}`).Find(data)

	var result struct {
		Sid          string   `json:"sid"`
		Upgrades     []string `json:"upgrades"`
		PingInterval int      `json:"pingInterval"`
		PingTimeout  int      `json:"pingTimeout"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	s.sid = result.Sid

	log.Printf("polling: %v", string(data))
	return nil
}

func (s *Socket) polling1() (result string, err error) {
	u := fmt.Sprintf("https://"+host+"/socket.io/?EIO=3&transport=polling&t=%v&sid=%v", tencode(), s.sid)
	request, _ := http.NewRequest("GET", u, nil)
	request.Header.Set("cookie", fmt.Sprintf("io=%v", s.sid))
	request.Header.Set("origin", "https://online-audio-converter.com")
	request.Header.Set("pragma", "no-cache")
	request.Header.Set("referer", "https://online-audio-converter.com/cn/")
	request.Header.Set("sec-fetch-dest", "empty")
	request.Header.Set("sec-fetch-mode", "cors")
	request.Header.Set("sec-fetch-site", "cross-site")
	resonse, err := oclient.Do(request)
	if err != nil {
		log.Println("polling1 Do:", err)
		return result, err
	}
	defer resonse.Body.Close()
	data, err := ioutil.ReadAll(resonse.Body)
	if err != nil {
		return result, err
	}

	log.Printf("polling1: %v", string(data))

	if resonse.StatusCode != http.StatusOK {
		log.Printf("polling1 StatusCode: %v", resonse.StatusCode)
		if resonse.StatusCode == http.StatusBadRequest {
			s.polling()
		}
		return
	}

	return string(data), nil
}

func (s *Socket) polling2(body string) error {
	u := fmt.Sprintf("https://"+host+"/socket.io/?EIO=3&transport=polling&t=%v&sid=%v", tencode(), s.sid)
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	request.Header.Set("cookie", "io="+s.sid)
	resonse, err := oclient.Do(request)
	if err != nil {
		log.Println("polling2 Do:", err)
		return err
	}
	defer resonse.Body.Close()
	data, err := ioutil.ReadAll(resonse.Body)
	if err != nil {
		log.Println("polling2 Read:", err)
		return err
	}

	log.Printf("polling2: %v", string(data))

	if resonse.StatusCode != http.StatusOK {
		log.Println("polling2 Code:", resonse.StatusCode)
		return fmt.Errorf("invalid status code: %v", resonse.StatusCode)
	}

	return nil
}

func (s *Socket) Socket() error {
	if s.mode == "" {
		s.mode = mode_socket
	}
	if s.mode != mode_socket {
		return fmt.Errorf("invalid mode")
	}

	u := fmt.Sprintf("wss://"+host+"/socket.io/?EIO=3&transport=websocket&sid=%v", s.sid)
	dailer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	var retryTimeout = time.Second

again:
	err := s.polling()
	if err != nil {
		return err
	}

	header := make(http.Header)
	conn, _, err := dailer.Dial(u, header)
	if err != nil && retryTimeout < 64*time.Second {
		log.Println("dailer.Dial", err)
		time.Sleep(retryTimeout)
		retryTimeout *= 2
		goto again
	}

	if err != nil {
		log.Println("dailer.Dial", err)
		return err
	}

	s.Conn = conn
	s.done = make(chan struct{})
	s.task.job = make(chan string)
	s.task.result = make(chan struct {
		err error
		url string
	})

	// start cmd
	if s.write(probe_req) {
		goto close
	}

	for {
		data, ok := s.read()
		if ok {
			goto close
		}

		switch data {
		case probe_resp:
			go s.heartbeat() // heartbeat after probe
			log.Println("probe success,", data)
		case heart_resp:
			atomic.AddInt64(&s.heart.counter, -1)
			log.Printf("receive heart: %v", atomic.LoadInt64(&s.heart.counter))
		default:
			if strings.HasPrefix(data, sid_prefix) {
				resolved := regexp.MustCompile(`{.*}`).FindString(data)
				var result struct {
					Sid          string   `json:"sid"`
					Upgrades     []string `json:"upgrades"`
					PingInterval int      `json:"pingInterval"`
					PingTimeout  int      `json:"pingTimeout"`
				}

				err = json.Unmarshal([]byte(resolved), &result)
				if err != nil {
					log.Println("cmd unmarshal", err, string(data))
					continue
				}

				s.sid = result.Sid

				go s.heartbeat() // heartbeat after probe
				continue
			}

			if strings.HasPrefix(data, cmd_prefix) {
				s.cmdDecode(data)
				continue
			}

			log.Println("origin data", data)
		}
	}

close:
	return errors.New("exception cloded")
}

func (s *Socket) SocketJob(src, format string) (url string, err error) {
	if s.mode != mode_socket {
		return url, fmt.Errorf("invalid job")
	}

	tempfile, err := Flow(src)
	if err != nil {
		return url, err
	}
	operationid := fmt.Sprintf("%v_%v", time.Now().UnixNano()/1e6, random(10))

	cmd := s.cmdEncode(tempfile, operationid, format)
	s.write(cmd)
	val := <-s.task.result

	return val.url, val.err
}

func (s *Socket) Poll() error {
	if s.mode == "" {
		s.mode = mode_polling
	}
	if s.mode != mode_polling {
		return fmt.Errorf("invalid mode")
	}

	err := s.polling()
	if err != nil {
		return err
	}
	s.polling2("1:2")
	s.polling1()

	s.done = make(chan struct{})
	s.task.job = make(chan string)
	s.task.result = make(chan struct {
		err error
		url string
	})

	for {
		var job string
		select {
		case <-s.done:
			return nil
		case val := <-s.task.job:
			job = val
			log.Println("job is comming....")
		default:
			job = "1:2"
		}

		s.polling2(job)

		result, err := s.polling1()
		if err != nil {
			log.Println("err:", err)
			continue
		}

		if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
			log.Println("heart cmd", result)
			continue
		}

		if strings.HasPrefix(result, "1:") {
			result = result[2:]
		}

		for len(result) > 3 {
			i := strings.Index(result, ":")
			length, _ := strconv.ParseInt(result[:i], 10, 64)
			data := result[i+1 : i+1+int(length)]
			s.cmdDecode(data)
			result = result[i+1+int(length):]
		}
	}
}

func (s *Socket) PollJob(src, format string) (url string, err error) {
	if s.mode != mode_polling {
		return url, fmt.Errorf("invalid job")
	}

	tempfile, err := Flow(src)
	if err != nil {
		return url, err
	}
	operationid := fmt.Sprintf("%v_%v", time.Now().UnixNano()/1e6, random(10))
	cmd := s.cmdEncode(tempfile, operationid, format)
	job := fmt.Sprintf("%v:%v", len(cmd), cmd)
	log.Println("job", job)
	s.task.job <- job
	val := <-s.task.result

	return val.url, val.err
}

func (s *Socket) read() (data string, isclose bool) {
	if s.closed {
		return data, s.closed
	}

	_, message, err := s.Conn.ReadMessage()
	if s.socketError(err) {
		return data, true
	}

	return string(message), false
}

func (s *Socket) write(data string) (isclose bool) {
	if s.closed {
		return s.closed
	}

	s.Lock()
	defer s.Unlock()
	err := s.Conn.WriteMessage(websocket.TextMessage, []byte(data))
	return s.socketError(err)
}

func (s *Socket) socketError(err error) (isclose bool) {
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

func (s *Socket) heartbeat() {
	log.Println("start heart beat")
	s.write(start_req)
	s.write(heart_req)
	atomic.AddInt64(&s.heart.counter, 1)
	log.Printf("heartbeat send: %v", atomic.LoadInt64(&s.heart.counter))
	s.heart.ticker = time.NewTicker(25 * time.Second)
	for {
		select {
		case <-s.heart.ticker.C:
			atomic.AddInt64(&s.heart.counter, 1)
			log.Printf("heartbeat send: %v", atomic.LoadInt64(&s.heart.counter))
			s.write(heart_req)

		case <-s.done:
			return
		}
	}
}

func (s *Socket) cmdEncode(tmpfilename, operationid, format string) string {
	log.Println("args", tmpfilename, operationid, format)
	var common = map[string]interface{}{
		"site_id":            "aconv",
		"uid":                uid,
		"user_id":            nil,
		"enable_user_system": false,
		"trackinfo": map[string]interface{}{
			"set_tag":           false,
			"track_tag_title":   "",
			"track_tag_artist":  "",
			"track_tag_album":   "",
			"track_tag_year":    "",
			"track_tag_genre":   "",
			"track_tag_comment": "",
		},
		"lang_id":  "cn",
		"host":     "online-audio-converter.com",
		"protocol": "https:",
	}

	common["tmp_filename"] = tmpfilename
	common["operation_id"] = operationid
	common["format"] = format // 目标格式, mp3,wav,m4a,flac,ogg,mp2,amr,m4r
	common["duration_in_seconds"] = 3
	common["preset"] = 2
	common["action_type"] = "encode"
	common["format_type"] = "audio"

	param := config[format]
	common["bitrate_type"] = param.BitrateType         // 比特率, constant, variable
	common["constant_bitrate"] = param.ConstantBitrate // 固定码率 32,40,48,56,64,80,96,112,128,160,192,224,256,320 kbps
	common["variable_bitrate"] = param.VariableBitrate // 可变码率 0,1,2,3,4,5,6,7,8,9
	common["sample_rate"] = param.SampleRate           // 采样率
	common["channels"] = param.Channels                // 通道 1,2
	common["fadein"] = param.Fadein                    // 淡入
	common["fadeout"] = param.Fadeout                  // 淡出
	common["reverse"] = param.Reverse                  // 倒放
	common["fastmode"] = false
	common["remove_voice"] = false
	common["preset_priority"] = false

	/*[
	    	"encode",
			{
				"site_id":"aconv",
				"uid":"WzOJjPokRKpPPgnJ9P85eeb1b4d3c035",
				"user_id":null,
				"operation_id":"1592467396827_fsanxfdbcp",
				"action_type":"encode",
				"enable_user_system":false,
				"format":"ogg",
				"preset":2,
				"format_type":"audio",
				"trackinfo":{
					"set_tag":false,
					"track_tag_title":"",
					"track_tag_artist":"",
					"track_tag_album":"",
					"track_tag_year":"",
					"track_tag_genre":"",
					"track_tag_comment":""
				},
				"bitrate_type":"constant",
				"constant_bitrate":"160",
				"variable_bitrate":"5",
				"sample_rate":"8000",
				"channels":"2",
				"fastmode":false,
				"fadein":true,
				"fadeout":true,
				"remove_voice":false,
				"reverse":true,
				"preset_priority":false,
				"tmp_filename":"s111RuUGertW.amr",
				"duration_in_seconds":3,
				"lang_id":"cn",
				"host":"online-audio-converter.com",
				"protocol":"https:"
			}
		]*/

	data, _ := json.Marshal(common)
	cmd := fmt.Sprintf(`%v["%v",%v]`, cmd_prefix, "encode", string(data))

	return cmd
}

const (
	step_handshake    = "handshake"
	step_progress     = "progress"
	step_final_result = "final_result"
)

func (s *Socket) cmdDecode(origin string) {
	data := strings.TrimPrefix(origin, cmd_prefix)
	var cmd [2]interface{}
	err := json.Unmarshal([]byte(data), &cmd)
	if err != nil {
		log.Println("invalid format", origin)
		return
	}

	realstu, ok := cmd[1].(map[string]interface{})
	if !ok {
		log.Println("invalid data", origin)
		return
	}

	switch realstu["message_type"] {
	case step_handshake:
		log.Printf("step:%v, pid:%d, operation_id:%v", step_handshake, int64(realstu["pid"].(float64)),
			realstu["operation_id"])
	case step_progress:
		log.Printf("step:%v, progress_value:%v, operation_id:%v", step_progress,
			realstu["progress_value"], realstu["operation_id"])
	case step_final_result:
		log.Printf("step:%v, success:%v, tmp_filename:%v, download_url:%v",
			step_final_result, realstu["success"], realstu["convertd"], realstu["download_url"])
		var (
			err error
			uri string
		)
		if realstu["success"].(bool) {
			uri = realstu["download_url"].(string)
		} else {
			err = fmt.Errorf("convert failed")
		}

		s.task.result <- struct {
			err error
			url string
		}{err: err, url: uri}
	}
}

///////////////////////////////////////////////////////////////////////////////

func tencode() string {
	var (
		s, u = 64, 0
	)

	a := map[string]int{}
	i := strings.Split("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_", "")
	for ; u < s; u++ {
		a[i[u]] = u
	}

	n := func(t int) string {
		var e string
		e = i[t%s] + e
		t = int(t / s)
		for t > 0 {
			e = i[t%s] + e
			t = int(t / s)
		}
		return e
	}

	return n(int(time.Now().UnixNano() / 1e6))
}

func tdecode(t string) int {
	var (
		s, u = 64, 0
	)

	a := map[string]int{}
	i := strings.Split("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_", "")
	for ; u < s; u++ {
		a[i[u]] = u
	}

	var e = 0
	for u = 0; u < len(t); u++ {
		e = e*s + a[string(t[u])]
	}

	return e
}

func random(length int) string {
	bs := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	result := make([]byte, 0, length)
	r := mrand.New(mrand.NewSource(time.Now().UnixNano())) // 产生随机数实例
	for i := 0; i < length; i++ {
		result = append(result, bs[r.Intn(len(bs))]) // 获取随机
	}
	return string(result)
}
