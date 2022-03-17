package backtrace

import (
	"testing"
)

func TestNQueue(t *testing.T) {
	for i := 3; i < 30; i++ {
		if result := NQueue(i); len(result) > 0 {
			t.Logf("%v 皇后问题:", i)
			for k := 0; k < len(result); k++ {
				str := "|"
				for _, v := range result[k] {
					val := " "
					if v == "Q" {
						val = "Q"
					}
					str += "" + val + "|"
				}

				t.Logf("%v", str)
			}
		}
	}
}
