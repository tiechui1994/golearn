package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"
)

var (
	agents = []string{
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",

		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:79.0) Gecko/20100101 Firefox/79.0",
	}
)

type CodeError int

func (err CodeError) Error() string {
	return http.StatusText(int(err))
}

type entry struct {
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	Domain     string    `json:"domain"`
	Path       string    `json:"path"`
	SameSite   string    `json:"samesite"`
	Secure     bool      `json:"secure"`
	HttpOnly   bool      `json:"httponly"`
	Persistent bool      `json:"persistent"`
	HostOnly   bool      `json:"host_only"`
	Expires    time.Time `json:"expires"`
	Creation   time.Time `json:"creation"`
	LastAccess time.Time `json:"lastaccess"`
	SeqNum     uint64    `json:"seqnum"`
}

type Jar struct {
	PsList cookiejar.PublicSuffixList `json:"pslist"`

	// mu locks the remaining fields.
	Mu sync.Mutex `json:"mu"`

	// entries is a set of entries, keyed by their eTLD+1 and subkeyed by
	// their name/domain/path.
	Entries map[string]map[string]entry `json:"entries"`

	// nextSeqNum is the next sequence number assigned to a new cookie
	// created SetCookies.
	NextSeqNum uint64 `json:"nextseqnum"`
}

func Serialize(jar *cookiejar.Jar) {
	oldpath := filepath.Join(ConfDir, "."+AppName+".json")
	localjar := (*Jar)(unsafe.Pointer(jar))
	fd, _ := os.Create(oldpath)
	json.NewEncoder(fd).Encode(localjar)
	fd.Sync()

	os.Rename(oldpath, filepath.Join(ConfDir, AppName+".json"))
}

func UnSerialize() *Jar {
	var localjar Jar
	fd, _ := os.Open(filepath.Join(ConfDir, AppName+".json"))
	err := json.NewDecoder(fd).Decode(&localjar)
	if err != nil {
		return nil
	}

	return &localjar
}

var (
	Debug      = false
	CookieSync = make(chan struct{})
	UserAgent  string

	AppName string
	ConfDir string

	jar http.CookieJar
)

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/tmp"
	}

	ConfDir = filepath.Join(home, ".config", "tool")
	os.MkdirAll(ConfDir, 0775)

	UserAgent = agents[int(time.Now().Unix())%len(agents)]

	localjar := UnSerialize()
	if localjar != nil {
		jar = (*cookiejar.Jar)(unsafe.Pointer(localjar))
	} else {
		jar, _ = cookiejar.New(nil)
	}

	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 5 * time.Minute,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Jar: jar,
	}

	go func() {
		timer := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timer.C:
				cookjar := jar.(*cookiejar.Jar)
				Serialize(cookjar)
			case <-CookieSync:
				cookjar := jar.(*cookiejar.Jar)
				Serialize(cookjar)
			}
		}
	}()
}

func request(method, u string, body io.Reader, header map[string]string) (raw json.RawMessage, err error) {
	request, _ := http.NewRequest(method, u, body)
	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	if Debug {
		log.Println(method, request.URL.Path)
	}

	if len(jar.Cookies(request.URL)) != 0 {
		request.Header.Set("cookie", "")
	}
	request.Header.Set("user-agent", UserAgent)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return raw, err
	}

	raw, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return raw, err
	}

	if Debug && len(raw) > 0 {
		log.Println(method, request.URL.Path, "data", string(raw))
	}

	if response.StatusCode >= 400 {
		return raw, CodeError(response.StatusCode)
	}

	return raw, err
}

func POST(u string, body io.Reader, header map[string]string) (raw json.RawMessage, err error) {
	return request("POST", u, body, header)
}

func PUT(u string, body io.Reader, header map[string]string) (raw json.RawMessage, err error) {
	return request("PUT", u, body, header)
}

func GET(u string, header map[string]string) (raw json.RawMessage, err error) {
	return request("GET", u, nil, header)
}

func WriteFile(filepath string, data interface{}) error {
	fd, err := os.Create(filepath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(fd)
	return encoder.Encode(data)
}

func ReadFile(filepath string, data interface{}) error {
	fd, err := os.Open(filepath)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(fd)
	decoder.UseNumber()
	return decoder.Decode(data)
}
