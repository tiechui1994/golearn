package api

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"math"
	"encoding/json"
	"bytes"
	"net/http"
	"io/ioutil"
	"log"
	"fmt"
	"time"
	"strings"
	"crypto/sha1"
	"crypto/hmac"
	"encoding/base64"
	"encoding/xml"
	"crypto/md5"
	"sort"
)

const (
	PNG  = "png"
	BMP  = "bmp"
	JPEG = "jpeg"
)

func GetImageWidthAndHeight(picturePath string) (w, h int) {
	var err error
	_, err = os.Stat(picturePath)
	if err != nil {
		panic("the image path: " + picturePath + " not exist")
	}

	fd, err := os.Open(picturePath)
	if err != nil {
		panic("open image error")
	}

	config, _, err := image.DecodeConfig(fd)
	if err != nil {
		panic("decode image error")
	}

	return config.Width, config.Height
}

func DrawPNG(srcPath string) {
	const (
		width  = 300
		height = 300
	)

	// 文件
	pngFile, _ := os.Create(srcPath)
	defer pngFile.Close()

	// Image, 进行绘图操作
	pngImage := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(pngImage, pngImage.Bounds(), image.White, image.ZP, draw.Src)

	// 圆内旋转线
	a, b, c := 5.0, 7.0, 2.2
	for p := 0.0; p <= 1800.0; p += 0.03125 {
		x := int(35*((a-b)*math.Cos(p)+c*math.Cos((a/b-1)*p))) + 150
		y := int(35*((a-b)*math.Sin(p)-c*math.Sin((a/b-1)*p))) + 150

		pngImage.Set(x, y, color.RGBA{
			R: uint8(250),
			G: uint8(4),
			B: uint8(4),
			A: uint8(255),
		})
	}

	// 以 png 的格式写入文件
	png.Encode(pngFile, pngImage)
}

type Session struct {
	ApiToken string `json:"api_token"`
	IsNew    int    `json:"is_new"`
	Nickname string `json:"nickname"`
	Userid   string `json:"userid"`
}

type Auth struct {
	AccessId      string `json:"access_id"`
	AccessSecret  string `json:"access_secret"`
	Bucket        string `json:"bucket"`
	RegionId      string `json:"region_id"`
	ExpiresIn     int    `json:"expires_in"`
	SecurityToken string `json:"security_token"`
	Endpoint      string `json:"endpoint"`
	Path struct {
		Audios    string `json:"audios"`
		Images    string `json:"images"`
		Resources string `json:"resources"`
		Videos    string `json:"videos"`
	} `json:"path"`
	Callback struct {
		CallbackUrl  string `json:"callbackUrl"`
		CallbackBody string `json:"callbackBody"`
	} `json:"callback"`
}

type Resource struct {
	Bucket      string `json:"bucket"`
	Filename    string `json:"filename"`
	ResourceIFd string `json:"resource_id"`
	Url         string `json:"url"`
	Type        string `json:"type"`
}

func Sessions(u string) (session Session, err error) {
	request, _ := http.NewRequest("POST", u, nil)
	request.Header.Set("accept-type", "application/json")

	response, err := scleint.Do(request)
	if err != nil {
		return session, err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	log.Println("data", string(data))

	if response.StatusCode != http.StatusOK {
		return session, fmt.Errorf("invalid status: %v", response.StatusCode)
	}

	var result struct {
		Data struct {
			User Session `json:"user"`
		} `json:"data"`
		Status string `json:"status"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return session, err
	}

	if result.Status == "1" {
		return result.Data.User, nil
	}

	return session, fmt.Errorf("status:%v", result.Status)
}

func Authentications(token, u string) (auth Auth, err error) {
	request, _ := http.NewRequest("GET", u, nil)
	request.Header.Set("accept-type", "application/json")
	request.Header.Set("authorization", fmt.Sprintf("Bearer %v", token))

	response, err := scleint.Do(request)
	if err != nil {
		return auth, err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	log.Println("data", string(data))

	if response.StatusCode != http.StatusOK {
		return auth, fmt.Errorf("invalid status: %v", response.StatusCode)
	}

	var result struct {
		Data   Auth   `json:"data"`
		Status string `json:"status"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return auth, err
	}

	if result.Status == "1" {
		return result.Data, nil
	}

	return auth, fmt.Errorf("status:%v", result.Status)
}

const (
	oss_date       = "Mon, 02 Jan 2006 15:04:05 GMT"
	oss_user_agent = "aliyun-sdk-js/6.1.1 Chrome 86.0.4240.75 on Linux 64-bit"
)

func ApplyUploadId(image string, auth Auth) (uploadid string, err error) {
	u := "https://" + auth.Bucket + "." + auth.Endpoint + "/" + auth.Path.Images + image + "?uploads="
	request, _ := http.NewRequest("POST", u, nil)
	request.Header.Set("content-type", "image/jpeg")
	request.Header.Set("x-oss-date", time.Now().UTC().Format(oss_date))
	request.Header.Set("x-oss-security-token", auth.SecurityToken)
	request.Header.Set("x-oss-user-agent", oss_user_agent)

	resources := fmt.Sprintf("/%v/%v", auth.Bucket, auth.Path.Images) + image + "?uploads"
	sign := signature(request.Header, "POST", resources, auth)
	authentication := fmt.Sprintf("OSS %v:%v", auth.AccessId, sign)
	request.Header.Set("authorization", authentication)

	response, err := scleint.Do(request)
	if err != nil {
		return uploadid, err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	log.Println("data", string(data))

	if response.StatusCode != http.StatusOK {
		return uploadid, fmt.Errorf("invalid status: %v", response.StatusCode)
	}

	var result struct {
		Bucket   string
		Key      string
		UploadId string
	}

	err = xml.Unmarshal(data, &result)
	if err != nil {
		return uploadid, err
	}

	return result.UploadId, nil

}

func UploadData(image, uploadId, path string, auth Auth) (etag string, err error) {
	fd, err := os.Open(path)
	if err != nil {
		return etag, err
	}

	u := "https://" + auth.Bucket + "." + auth.Endpoint + "/" + auth.Path.Images + image + "?partNumber=1&uploadId=" + uploadId
	request, _ := http.NewRequest("PUT", u, fd)
	request.Header.Set("content-type", "image/jpeg")
	request.Header.Set("x-oss-date", time.Now().UTC().Format(oss_date))
	request.Header.Set("x-oss-security-token", auth.SecurityToken)
	request.Header.Set("x-oss-user-agent", oss_user_agent)

	resources := fmt.Sprintf("/%v/%v", auth.Bucket, auth.Path.Images) + image + "?partNumber=1&uploadId=" + uploadId
	sign := signature(request.Header, "PUT", resources, auth)
	authentication := fmt.Sprintf("OSS %v:%v", auth.AccessId, sign)
	request.Header.Set("authorization", authentication)

	response, err := scleint.Do(request)
	if err != nil {
		return etag, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return etag, fmt.Errorf("invalid status: %v", response.StatusCode)
	}

	return response.Header.Get("ETag"), nil
}

func GetUploadInfo(image, uploadId, eTag string, auth Auth) (r Resource, err error) {
	data := fmt.Sprintf(`
	<?xml version="1.0" encoding="UTF-8"?>
	<CompleteMultipartUpload>
	<Part>
	<PartNumber>1</PartNumber>
	<ETag>%v</ETag>
	</Part>
	</CompleteMultipartUpload>`, eTag)

	log.Println("data body", data)
	u := "https://" + auth.Bucket + "." + auth.Endpoint + "/" + auth.Path.Images + image + "?uploadId=" + uploadId
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(data))
	request.Header.Set("content-type", "application/xml")
	request.Header.Set("x-oss-date", time.Now().UTC().Format(oss_date))
	request.Header.Set("x-oss-security-token", auth.SecurityToken)
	request.Header.Set("x-oss-user-agent", oss_user_agent)
	request.Header.Set("content-md5", Md5(data))
	callback := fmt.Sprintf(`{"callbackUrl":"%v","callbackBody":"%v"}`,
		auth.Callback.CallbackUrl, auth.Callback.CallbackBody)
	request.Header.Set("x-oss-callback", base64.StdEncoding.EncodeToString([]byte(callback)))

	resources := fmt.Sprintf("/%v/%v", auth.Bucket, auth.Path.Images) + image + "?uploadId=" + uploadId
	sign := signature(request.Header, "POST", resources, auth)
	authentication := fmt.Sprintf("OSS %v:%v", auth.AccessId, sign)
	request.Header.Set("authorization", authentication)

	response, err := scleint.Do(request)
	if err != nil {
		return r, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(response.Body)
		log.Println("data", string(data))
		return r, fmt.Errorf("invalid status: %v", response.StatusCode)
	}
	var result struct {
		Data struct {
			Resource Resource `json:"resource"`
		} `json:"data"`
		Status string `json:"status"`
	}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return r, err
	}

	return result.Data.Resource, nil
}

func signature(header http.Header, method, resources string, auth Auth) string {
	ossHeaders := make([]string, 0)
	for k := range header {
		lower := strings.ToLower(k)
		if strings.HasPrefix(lower, "x-oss-") {
			ossHeaders = append(ossHeaders, lower+":"+header.Get(lower))
		}
	}

	sort.Strings(ossHeaders)

	sign := []string{
		method,
		header.Get("content-md5"),
		header.Get("content-type"),
		header.Get("x-oss-date"),
	}

	sign = append(sign, ossHeaders...)
	sign = append(sign, resources)

	signstr := strings.Join(sign, "\n")

	key := []byte(auth.AccessSecret)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(signstr))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Md5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

func Upload() {
	u1 := "https://apiwm.aoscdn.com/api/sessions"
	session1, err := Sessions(u1)
	if err != nil {
		log.Println("Sessions1", err)
	}
	u2 := "https://apimts.aoscdn.com/api/sessions"
	session2, err := Sessions(u2)
	if err != nil {
		log.Println("Sessions2", err)
	}

	u1 = "https://apiwm.aoscdn.com/api/authentications"
	auth, err := Authentications(session1.ApiToken, u1)
	if err != nil {
		u2 = "https://apimts.aoscdn.com/api/authentications"
		auth, err = Authentications(session2.ApiToken, u2)
	}

	if err != nil {
		log.Println("Authentications2", err)
		return
	}

	img := "12356.jpeg"
	path := "/home/quinn/Downloads/ai/11.jpg"
	uploadid, err := ApplyUploadId(img, auth)
	if err != nil {
		log.Println("ApplyUploadId", err)
		return
	}

	etag, err := UploadData(img, uploadid, path, auth)
	if err != nil {
		log.Println("UploadData", err)
		return
	}

	resource, err := GetUploadInfo(img, uploadid, etag, auth)
	if err != nil {
		log.Println("GetUploadInfo", err)
		return
	}

	log.Printf("data: %+v", resource)
}
