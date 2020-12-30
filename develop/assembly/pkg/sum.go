package pkg

// sum = 1+2+..+n
//go:noescape
func Sum(n int) int

/*
func Sum(n int) int {
	if n > 0 {
		return n+Sum(n-1)
	}

	return 0
}

func Sum(n int) int {
	var AX = n
	var BX = 0

	if n > 0 {
		goto STEP
	}
	goto END

STEP:
	AX -= 1
	BX = Sum(AX)

	AX = n // 恢复
	BX += AX

	return BX

END:
	return 0
}

*/
