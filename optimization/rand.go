package main

import (
	"math/rand"
)

const (
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

type random struct {
	buffer []byte
}

var rnd random

func (r *random) randRunes(n int) string {
	if cap(r.buffer) < n {
		r.buffer = make([]byte, n)
	}
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			r.buffer[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(r.buffer)
}

func (r *random) randString(n int) string {
	if cap(r.buffer) < n {
		r.buffer = make([]byte, n)
	}
	for i := 0; i < n; i++ {
		r.buffer[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(r.buffer)
}

func RandRunes(n int) string {
	return rnd.randRunes(n)
}

func RandString(n int) string {
	return rnd.randString(n)
}
