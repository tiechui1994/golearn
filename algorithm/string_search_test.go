package algorithm

import (
	"testing"
	"time"
)

var str = "BBCABCDABABCDABCDABDE"

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

/**

012345678901234567890
BBCABCDABABCDABCDABDE
ABCDABD
   ABCDABD         [第一步, 好后缀匹配CD, 和 坏字符串'C'是一样的]
     ABCDABD       [第二步, 好后缀匹配A, 和 坏字符串'A'是一样的 ]
         ABCDABD
		     ABCDABD



01234567890123
ABCAACBABBACAB
ABCBAB
   ABCBAB
       ABCBAB
		ABCBAB
**/
func TestBM(t *testing.T) {
	//t.Logf("%v", BM(str, "ABCDABD"))
	t.Logf("%v", BM("ABCAACBABBACAB", "ABCBAB"))
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
