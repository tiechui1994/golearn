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
	CHN_Man   = "Amy"
	CHN_Woman = "Zhiyu"

	US_Man   = "Simeon"
	US_Woman = "Emma"

	AUS_Man   = "Erix"
	AUS_Woman = "Chloe"

	UK_Man  = "Harry"
	UK_Woman = "Stephanie"
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
	request, _ := http.NewRequest("POST", u, bytes.NewBufferString(body))
	request.Header.Set("content-type", "application/x-amz-json-1.0")
	request.Header.Set("x-amz-date", now.UTC().Format(time_formate))
	request.Header.Set("user-agent", user_agent)
	request.Header.Set("x-amz-security-token", s.SessionToken)
	request.Header.Set("authorization", s.SpeechSigne("polly."+s.Region+".amazonaws.com", body, now))

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

func (s *Speech) SpeechSigne(host, body string, now time.Time) string {
	// basic params
	access_key := s.AceessKeyId
	secret_key := s.SecretKey
	region := s.Region

	timestamp := now.UTC().Format(time_formate)
	datestamp := now.UTC().Format("20060102")

	bucket := "v1"
	object_key := "speech"
	request_params := ""

	service := "polly"

	// params
	std_resource := "/" + bucket + "/" + object_key
	std_querystring := request_params

	headers := []string{
		"content-type:application/x-amz-json-1.0",
		"host:" + host,
		"user-agent:" + user_agent,
		"x-amz-date:" + timestamp,
		"x-amz-security-token:" + s.SessionToken,
		"",
	}
	std_headers := strings.Join(headers, "\n")

	signed_headers := "content-type;host;user-agent;x-amz-date;x-amz-security-token"

	payload_hash := Sha256(body)
	log.Println("payload_hash", payload_hash)

	std_request := strings.Join([]string{"POST", std_resource, std_querystring, std_headers,
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
