package api

import (
	"net/url"
	"net/http"
	"bytes"
	"time"
	"crypto/tls"
	"net"
	"io/ioutil"
	"log"
	"strings"
	"fmt"
	"encoding/json"

	"github.com/pborman/uuid"
)

/**
doc:
https://webapps.stackexchange.com/questions/126394/cant-download-large-file-from-google-drive-as-one-single-folder-always-splits
*/

const (
	clientid    = ""
	clientsecrt = ""
)

var dclient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Second*60)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type Config struct {
	AccessToken  string
	RefreshToken string
	Expired      time.Time
	clientid     string
	clientsecret string
	redirecturi  string
}

func New() *Config {
	return &Config{
		clientid:     clientid,
		clientsecret: clientsecrt,
		redirecturi:  "https://www.baidu.com",
	}
}

func (c *Config) refreshToken() error {
	values := make(url.Values)
	values.Set("client_id", c.clientid)
	values.Set("client_secret", c.clientsecret)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", c.RefreshToken)

	u := "https://oauth2.googleapis.com/token"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(values.Encode()))
	request.Header.Set("content-type", "application/x-www-form-urlencoded")
	response, err := dclient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Println("data:", string(data))

	var result struct {
		AccessToken  string `json:"access_token"`  //accesstoken
		RefreshToken string `json:"refresh_token"` //refreshtoken
		ExpiresIn    int64  `json:"expires_in"`    //accesstoken有效期
	}

	err = json.Unmarshal(data, &result)
	if nil != err {
		return err
	}

	c.AccessToken = result.AccessToken
	c.Expired = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

func (c *Config) getAccessToken(code string) (err error) {
	var values = make(url.Values)
	code, _ = url.PathUnescape(code)
	values.Set("code", code)
	values.Set("client_id", c.clientid)
	values.Set("client_secret", c.clientsecret)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", c.redirecturi)

	u := "https://oauth2.googleapis.com/token"
	request, err := http.NewRequest("POST", u, strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Host", "oauth2.googleapis.com")

	response, err := dclient.Do(request)
	if nil != err {
		log.Println("send request:", err)
		return err
	}

	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)
	if nil != err {
		return err
	}

	log.Println("data", string(respBody))

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid request status: %v", response.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`  //accesstoken
		RefreshToken string `json:"refresh_token"` //refreshtoken
		ExpiresIn    int64  `json:"expires_in"`    //accesstoken有效期
	}

	err = json.Unmarshal(respBody, &result)
	if nil != err {
		return err
	}

	c.AccessToken = result.AccessToken
	c.RefreshToken = result.RefreshToken
	c.Expired = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

func (c *Config) GetGoogleCode() string {
	var values = make(url.Values)
	values.Set("client_id", c.clientid)
	values.Set("scope", "https://www.googleapis.com/auth/drive.readonly https://www.googleapis.com/auth/drive.file")
	values.Set("access_type", "offline")
	values.Set("include_granted_scopes", "true")
	values.Set("prompt", "consent") // 新增
	values.Set("response_type", "code")
	values.Set("redirect_uri", c.redirecturi)
	values.Set("state", uuid.New())

	params := values.Encode()
	uRL := "https://accounts.google.com/o/oauth2/v2/auth?" + params
	log.Println("url:", uRL)

	return uRL
}

func (c *Config) Task() {
	delay := c.Expired.Sub(time.Now())
	timer := time.NewTimer(delay)
	for {
		select {
		case <-timer.C:
			c.refreshToken()
			delay = c.Expired.Sub(time.Now())
			timer.Reset(delay)
		}
	}
}
