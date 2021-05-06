package tool

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"hash"
	"os"
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
