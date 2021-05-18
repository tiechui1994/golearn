package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"sync"
	"time"
)

const teambition = "https://tcs.teambition.net"

var (
	escape = func() func(name string) string {
		replace := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
		return func(name string) string {
			return replace.Replace(name)
		}
	}()
)

func init() {
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
	}
}

func POST(u string, body io.Reader, header map[string]string) (raw json.RawMessage, err error) {
	request, _ := http.NewRequest("POST", u, body)
	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return raw, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return raw, errors.New(response.Status)
	}

	return ioutil.ReadAll(response.Body)
}

func Upload(token, path string) (url string, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return url, err
	}

	info, _ := fd.Stat()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	w.WriteField("name", info.Name())
	w.WriteField("type", "application/octet-stream")
	w.WriteField("size", fmt.Sprintf("%v", info.Size()))
	w.WriteField("lastModifiedDate", info.ModTime().Format("Mon, 02 Jan 2006 15:04:05 GMT+0800 (China Standard Time)"))

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escape("file"), escape(info.Name())))
	h.Set("Content-Type", "application/octet-stream")
	writer, _ := w.CreatePart(h)
	io.Copy(writer, fd)

	w.Close()

	u := teambition + "/upload"
	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  w.FormDataContentType(),
	}

	data, err := POST(u, &body, header)
	if err != nil {
		return url, err
	}

	var result struct {
		ChunkSize   int    `json:"chunkSize"`
		Chunks      int    `json:"chunks"`
		FileKey     string `json:"fileKey"`
		FileName    string `json:"fileName"`
		FileSize    int    `json:"fileSize"`
		Storage     string `json:"storage"`
		DownloadUrl string `json:"downloadUrl"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return url, err
	}

	return result.DownloadUrl, err
}

func UploadChunk(token, path string) (url string, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return url, err
	}

	info, _ := fd.Stat()

	u := teambition + "/upload/chunk"
	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
	bin := fmt.Sprintf(`{"fileName":"%v","fileSize":%v,"lastUpdated":"%v"}`,
		info.Name(), info.Size(), info.ModTime().Format("2006-01-02T15:04:05.00Z"))
	fmt.Println(bin)
	data, err := POST(u, bytes.NewBufferString(bin), header)
	if err != nil {
		return url, err
	}

	fmt.Println("chunk", string(data), err)

	var result struct {
		ChunkSize   int    `json:"chunkSize"`
		Chunks      int    `json:"chunks"`
		FileKey     string `json:"fileKey"`
		FileName    string `json:"fileName"`
		FileSize    int    `json:"fileSize"`
		Storage     string `json:"storage"`
		DownloadUrl string `json:"downloadUrl"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return url, err
	}

	var wg sync.WaitGroup
	wg.Add(result.Chunks)
	for i := 1; i <= result.Chunks; i++ {
		idx := i
		go func(idx int) {
			defer wg.Done()
			data := make([]byte, result.ChunkSize)
			n, _ := fd.ReadAt(data, int64((idx-1)*result.ChunkSize))
			data = data[:n]

			u := teambition + fmt.Sprintf("/upload/chunk/%v?chunk=%v&chunks=%v", result.FileKey, idx, result.Chunks)
			header := map[string]string{
				"Authorization": "Bearer " + token,
				"Content-Type":  "application/octet-stream",
			}
			data, err = POST(u, bytes.NewBuffer(data), header)
			if err != nil {
				fmt.Println("chunk", idx, err)
			}
		}(idx)
	}

	wg.Wait()

	u = teambition + fmt.Sprintf("/upload/chunk/%v", result.FileKey)
	header = map[string]string{
		"Content-Length": "0",
		"Authorization":  "Bearer " + token,
		"Content-Type":   "application/json",
	}
	data, err = POST(u, nil, header)

	fmt.Println("merge", string(data), err)

	if err != nil {
		return url, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return url, err
	}

	return result.DownloadUrl, err
}

func main() {
}
