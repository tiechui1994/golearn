package pkg

// 等差数列
//go:nosplit
func Sequence(count, start, step int) int

/*
// 等差数列
func Sequence(count, start, step int) int {
	result := start
	for i := 0; i < count; i++ {
		result += step
	}

	return result
}

// 对上述的 Sequence 进行汇编语法改造:
func Sequence(count, start, step int) int {
	var i = 0
	var result = 0

LOOP_BEGIN:
	result = start

LOOP_IF:
	if i < count {
		goto LOOP_BODY
	}
	goto LOOP_END

LOOP_BODY:
	i = i + 1
	result = result + step
	goto LOOP_IF

LOOP_END:
	return result
}

*/
