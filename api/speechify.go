package api

import (
	"fmt"
	"time"
	"net/http"
	"net"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"log"
	"crypto/hmac"
	"crypto/sha256"
	"strings"
	"encoding/hex"
	"os"
	"io"
	"mime/multipart"
	"net/url"

	"github.com/pborman/uuid"
)

var scleint = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Minute)
		},
	},
	Timeout: time.Minute,
}

// voiceid
const (
	// 中文
	CN_Man   = "Amy"
	CN_Woman = "Zhiyu"

	// 英语
	EN_USA_Man   = "Simeon"
	EN_USA_Woman = "Emma"

	EN_AUS_Man   = "Erix"
	EN_AUS_Woman = "Chloe"

	EN_UK_Man_1 = "Harry"
	EN_UK_Man_2 = "Child"
	EN_UK_Woman = "Stephanie"

	EN_Man   = "Dylan"
	EN_Woman = "Amrita"

	// 日语
	JPN_Man   = "Riku"
	JPN_Woman = "Rin"

	// 韩语
	KOR_Woman = "Seo-yun"

	// 法语
	FR_FRE_Man   = "Thadd"
	FR_FRE_Woman = "Adele"

	FR_CAN_Woman = "Annie"

	// 德语
	GE_Man   = "Frank"
	GE_Woman = "Andrea"

	// 俄语
	RU_Man   = "Vasily"
	RU_Woman = "Sofia"

	// 西班牙语
	SP_SPA_Man     = "Diego"
	SP_SPA_Woman_1 = "Conchita"
	SP_SPA_Woman_2 = "Mia"

	SP_USA_Man   = "Miguel"
	SP_USA_Woman = "Liliana"

	// 意大利语
	IT_Man   = "Matteo"
	IT_Woman = "Luna"
)

// language
const (
	Lan_zhs = "zh-Hans" // Chinese (Simplified)
	Lan_zht = "zh-Hant" // Chinese (Traditional)

	Lan_en = "en" // English
	Lan_fr = "fr" // French
	Lan_de = "de" // German
	Lan_el = "el" // Greek
	Lan_it = "it" // Italian
	Lan_ja = "ja" // Japanese
	Lan_ko = "ko" // Korean, 韩语
	Lan_ru = "ru" // Russian
	Lan_es = "es" // Spanish
)

const (
	eu_west_1 = "eu-west-1:2b30bfe4-8100-4a06-9503-72d408b480ed"
	us_west_2 = "us-west-2:11efc284-3bdf-41aa-8f0d-a0c8956ab1e6"

	user_agent   = "aws-sdk-iOS/2.10.3 iOS/13.5.1 en_US"
	time_formate = "20060102T150405Z"
)

var resions = map[string]string{
	"eu-west-1": eu_west_1,
	"us-west-2": us_west_2,
}

type Speech struct {
	Region       string
	AceessKeyId  string
	Expiration   uint64
	SecretKey    string
	SessionToken string
}

func (s *Speech) Identity() error {
	if _, ok := resions[s.Region]; !ok {
		return fmt.Errorf("invalid regions")
	}

	body := fmt.Sprintf(`{"IdentityId": "%v"}`, resions[s.Region])
	u := "https://cognito-identity." + s.Region + ".amazonaws.com"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	request.Header.Set("content-type", "application/x-amz-json-1.1")
	request.Header.Set("x-amz-target", "AWSCognitoIdentityService.GetCredentialsForIdentity")
	request.Header.Set("x-amz-date", time.Now().Format(time_formate))
	request.Header.Set("user-agent", user_agent)

	response, err := scleint.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	log.Println("data", string(data))

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status: %v", response.StatusCode)
	}

	var result struct {
		Credentials struct {
			AccessKeyId  string
			Expiration   json.Number
			SecretKey    string
			SessionToken string
		}
		IdentityId string
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}

	exp, _ := result.Credentials.Expiration.Float64()
	s.AceessKeyId = result.Credentials.AccessKeyId
	s.Expiration = uint64(exp)
	s.SecretKey = result.Credentials.SecretKey
	s.SessionToken = result.Credentials.SessionToken

	return nil
}

func (s *Speech) Speech(dest string) error {
	if _, ok := resions[s.Region]; !ok {
		return fmt.Errorf("invalid regions")
	}

	body := fmt.Sprintf(`{"VoiceId": "Zhiyu", "OutputFormat": "mp3", "Text": "123456789"}`)
	u := "https://polly." + s.Region + ".amazonaws.com/v1/speech"

	now := time.Now()
	param := signparam{
		method:         "POST",
		host:           "polly." + s.Region + ".amazonaws.com",
		bucket:         "v1",
		object_key:     "speech",
		request_params: "",
		paload:         body,

		service: "polly",

		signheadername: []string{"content-type", "host", "user-agent", "x-amz-date", "x-amz-security-token"},
		signheadervals: map[string]string{
			"content-type":         "application/x-amz-json-1.0",
			"host":                 "polly." + s.Region + ".amazonaws.com",
			"user-agent":           user_agent,
			"x-amz-date":           now.UTC().Format(time_formate),
			"x-amz-security-token": s.SessionToken,
		},
	}

	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	request.Header.Set("content-type", "application/x-amz-json-1.0")
	request.Header.Set("x-amz-date", now.UTC().Format(time_formate))
	request.Header.Set("user-agent", user_agent)
	request.Header.Set("x-amz-security-token", s.SessionToken)
	request.Header.Set("authorization", s.Signature(param, now))

	response, err := scleint.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	log.Println("speech len", response.ContentLength)

	fd, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(fd, response.Body)
	return err
}

func (s *Speech) TextSplit() error {
	if _, ok := resions[s.Region]; !ok {
		return fmt.Errorf("invalid regions")
	}

	body := fmt.Sprintf(`{"VoiceId": "Zhiyu", "OutputFormat": "json", "SpeechMarkTypes":["word"],
		"Text": "中国, 加油"}`)
	u := "https://polly." + s.Region + ".amazonaws.com/v1/speech"

	now := time.Now()
	param := signparam{
		method:         "POST",
		host:           "polly." + s.Region + ".amazonaws.com",
		bucket:         "v1",
		object_key:     "speech",
		request_params: "",
		paload:         body,

		service: "polly",

		signheadername: []string{"content-type", "host", "user-agent", "x-amz-date", "x-amz-security-token"},
		signheadervals: map[string]string{
			"content-type":         "application/x-amz-json-1.0",
			"host":                 "polly." + s.Region + ".amazonaws.com",
			"user-agent":           user_agent,
			"x-amz-date":           now.UTC().Format(time_formate),
			"x-amz-security-token": s.SessionToken,
		},
	}

	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	request.Header.Set("content-type", "application/x-amz-json-1.0")
	request.Header.Set("x-amz-date", now.UTC().Format(time_formate))
	request.Header.Set("user-agent", user_agent)
	request.Header.Set("x-amz-security-token", s.SessionToken)
	request.Header.Set("authorization", s.Signature(param, now))

	response, err := scleint.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	log.Println("text info", string(data))

	return nil
}

// upload pdf
func ConvertPDFToText(src string) (msg string, err error) {
	token, err := getSecureToken()
	if err != nil {
		return msg, err
	}

	uid := strings.ToUpper(uuid.New())
	log.Println("uuid", uid)
	u, err := applayUrl(uid, token.AccessToken, src)
	if err != nil {
		log.Println("applayUrl err", err)
		return msg, err
	}

	log.Println("upload uri", u)

	_, err = uploadToGoogle(u, src)
	//if err != nil {
	//	log.Println("uploadToGoogle err", err)
	//	return msg, err
	//}

	waitGoogleConvert(uid)

	return msg, nil
}

const (
	google_key = "AIzaSyDGXMWz_oHUC8A6z1K6ge5n0PZ8LUMkgl4"
)

type tokeninfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

type uploadinfo struct {
	Name               string `json:"name"`
	Bucket             string `json:"bucket"`
	Generation         string `json:"generation"`
	Metageneration     string `json:"metageneration"`
	ContentType        string `json:"contentType"`
	StorageClass       string `json:"storageClass"`
	Size               string `json:"size"`
	Md5Hash            string `json:"md5Hash"`
	ContentEncoding    string `json:"contentEncoding"`
	ContentDisposition string `json:"contentDisposition"`
	Etag               string `json:"etag"`
	DownloadTokens     string `json:"downloadTokens"`
}

// 1 get token
func getSecureToken() (token tokeninfo, err error) {
	body := `{
		"grantType":"refresh_token",
		"refresh_token":"AE0u-NfPxMI2g12FPjPk9Tf_gtakLHi4KOsc2aThtRxdRN9kgBcNWmPKTmo_jBy8cjhZCHlk1E64ElqV8DSZfSw7szl05f58zrO_wQ36bX2GtC6juYhghACSqf0in4k9euTo7IPty3MLsJUNVC7rpx3BMEN06okf1fIoCwFMFw4AyYpWe-PIoPa1d8G9RAaoNq-gE1JsM2wz"
	}`
	u := "https://securetoken.googleapis.com/v1/token?key=" + google_key
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))

	request.Header.Set("content-type", "application/json")
	request.Header.Set("user-agent", user_agent)

	response, err := scleint.Do(request)
	if err != nil {
		return token, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return token, err
	}

	log.Println("data", string(data))

	err = json.Unmarshal(data, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}

// 1. applay
func applayUrl(uid, token, src string) (uploadurl string, err error) {
	name := fmt.Sprintf("pdfDocuments/%v.pdf", uid)
	endpoint := "https://firebasestorage.googleapis.com/v0/b/speechifymobile.appspot.com/o/"

	params := fmt.Sprintf("uploadType=%v&name=%v", "resumable", url.PathEscape(name))
	u := endpoint + url.PathEscape(name) + "?" + params
	body := fmt.Sprintf(`{"contentType":"%v","name":"%v"}`, Escape("application/pdf"), Escape(name))

	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	log.Println("body", body)
	log.Println("u", u)

	stats, err := os.Stat(src)
	if err != nil {
		return uploadurl, err
	}

	request.Header.Set("accept", "*/*")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("user-agent", user_agent)
	request.Header.Set("x-goog-upload-command", "start")
	request.Header.Set("x-goog-upload-content-type", "application/pdf")
	request.Header.Set("authorization", "Firebase "+token)
	request.Header.Set("x-goog-upload-content-length", fmt.Sprintf("%v", stats.Size()))

	response, err := scleint.Do(request)
	if err != nil {
		return uploadurl, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return uploadurl, fmt.Errorf("invalid code: %v", response.StatusCode)
	}

	log.Println("header", response.Header)

	xGuploaderUploadId := response.Header.Get("x-guploader-uploadid")
	//xGoogUploadStatus := response.Header.Get("x-goog-upload-status")
	//xGoogUploadControlUrl := response.Header.Get("x-goog-upload-control-url")
	//xGoogUploadUrl := response.Header.Get("x-goog-upload-url")

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return uploadurl, err
	}

	log.Println("upload data", string(data))

	var info uploadinfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return uploadurl, err
	}

	params = fmt.Sprintf("%v&upload_id=%v&upload_protocol=%v", params, xGuploaderUploadId, "resumable")
	uploadurl = endpoint + url.PathEscape(name) + "?" + params

	return uploadurl, nil
}

// 2. upload
func uploadToGoogle(uploadurl, src string) (info uploadinfo, err error) {
	fd, err := os.Open(src)
	if err != nil {
		return info, err
	}

	request, _ := http.NewRequest("PUT", uploadurl, fd)

	up, _ := url.ParseQuery(uploadurl)
	request.Header.Set("content-type", "application/octet-stream")
	request.Header.Set("user-agent", "com.cliffweitzman.speechifyMobile2/2.2.1 iPhone/13.5.1 hw/iPhone9_1 (GTMSUF/1)")
	request.Header.Set("x-goog-upload-protocol", up.Get("upload_protocol"))
	request.Header.Set("x-goog-upload-offsete", "0")
	request.Header.Set("x-goog-upload-command", "upload, finalize")

	response, err := scleint.Do(request)
	if err != nil {
		return info, err
	}
	defer response.Body.Close()

	log.Println("code", response.StatusCode)

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return info, err
	}

	log.Println("upload data", string(data))

	err = json.Unmarshal(data, &info)
	return info, err
}

// 3. vision
func waitGoogleConvert(uid string) (ok bool, err error) {
	acccesstoken, err := getAccessToken()
	if err != nil {
		return ok, err
	}
	acccesstoken = "ya29.c.Ko8B0Ad3FEClEzP7oTcKv-D8-BRgmy8w_zLcDxXmXJk2sMVV2c-0uR9ID2Tm6Rnay-umB-pbGfIqDM7TtDOTmBU7ofIB2NWrl8nUh29Z0HtLERnrnymuapktroygffa6iAGuYSo-MLBTN2_3FWD08ODulCpVbxxAJBzTBh3Vq4FOd4Huooe6poK6OxHQvAoyhRc"
	uri, err := getVisionUri(uid, acccesstoken)
	if err != nil {
		log.Println("getVisionUri", err)
		return ok, err
	}

	log.Println("uri is", uri)

	for {
		config, err := isStoped(uri, acccesstoken)
		if err != nil {
			return ok, err
		}
		if config.Done {
			ok = true
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return ok, nil
}

func getAccessToken() (token string, err error) {
	params := make(url.Values)
	params.Set("assertion", "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE1OTI5ODg1NjgsImV4cCI6MTU5Mjk5MjE2OCwiYXVkIjoiaHR0cHM6XC9cL3d3dy5nb29nbGVhcGlzLmNvbVwvb2F1dGgyXC92NFwvdG9rZW4iLCJzY29wZSI6Imh0dHBzOlwvXC93d3cuZ29vZ2xlYXBpcy5jb21cL2F1dGhcL2Nsb3VkLXBsYXRmb3JtIiwiaXNzIjoic3BlZWNoaWZ5LXByb2R1Y3Rpb24tYWNjb3VudEBzcGVlY2hpZnltb2JpbGUuaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20ifQ.Q9rbQgsE6iKeRJ5_NoVpICDhwtjf7uNq9Tr543FGpoRlHD0aEcaAXP34oTNd4vcppmzSZhcqc2rxFZ4dl7XaI8lQSzdNIVrBhUZ2NfYVMEPsEtdYKpRp5IyrkOCe1rZQQIFzSFyPeDQuil9iObhL48ErU31vvG7BER6MYIv_pw2rtv1ruQstWcJo6Uwh886byygXSwUr6Nm68J39F4gDEfBOSyhPco7dY_byGRQIj0Mtf3QUrzj0viobL2Rz1rgM_xq8ZB6ILnqPYJkJuxMUHwssi5nqbSQgs9lvInywS-wjtRxlkwXmbn4Vb8_DOkl8HTvn_uVapCZC5tlc1-OaDw")
	params.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")

	u := "https://www.googleapis.com/oauth2/v4/token"
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(params.Encode()))

	request.Header.Set("content-type", "application/x-www-form-urlencoded")
	request.Header.Set("user-agent", user_agent)

	response, err := scleint.Do(request)
	if err != nil {
		return token, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return token, err
	}

	log.Println("data", response.StatusCode, string(data))

	var result struct {
		AccessToken string `json:"access_token"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return token, err
	}

	return result.AccessToken, nil
}

func getVisionUri(uid, token string) (uri string, err error) {
	input := fmt.Sprintf("gs://pdf-speechifymobile/extraction/%v/", uid)
	output := fmt.Sprintf("gs://speechifymobile.appspot.com/pdfDocuments/%v.pdf", uid)
	body := fmt.Sprintf(`{
		"requests":[{
			"outputConfig":{
				"batchSize":100,
				"gcsDestination":{
					"uri":"%v"
				}
			},
			"inputConfig":{
				"gcsSource":{
					"uri":"%v"
				},
				"mimeType":"application\/pdf"
			},
			"features":[{"type":"DOCUMENT_TEXT_DETECTION"}]
		}]
	}`, Escape(input), Escape(output))

	u := "https://vision.googleapis.com/v1/files:asyncBatchAnnotate?prettyPrint=false&key=" + google_key
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))

	request.Header.Set("content-type", "application/json")
	request.Header.Set("user-agent", user_agent)
	request.Header.Set("authorization", "Bearer "+token)
	request.Header.Set("accept", "application/json")

	response, err := scleint.Do(request)
	if err != nil {
		return uri, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return uri, err
	}

	log.Println("getVisionUri", string(data))

	var result struct {
		Name string `json:"name"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return uri, err
	}

	return result.Name, nil
}

type visionresult struct {
	Name string `json:"name"`
	Metadata struct {
		State      string `json:"state"`
		CreateTime string `json:"createTime"`
		UpdateTime string `json:"updateTime"`
	} `json:"metadata"`
	Done bool `json:"done"`
	Response struct {
		Responses []struct {
			OutputConfig struct {
				GcsDestination struct {
					Uri string `json:"uri"`
				} `json:"gcsDestination"`
				BatchSize int `json:"batchSize"`
			} `json:"outputConfig"`
		} `json:"responses"`
	} `json:"response"`
}

func isStoped(uri, token string) (config visionresult, err error) {
	u := "https://vision.googleapis.com/v1/" + uri
	request, _ := http.NewRequest("GET", u, nil)

	request.Header.Set("user-agent", user_agent)
	request.Header.Set("authorization", "Bearer "+token)

	/*
	{
	"name": "projects/speechifymobile/operations/43bff3dff849cd17",
	"metadata": {
		"@type": "type.googleapis.com/google.cloud.vision.v1.OperationMetadata",
		"state": "DONE",
		"createTime": "2020-06-24T08:49:32.067379449Z",
		"updateTime": "2020-06-24T08:49:40.891103406Z"
	},
	"done": true,
	"response": {
		"@type": "type.googleapis.com/google.cloud.vision.v1.AsyncBatchAnnotateFilesResponse",
		"responses": [{
			"outputConfig": {
				"gcsDestination": {
					"uri": "gs://pdf-speechifymobile/extraction/35C6582F-14B9-428B-84FF-23B55191A42A/"
				},
				"batchSize": 100
			}
		}]
	}
	}
	*/

	response, err := scleint.Do(request)
	if err != nil {
		return config, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return config, err
	}

	log.Println("data", string(data))

	err = json.Unmarshal(data, &config)
	return config, err
}

// 4. getetxt
func getPDFText() {

}

// Microsoft OCR
func OCR(src, lan string) (msg string, err error) {
	fd, err := os.Open(src)
	if err != nil {
		return msg, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	form, err := writer.CreateFormFile("file", fd.Name())
	if err != nil {
		return msg, err
	}
	_, err = io.Copy(form, fd)
	if err != nil {
		return msg, err
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	u := "https://westus.api.cognitive.microsoft.com/vision/v2.0/ocr?detectOrientation=1&language=" + lan
	request, _ := http.NewRequest("POST", u, &body)
	request.Header.Set("content-type", contentType)
	request.Header.Set("ocp-apim-subscription-key", "6bcc78216edd472eb4b2c902d1f2e6a0")
	request.Header.Set("user-agent", "SpeechifyMobile/2.2.1 (com.cliffweitzman.speechifyMobile2; build:56918.5.26344858; iOS 13.5.1) Alamofire/4.8.2")

	/*
	Host: westus.api.cognitive.microsoft.com
	Content-Type: multipart/form-data; boundary=alamofire.boundary.72cf606a58ffdeeb
	Ocp-Apim-Subscription-Key: 6bcc78216edd472eb4b2c902d1f2e6a0
	User-Agent: SpeechifyMobile/2.2.1 (com.cliffweitzman.speechifyMobile2; build:56918.5.26344858; iOS 13.5.1) Alamofire/4.8.2
	*/

	response, err := scleint.Do(request)
	if err != nil {
		return msg, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return msg, err
	}

	log.Println("ocr info", string(data))

	var result struct {
		Language    string  `json:"language"`
		TextAngle   float64 `json:"textAngle"`
		Orientation string  `json:"orientation"`
		Regions []struct {
			BoundingBox string `json:"boundingBox"`
			Lines []struct {
				BoundingBox string `json:"boundingBox"`
				Words []struct {
					BoundingBox string `json:"boundingBox"`
					Text        string `json:"text"`
				} `json:"words"`
			} `json:"lines"`
		} `json:"regions"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return msg, err
	}

	for _, region := range result.Regions {
		for _, line := range region.Lines {
			for _, word := range line.Words {
				msg += " " + word.Text
			}
			msg += "\n"
		}
	}

	log.Println("msg\n", msg)

	return msg, nil
}

type signparam struct {
	method         string
	host           string
	bucket         string
	object_key     string
	request_params string
	paload         string

	service string

	signheadername []string
	signheadervals map[string]string
}

func (s *Speech) Signature(param signparam, now time.Time) string {
	// basic params
	access_key := s.AceessKeyId
	secret_key := s.SecretKey
	region := s.Region

	timestamp := now.UTC().Format(time_formate)
	datestamp := now.UTC().Format("20060102")

	bucket := param.bucket
	object_key := param.object_key
	request_params := param.request_params

	service := param.service

	// params
	std_resource := "/" + bucket + "/" + object_key
	std_querystring := request_params

	headers := make([]string, 0, len(param.signheadername)+1)
	for _, v := range param.signheadername {
		headers = append(headers, v+":"+param.signheadervals[v])
	}
	headers = append(headers, "")
	std_headers := strings.Join(headers, "\n")

	signed_headers := strings.Join(param.signheadername, ";")

	payload_hash := Sha256(param.paload)
	log.Println("payload_hash", payload_hash)

	std_request := strings.Join([]string{param.method, std_resource, std_querystring, std_headers,
		signed_headers, payload_hash}, "\n")

	// ssemble string-to-sign
	hash_alg := "AWS4-HMAC-SHA256"
	credential_scope := datestamp + "/" + region + "/" + service + "/" + "aws4_request"

	msg_tokens := []string{hash_alg, timestamp, credential_scope, Sha256(std_request)}
	msg := strings.Join(msg_tokens, "\n")
	key := signatureKey(secret_key, datestamp, region, service)

	signature := hex.EncodeToString([]byte(Hash(key, msg)))
	log.Println("signature", signature)

	v4auth_header := hash_alg + " " +
		"Credential=" + access_key + "/" + credential_scope + ", " +
		"SignedHeaders=" + signed_headers + ", " +
		"Signature=" + signature

	return v4auth_header
}

func signatureKey(key, date, region, service string) string {
	keyDate := Hash("AWS4"+key, date)
	keyRegion := Hash(keyDate, region)
	keyService := Hash(keyRegion, service)
	keySigining := Hash(keyService, "aws4_request")
	return keySigining
}

func Hash(key, msg string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	return string(h.Sum(nil))
}

func Sha256(msg string) string {
	sha := sha256.New()
	sha.Write([]byte(msg))
	return hex.EncodeToString(sha.Sum(nil))
}

func Escape(data string) string {
	return strings.Replace(data, "/", "\\/", -1)
}
