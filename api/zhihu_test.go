package api

import (
	"testing"
	"fmt"
)

func TestGetImageWidthAndHeight(t *testing.T) {
	w, h := GetImageWidthAndHeight("/home/user/Downloads/ai/aa.png")
	fmt.Println("png", w, h)
	w, h = GetImageWidthAndHeight("/home/user/Downloads/ai/test.jpg")
	fmt.Println("jpeg", w, h)
}

func TestUpload(t *testing.T) {
	Upload()
}