package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const lightyear = "https://api.lightcloudai.com"

// n=1620826903465, r=v2.1, t=POST

// version="v2.1"
// ts=1620826903465
// method="POST"
func Signature(version, ts, method string) string {
	salt, _ := base64.StdEncoding.DecodeString("cG5oMkNMQTJNcFhFdkxkenNWbXE4RjJMR0t6VUhOOVc=")
	fmt.Println(string(salt))

	strs := []string{"v3", "LightYearApp", ts, method, string(salt)}
	data := []byte(strings.Join(strs, "|"))

	i, j := 0, len(data)-1
	for i <= j {
		data[i], data[j] = data[j], data[i]
		i++
		j--
	}

	md := md5.New()
	md.Write(data)
	sum := hex.EncodeToString(md.Sum(nil))

	md = md5.New()
	md.Write([]byte(sum))
	return hex.EncodeToString(md.Sum(nil))
}

func SignupOne(id, email string) error {
	ts := fmt.Sprintf("%v", time.Now().UnixNano()/1000000)
	u := lightyear + "/api/v3/auth/signup/one?" + fmt.Sprintf("version=v2.1&timestamp=%v&sig=%v", ts,
		Signature("v2.1", ts, "POST"))
	var body struct {
		DeviceId       string `json:"deviceId"`
		DeviceName     string `json:"deviceName"`
		DevicePlatform string `json:"devicePlatform"`
		Email          string `json:"email"`
	}
	body.DeviceId = id
	body.DeviceName = "web"
	body.DevicePlatform = "web"
	body.Email = email
	bin, _ := json.Marshal(body)

	request, _ := http.NewRequest("POST", u, bytes.NewBuffer(bin))
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("user-agent", "AppleWebKit/537.36 (KHTML, like Gecko)")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
		return err
	}

	var result struct {
		Data       interface{} `json:"data"`
		Messsage   string      `json:"messsage"`
		StatusCode int         `json:"statusCode"`
	}

	data, _ := ioutil.ReadAll(response.Body)

	log.Println(string(data))
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return err
	}

	if result.StatusCode != 200 {
		return errors.New(result.Messsage)
	}

	return nil
}

func SignupTwo(id, email, pwd, code string) error {
	ts := fmt.Sprintf("%v", time.Now().UnixNano()/1000000)
	u := lightyear + "/api/v3/auth/signup/two?" + fmt.Sprintf("version=v2.1&timestamp=%v&sig=%v", ts,
		Signature("v2.1", ts, "POST"))
	var body struct {
		DeviceId       string `json:"deviceId"`
		DeviceName     string `json:"deviceName"`
		DevicePlatform string `json:"devicePlatform"`
		Email          string `json:"email"`
		InviteCode     string `json:"invite_code"`
		Password       string `json:"password"`
		Referrer       string `json:"referrer"`
		VerifyCode     string `json:"verify_code"`
	}
	body.DeviceId = id
	body.DeviceName = "web"
	body.DevicePlatform = "web"
	body.Email = email
	body.Password = pwd
	body.VerifyCode = code
	bin, _ := json.Marshal(body)

	request, _ := http.NewRequest("POST", u, bytes.NewBuffer(bin))
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("user-agent", "AppleWebKit/537.36 (KHTML, like Gecko)")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
		return err
	}

	var result struct {
		Data       interface{} `json:"data"`
		Messsage   string      `json:"messsage"`
		StatusCode int         `json:"statusCode"`
	}

	data, _ := ioutil.ReadAll(response.Body)

	log.Println(string(data))
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return err
	}

	if result.StatusCode != 200 {
		return errors.New(result.Messsage)
	}

	return nil
}

func SingIn(id, email, pwd string) {
	ts := fmt.Sprintf("%v", time.Now().UnixNano()/1000000)
	u := lightyear + "/api/v3/auth/signin?" + fmt.Sprintf("version=v2.1&timestamp=%v&sig=%v", ts,
		Signature("v2.1", ts, "POST"))
	var body struct {
		DeviceId       string `json:"deviceId"`
		DeviceName     string `json:"deviceName"`
		DevicePlatform string `json:"devicePlatform"`
		Email          string `json:"email"`
		Password       string `json:"password"`
	}

	body.DeviceId = id
	body.DeviceName = "web"
	body.DevicePlatform = "web"
	body.Email = email
	body.Password = base64.StdEncoding.EncodeToString([]byte(pwd))
	bin, _ := json.Marshal(body)

	request, _ := http.NewRequest("POST", u, bytes.NewBuffer(bin))
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("user-agent", "AppleWebKit/537.36 (KHTML, like Gecko)")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}

	var result struct {
		Data       interface{} `json:"data"`
		Messsage   string      `json:"messsage"`
		StatusCode int         `json:"statusCode"`
	}

	data, _ := ioutil.ReadAll(response.Body)

	log.Println(string(data))
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return
	}

	if result.StatusCode != 200 {
		log.Println(result.StatusCode, result.Messsage)
		return
	}

	log.Println(result.Messsage, result.Data)
}

func main() {
	id := "129a716fc5efefdc8d1394707323b4c1"
	email := "oxwcybklburqjwf@supersave.net"

try:
	err := SignupOne(id, email)
	if err != nil {
		return
	}

	var code string
	fmt.Fscanf(os.Stdin, "%s", &code)
	err = SignupTwo(id, email, "0214.abc", code)
	if err != nil {
		goto try
	}
}
