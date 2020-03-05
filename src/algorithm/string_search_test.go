package algorithm

import "testing"

func TestBruteForce(t *testing.T) {
	i := BruteForce("ABCXACX", "CX")
	t.Logf("res: %v", i)
}
