package optimization

import (
	"testing"
)

func BenchmarkRandStringRunes(b *testing.B) {
	for i:=0; i<b.N; i++ {
		RandStringRunes(1000)
	}
}

func BenchmarkRandString(b *testing.B) {
	for i:=0; i<b.N; i++ {
		RandString(1000)
	}
}