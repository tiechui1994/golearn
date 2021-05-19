package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	agent = "Mozilla/5.0 (X11; Linux x86_64) Chrome/90.0.4430.93 Safari/537.36"
)

type CodeError int

func (err CodeError) Error() string {
	return http.StatusText(int(err))
}

func init() {
	jar, _ := cookiejar.New(nil)
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
}

func POST(u string, body io.Reader, header map[string]string) (raw json.RawMessage, err error) {
	request, _ := http.NewRequest("POST", u, body)
	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}
	request.Header.Set("user-agent", agent)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return raw, err
	}

	raw, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return raw, err
	}

	if response.StatusCode >= 400 {
		return raw, CodeError(response.StatusCode)
	}

	return raw, err
}

func GET(u string, header map[string]string) (raw json.RawMessage, err error) {
	request, _ := http.NewRequest("GET", u, nil)
	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}
	request.Header.Set("user-agent", agent)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return raw, err
	}

	raw, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return raw, err
	}

	if response.StatusCode >= 400 {
		return raw, CodeError(response.StatusCode)
	}

	return raw, err
}
