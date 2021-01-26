package qrcode

import (
	"testing"
	"os"
)

func TestQrdata(t *testing.T) {
	list := optimal_data_chunks([]byte("https://www.baidu.com 12345 122112 1212 1221"), 4)
	for _, v := range list {
		t.Log(string(v.data), len(v.data))
	}
}

func TestIter(t *testing.T) {
	N := 9
	iter := iter(N)

	var i int
	for next(iter, &i) < N {

		if i%3 == 0 {
			skip(iter)
		}
		// 0, 2, 3, 5, 6, 8
		t.Log(i)
	}
}

func TestLoop(t *testing.T) {
	rs_blocks := func(version, correct int) [][2]int {
		ans := make([][2]int, 0)
		for i := 0; i < 3; i++ {
			ans = append(ans, [...]int{version, correct})
		}
		return ans
	}

	BIT_LIMIT_TABLE = make([][]int, 0)
	for correct := 0; correct < 2; correct++ {
		subarray := []int{0}
		for version := 1; version < 4; version++ {
			rsblocks := rs_blocks(version, correct)
			sum := 0
			for _, v := range rsblocks {
				sum += v[0]
			}
			subarray = append(subarray, 8*sum)
		}
		BIT_LIMIT_TABLE = append(BIT_LIMIT_TABLE, subarray)
	}
	t.Log(BIT_LIMIT_TABLE)
}

func TestQrcode(t *testing.T) {
	code := MakeQrcode(1, ERROR_CORRECT_H, 2, 0)
	data := "https://stackoverflow.com/questions/45086162/docker-mysql-error-1396-hy000-operation-create-user-failed-for-root"
	code.AddData([]byte(data), 20)
	code.PrintAscii(nil, true)
	t.Log(code.version, code.count, /*size=*/ 21+(code.version-1)*4+2*4 /*border*/)
}

func TestPNG(t *testing.T) {
	code := MakeQrcode(1, ERROR_CORRECT_M, 1, 0)
	data := "https://stackoverflow.com/questions/45086162/docker-mysql-error-1396-hy000-operation-create-user-failed-for-root"
	code.AddData([]byte(data), 20)
	png, _ := code.PNG(400)
	fd, _ := os.Create("./www.png")
	fd.Write(png)

	jpeg, _ := code.JPEG(400)
	fd, _ = os.Create("./www.jpeg")
	fd.Write(jpeg)
}
