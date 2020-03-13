package algorithm

import (
	"testing"
	"time"
)

var str = "BBCABCDABABCDABCDABDE"
// BBCABCDABABCDABCDABDE
// ABCDABD
//    ABCDABD
//
//

func BenchmarkBruteForce(t *testing.B) {
	var count int64
	var j int
	for i := 0; i < t.N; i++ {
		func() {
			now := time.Now()
			defer func() {
				ns := time.Now().Sub(now).Nanoseconds()
				count += ns
			}()
			j = BruteForce(str, "ABCDABD")
		}()
	}

	t.Logf("avg: %v, %v", count/int64(t.N), j)
}

func BenchmarkKMP(t *testing.B) {
	var count int64
	var j int
	for i := 0; i < t.N; i++ {
		func() {
			now := time.Now()
			defer func() {
				ns := time.Now().Sub(now).Nanoseconds()
				count += ns
			}()
			j = KMP(str, "ABCDABD")
		}()
	}
	t.Logf("avg: %v, %v", count/int64(t.N), j)
}

func TestBM(t *testing.T) {
	t.Logf("%v", BM(str, "ABCDABD"))
}

func TestSuffix(t *testing.T) {
	p := "git@gitlab.broadlink.com.cn:cloud/lhsdk.git"
	suffix1 := suffix(p)
	suffix2 := suffix_new(p)
	for i := 0; i < len(suffix1); i++ {
		if suffix1[i] != suffix2[i] {
			t.Errorf("Failed")
		} else {
			t.Logf("index [%v] = %v", i, suffix1[i])
		}
	}
	t.Logf("Success")
}
