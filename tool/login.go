package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// "pkcs1_oaep"
type OAEP struct {
	alg    string
	hash   hash.Hash
	pubkey string
	prikey string
}

func (o *OAEP) init() {
	switch o.alg {
	case "sha1":
		o.hash = sha1.New()
	case "sha256":
		o.hash = sha256.New()
	case "md5":
		o.hash = md5.New()
	}
}

func (o *OAEP) Encrypt(msg []byte) string {
	block, _ := pem.Decode([]byte(o.pubkey))
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println("Load public key error")
		panic(err)
	}

	encrypted, err := rsa.EncryptOAEP(o.hash, rand.Reader, pub.(*rsa.PublicKey), msg, nil)
	if err != nil {
		fmt.Println("Encrypt data error")
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(encrypted)
}

func (o *OAEP) Decrypt(encrypted string) []byte {
	block, _ := pem.Decode([]byte(o.prikey))
	var pri *rsa.PrivateKey
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("Load private key error")
		panic(err)
	}

	decodedData, err := base64.StdEncoding.DecodeString(encrypted)
	ciphertext, err := rsa.DecryptOAEP(o.hash, rand.Reader, pri, decodedData, nil)
	if err != nil {
		fmt.Println("Decrypt data error")
		panic(err)
	}

	return ciphertext
}

func GenKey() (private, public string) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		os.Exit(1)
	}

	var writer bytes.Buffer

	// RSA PRIVATE KEY
	pem.Encode(&writer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	private = writer.String()

	writer.Reset()

	// PUBLIC KEY
	block := &pem.Block{
		Type: "PUBLIC KEY",
	}
	block.Bytes, _ = x509.MarshalPKIXPublicKey(&key.PublicKey)
	pem.Encode(&writer, block)
	public = writer.String()

	return private, public
}

func Login(clientid, pubkey, token, email, phone, pwd string) (access string, err error) {
	key, _ := base64.StdEncoding.DecodeString(pubkey)
	oeap := OAEP{
		pubkey: string(key),
		alg:    "sha1",
	}
	oeap.init()

	password := oeap.Encrypt([]byte(pwd))

	body := map[string]string{
		"password":      password,
		"client_id":     clientid,
		"response_type": "session",
		"publicKey":     pubkey,
		"token":         token,
	}

	var u string
	if email != "" {
		u = "/api/login/email"
		body["email"] = email
	} else {
		u = "/api/login/phone"
		body["phone"] = phone
	}

	u = "https://account.teambition.com" + u
	var writer bytes.Buffer
	json.NewEncoder(&writer).Encode(body)
	header := map[string]string{
		"content-type": "application/json",
	}
	fmt.Println(writer.String())
	data, err := POST(u, &writer, header)
	if err != nil {
		if _, ok := err.(CodeError); ok {
			log.Println("login", string(data))
		}
		return access, err
	}

	var result struct {
		AbnormalLogin      string `json:"abnormalLogin"`
		HasGoogleTwoFactor bool   `json:"hasGoogleTwoFactor"`
		Token              string `json:"token"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return access, err
	}

	if result.HasGoogleTwoFactor {
		err = TwoFactor(clientid, token, result.Token)
		if err != nil {
			return access, err
		}
	}

	return result.Token, nil
}

func TwoFactor(clientid, token, verify string) error {
	var code string
	fmt.Printf("Input Auth Code:")
	fmt.Scanf("%v", &code)

	body := map[string]string{
		"authcode":      code,
		"client_id":     clientid,
		"response_type": "session",
		"token":         token,
		"verify":        verify,
	}

	u := "https://account.teambition.com/api/login/two-factor"
	var writer bytes.Buffer
	json.NewEncoder(&writer).Encode(body)
	header := map[string]string{
		"content-type": "application/json",
	}
	data, err := POST(u, &writer, header)
	if err != nil {
		if _, ok := err.(CodeError); ok {
			log.Println("two-factor", string(data))
			return err
		}

		return err
	}

	return nil
}

func main() {
	DEBUG = true
	data, err := GET("https://account.teambition.com/login", nil)
	if err != nil {
		return
	}

	reaccout := regexp.MustCompile(`<script id="accounts-config" type="application/json">(.*?)</script>`)
	republic := regexp.MustCompile(`<script id="accounts-ssr-props" type="application/react-ssr-props">(.*?)</script>`)

	str := strings.Replace(string(data), "\n", "", -1)
	a := reaccout.FindAllStringSubmatch(str, 1)
	p := republic.FindAllStringSubmatch(str, 1)

	var config struct {
		TOKEN     string
		CLIENT_ID string
	}

	json.Unmarshal([]byte(a[0][1]), &config)

	var public struct {
		Fsm struct {
			Config struct {
				Pub struct {
					Algorithm string `json:"algorithm"`
					PublicKey string `json:"publicKey"`
				} `json:"pub"`
			} `json:"config"`
		} `json:"fsm"`
	}

	pub, _ := url.QueryUnescape(p[0][1])
	json.Unmarshal([]byte(pub), &public)

	Login(config.CLIENT_ID, public.Fsm.Config.Pub.PublicKey, config.TOKEN, "tiechui1994@163.com", "", "0214Abcd")
}
