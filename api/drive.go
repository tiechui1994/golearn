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
)

/**
doc:
https://webapps.stackexchange.com/questions/126394/cant-download-large-file-from-google-drive-as-one-single-folder-always-splits
*/

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
	clientid     string
}

func New(accesstoken, refreshtoken string) *Config {
	return &Config{
		AccessToken:  accesstoken,
		RefreshToken: refreshtoken,
		clientid:     "407408718192.apps.googleusercontent.com",
	}
}

func (c *Config) refreshToken() error {
	values := make(url.Values)
	values.Set("client_id", c.clientid)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", c.RefreshToken)

	u := "https://www.googleapis.com/oauth2/v4/token"
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

	return nil
}
