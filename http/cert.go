package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func X509Parse(data []byte) {
	// -----BEGIN RSA PRIVATE KEY-----
	_, err := x509.ParsePKCS1PrivateKey(data)
	if err == nil {
		fmt.Println("ParsePKCS1PrivateKey")
	}

	//  -----BEGIN RSA PUBLIC KEY-----
	_, err = x509.ParsePKCS1PublicKey(data)
	if err == nil {
		fmt.Println("ParsePKCS1PublicKey")
	}

	// -----BEGIN PRIVATE KEY-----
	_, err = x509.ParsePKCS8PrivateKey(data)
	if err == nil {
		fmt.Println("ParsePKCS8PrivateKey")
	}

	// -----BEGIN CERTIFICATE-----
	_, err = x509.ParseCertificate(data)
	if err == nil {
		fmt.Println("ParseCertificate")
	}

	// -----BEGIN CERTIFICATE REQUEST-----
	_, err = x509.ParseCertificateRequest(data)
	if err == nil {
		fmt.Println("ParseCertificateRequest")
	}

	// -----BEGIN PUBLIC KEY-----
	_, err = x509.ParsePKIXPublicKey(data)
	if err == nil {
		fmt.Println("ParsePKIXPublicKey")
	}
}

func Verfiy() {
	private, _ := rsa.GenerateKey(rand.Reader, 1024)
	public := &private.PublicKey

	data := []byte("Hello")
	hash := crypto.SHA256
	h := hash.New()
	h.Write(data)
	hashed := h.Sum(nil)

	var err error
	sign, err := rsa.SignPSS(rand.Reader, private, hash, hashed, nil)
	fmt.Println("sign", hex.EncodeToString(sign), err)

	err = rsa.VerifyPSS(public, hash, hashed, sign, nil)
	fmt.Println(err)
}

func Code() {
	private, _ := rsa.GenerateKey(rand.Reader, 1024)
	public := &private.PublicKey
}

func main() {
	data, _ := ioutil.ReadFile("http/certs/client/client.key.text")
	block, _ := pem.Decode(data)
	X509Parse(block.Bytes)

	data, _ = ioutil.ReadFile("http/certs/client/client.csr")
	block, _ = pem.Decode(data)
	X509Parse(block.Bytes)

	data, _ = ioutil.ReadFile("http/certs/client/client.crt")
	block, _ = pem.Decode(data)
	X509Parse(block.Bytes)

	data, _ = ioutil.ReadFile("http/certs/client/client.pub")
	block, _ = pem.Decode(data)
	X509Parse(block.Bytes)

	data, _ = ioutil.ReadFile("http/certs/client/client.key")
	block, _ = pem.Decode(data)
	data, _ = x509.DecryptPEMBlock(block, []byte("abc123_"))
	X509Parse(data)

	Verfiy()
}
