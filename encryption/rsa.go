package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

var (
	pub_key string
	pri_key string
)

func RSAEncrpty(msg []byte) (cipher []byte, err error) {
	if pub_key == "" {
		data, _ := ioutil.ReadFile("./rsa/rsa_public_key.pem")
		pub_key = string(data)
	}

	// resolve pem format public key
	block, _ := pem.Decode([]byte(pub_key))
	if block == nil {
		return nil, fmt.Errorf("public key error")
	}

	// resolve public key
	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub := pubkey.(*rsa.PublicKey)

	return rsa.EncryptPKCS1v15(rand.Reader, pub, msg)
}

func RSADecrpty(msg []byte) (plain []byte, err error) {
	if pri_key == "" {
		data, _ := ioutil.ReadFile("./rsa/rsa_private_key.pem")
		pri_key = string(data)
	}

	// resolve pem format private key
	block, _ := pem.Decode([]byte(pri_key))
	if block == nil {
		return nil, fmt.Errorf("private key error")
	}

	// resolve private key
	privkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privkey, msg)
}
