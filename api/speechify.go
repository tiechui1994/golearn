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
	"crypto/cipher"
	"crypto/aes"
	"crypto/md5"
	"encoding/gob"
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
	user_agent   = "aws-sdk-iOS/2.10.3 iOS/13.5.1 en_US"
	time_formate = "20060102T150405Z"
)

var (
	eu_west_1 string
	us_west_2 string
	resions   map[string]string

	ocr_key string
)

func init() {
	var private = map[string]string{}
	data, _ := ioutil.ReadFile("./data/key.data")
	plain, err := aesDecrypt(string(data), "")
	if err != nil {
		return
	}

	gob.NewDecoder(bytes.NewBuffer(plain)).Decode(&private)

	eu_west_1 = private["eu_west_1"]
	us_west_2 = private["us_west_2"]
	ocr_key = private["ocr_key"]

	resions = map[string]string{
		"eu-west-1": eu_west_1,
		"us-west-2": us_west_2,
	}
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

func (s *Speech) Speech(text, dest string) error {
	if _, ok := resions[s.Region]; !ok {
		return fmt.Errorf("invalid regions")
	}

	// OutputFormat: ogg_vorbis, json, mp3, pcm
	body := fmt.Sprintf(`{"VoiceId": "Zhiyu", "OutputFormat": "mp3", "Text": "%v"}`, text)
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

	log.Println("speech content-length", response.ContentLength)

	fd, err := os.Create(dest)
	if err != nil {
		log.Println("Create", err)
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
	request.Header.Set("ocp-apim-subscription-key", ocr_key)
	request.Header.Set("user-agent", "SpeechifyMobile/2.2.1 (com.cliffweitzman.speechifyMobile2; build:56918.5.26344858; iOS 13.5.1) Alamofire/4.8.2")

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

func MD5(msg string) string {
	sha := md5.New()
	sha.Write([]byte(msg))
	return hex.EncodeToString(sha.Sum(nil))
}

func Escape(data string) string {
	return strings.Replace(data, "/", "\\/", -1)
}

func aesEncrypt(msg, key string) (data []byte) {
	iv := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	// 创建加密算法aes
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Printf("Error: NewCipher(%d bytes) = %s", len(key), err)
		os.Exit(-1)
	}

	//加密字符串
	cfb := cipher.NewCFBEncrypter(c, iv)
	ciphertext := make([]byte, len(msg))
	cfb.XORKeyStream(ciphertext, []byte(msg))
	return ciphertext
}

func aesDecrypt(msg, key string) (data []byte, err error) {
	iv := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	// 创建加密算法aes
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Printf("Error: NewCipher(%d bytes) = %s", len(key), err)
		return data, err
	}

	//加密字符串
	cfb := cipher.NewCFBDecrypter(c, iv)
	plaintext := make([]byte, len(msg))
	cfb.XORKeyStream(plaintext, []byte(msg))
	return plaintext, nil
}
