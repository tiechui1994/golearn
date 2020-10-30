package encryption

import (
	"encoding/base64"
	"testing"
)

func TestRSA(t *testing.T) {
	origin := []byte("Hello World")
	cipher, err := RSAEncrpty(origin)
	if err != nil {
		t.Fatal("RSAEncrpty", err)
	}

	t.Log("cipher", base64.StdEncoding.EncodeToString(cipher))

	plain, err := RSADecrpty(cipher)
	if err != nil {
		t.Fatal("RSADecrpty", err)
	}

	t.Log("plain", string(plain))
}
