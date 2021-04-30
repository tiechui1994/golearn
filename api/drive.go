package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

/**
doc:
https://webapps.stackexchange.com/questions/126394/cant-download-large-file-from-google-drive-as-one-single-folder-always-splits
*/

const (
	google = "https://developers.google.com"
)

type CodeError int

func (e CodeError) Error() string {
	return http.StatusText(int(e))
}

var dclient = &http.Client{
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
}

var config struct {
	AccessToken  string
	RefreshToken string
	Expired      time.Time
	tokenuri     string
}

func init() {
	config.tokenuri = "https://oauth2.googleapis.com/token"
}

func httpPost(u string, header map[string]string, body string) (raw json.RawMessage, err error) {
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	response, err := dclient.Do(request)
	if err != nil {
		return raw, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return raw, CodeError(response.StatusCode)
	}

	return ioutil.ReadAll(response.Body)
}

func buildAuthorizeUri() (uri string, err error) {
	var body struct {
		Scope        []string `json:"scope"`
		ResponseType string   `json:"response_type"`
		AuthUri      string   `json:"auth_uri"`
		Prompt       string   `json:"prompt"`
		AccessType   string   `json:"access_type"`
	}

	body.Scope = []string{"https://www.googleapis.com/auth/drive.readonly"}
	body.ResponseType = "code"
	body.AuthUri = "https://accounts.google.com/o/oauth2/v2/auth"
	body.Prompt = "consent"
	body.AccessType = "offline"

	bin, _ := json.Marshal(body)
	u := google + "/oauthplayground/buildAuthorizeUri"
	data, err := httpPost(u, nil, string(bin))
	if err != nil {
		log.Println("Authorize:", err)
		return uri, err
	}

	var result struct {
		AuthorizeUri string `json:"authorize_uri"`
		Success      bool   `json:"success"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return uri, err
	}

	if !result.Success {
		return uri, errors.New("BuildAuthorizeUri failed")
	}

	return result.AuthorizeUri, nil
}

func exchangeAuthCode(code string) error {
	var body struct {
		Code     string `json:"code"`
		TokenUri string `json:"token_uri"`
	}

	body.Code = code
	body.TokenUri = config.tokenuri
	bin, _ := json.Marshal(body)
	u := google + "/oauthplayground/exchangeAuthCode"

	data, err := httpPost(u, nil, string(bin))
	if err != nil {
		log.Println("AuthCode:", err)
		return err
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Success      bool   `json:"success"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	if !result.Success {
		return errors.New("ExchangeAuthCode failed")
	}

	config.AccessToken = result.AccessToken
	config.RefreshToken = result.RefreshToken
	config.Expired = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

func refreshAccessToken() error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
		TokenUri     string `json:"token_uri"`
	}

	body.RefreshToken = config.RefreshToken
	body.TokenUri = config.tokenuri
	bin, _ := json.Marshal(body)
	u := google + "/oauthplayground/refreshAccessToken"

	data, err := httpPost(u, nil, string(bin))
	if err != nil {
		log.Println("AuthCode:", err)
		return err
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Success      bool   `json:"success"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	if !result.Success {
		return errors.New("ExchangeAuthCode failed")
	}

	config.AccessToken = result.AccessToken
	config.RefreshToken = result.RefreshToken
	config.Expired = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	log.Println("AccessToken", config.AccessToken)
	log.Println("RefreshToken", config.RefreshToken)
	log.Println("Expired", config.Expired.Local())

	return nil
}

func main() {
	config.AccessToken = "ya29.a0AfH6SMDa5IOLecESbKeuVY5f_dPDIY2qEgTEgb9MGhcuKh2yre-r65Ty5eQSbPmjqxofpklTCEPcr38MW3NFU3PYm3kvGA22M6GLark-9Mu9aTO5-wYBxDDrYGgop5K2yzDXkJEAbXj0T-Bzs1G6fAMtTogN"
	config.RefreshToken = "1//04-FprfeMnoeMCgYIARAAGAQSNwF-L9IrfanDIAWX5zc55iitnj1Oq4nGHOdIks5FyXHTqZIDsJ9JHpxD-_zuSliprhmp_sq_o1U"
	config.Expired = time.Now().Add(3000 * time.Second)

	go func() {
		refreshAccessToken()
		tciker := time.NewTicker(3000 * time.Second)
		for {
			select {
			case <-tciker.C:
				refreshAccessToken()
			}
		}
	}()

	timer := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-timer.C:
			timer.Reset(30*time.Minute)
			cmd := fmt.Sprintf(`curl -C - -H 'Authorization: Bearer %v' \
			-o /media/user/data/iso/www/macOSX.iso \
			https://www.googleapis.com/drive/v3/files/18eeA54RApJf8Zt0M5oeNeXvX1VTQt7KC?alt=media`, config.AccessToken)
			log.Println(cmd)
			cm := exec.Command("bash", "-c", cmd)
			cm.Stdin = os.Stdin
			cm.Stdout = os.Stdout
			cm.Stderr = os.Stderr
			cm.Run()
		}
	}
}

type FileDownloader struct {
	fileSize int
	curPos   int

	url            string
	outputFileName string
	totalPart      int //下载线程
	outputDir      string
}

//NewFileDownloader .
func NewFileDownloader(url, fileName, outputDir string, totalPart int) *FileDownloader {
	if outputDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		outputDir = wd
	}

	return &FileDownloader{
		fileSize:       0,
		url:            url,
		outputFileName: fileName,
		outputDir:      outputDir,
		totalPart:      totalPart,
	}
}

func Main() {
	startTime := time.Now()
	var url string //下载文件的地址
	url = "https://download.jetbrains.com/go/goland-2020.2.2.dmg"
	downloader := NewFileDownloader(url, "", "", 10)
	if err := downloader.Run(5); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n 文件下载完成耗时: %f second\n", time.Now().Sub(startTime).Seconds())
}

func (d *FileDownloader) head() (int, error) {
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return 0, err
	}
	w, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0, err
	}
	if w.StatusCode >= 400 {
		return 0, errors.New(fmt.Sprintf("Can't process, response is %v", w.StatusCode))
	}

	// check: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Ranges
	if w.Header.Get("Accept-Ranges") != "bytes" {
		return 0, errors.New("server not support range send")
	}

	contentDisposition := w.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)

		if err != nil {
			panic(err)
		}
		d.outputFileName = params["filename"]
	} else {
		d.outputFileName = filepath.Base(w.Request.URL.Path)
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Length
	return strconv.Atoi(w.Header.Get("Content-Length"))
}

func (d *FileDownloader) Run(n int) error {
	if n < 0 {
		n = 10
	}

	totalSize, err := d.head()
	if err != nil {
		return err
	}

	d.fileSize = totalSize

	const frame = 2 * 1024 // 8K
	stepSize := n * frame

	path := filepath.Join(d.outputDir, d.outputFileName)
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()
	st, err := os.Create(path + ".st")
	if err != nil {
		return err
	}
	defer st.Close()

	data := make([]byte, stepSize)
	var wg sync.WaitGroup
	for pos := 0; pos < d.fileSize; pos += stepSize {
		count := n
		size := stepSize
		if count+stepSize >= d.fileSize {
			count = (d.fileSize-pos)/frame + 1
			if (d.fileSize-pos)%frame == 0 {
				count -= 1
			}
			size = d.fileSize - pos
		}

		wg.Add(count)
		for i := 0; i < count; i++ {
			idx := i
			go func(idx int) {
				defer wg.Done()
				start := idx * frame
				end := (idx + 1) * frame

				from := pos + idx*frame
				to := pos + (idx+1)*frame
				if to > d.fileSize {
					to = d.fileSize
				}
				err := d.downloadPart(data[start:end], from, to)
				if err != nil {
					log.Printf("download failed:[%v], [%v]", err, idx)
				}
			}(idx)
		}

		wg.Wait()

		n, err := fd.WriteAt(data[:size], int64(pos))
		return nil
		if err != nil {
			return err
		}

		if n != size {
			log.Println("Write faild")
			return err
		}

		fd.Sync()
	}

	return nil
}

// 下载分片
func (d FileDownloader) downloadPart(data []byte, from, to int) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}

	log.Printf("download from:%d to:%d", from, to)
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", from, to))
	w, err := dclient.Do(r)
	if err != nil {
		return err
	}
	defer w.Body.Close()

	if w.StatusCode >= http.StatusBadRequest {
		return errors.New(w.Status)
	}

	n, err := w.Body.Read(data)
	if err != nil {
		return err
	}

	if n != to-from {
		log.Println(n, to-from+1)
		return errors.New("shard failed")
	}

	return nil
}

func (d FileDownloader) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(method, d.url, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("User-Agent", "mojocn")
	return r, nil
}
