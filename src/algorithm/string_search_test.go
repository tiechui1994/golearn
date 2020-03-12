package algorithm

import (
	"testing"
	"time"
)

var str = "BBCABCDABABCDABCDABDE"

func BenchmarkBruteForce(t *testing.B) {
	var count int64
	for i := 0; i < t.N; i++ {
		var i int
		func() {
			now := time.Now()
			defer func() {
				ns := time.Now().Sub(now).Nanoseconds()
				count += ns
			}()
			i = BruteForce(str, "ABCDABD")
		}()
	}

	t.Logf("avg: %v", count/int64(t.N))
}

func BenchmarkKMP(t *testing.B) {
	var count int64
	for i := 0; i < t.N; i++ {
		var i int
		func() {
			now := time.Now()
			defer func() {
				ns := time.Now().Sub(now).Nanoseconds()
				count += ns
			}()
			i = KMP(str, "ABCDABD")
		}()
	}
	t.Logf("avg: %v", count/int64(t.N))
}

func TestBM(t *testing.T) {
	t.Logf("%v", BM(str, "ABCDABD"))
}