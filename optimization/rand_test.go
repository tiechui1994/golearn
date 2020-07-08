package optimization

import "testing"

func BenchmarkRandStringRunes(b *testing.B) {
	RandString(1000000)
}