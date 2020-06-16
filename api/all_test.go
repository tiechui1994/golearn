package api

import (
	"testing"
)

func TestUploadFlow(t *testing.T) {
	err := UploadFlow("/home/user/Downloads/china.mp3")
	t.Log("err", err)
}
