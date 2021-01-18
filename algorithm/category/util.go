package category

func max(args ...int) int {
	max := args[0]
	for i := 1; i < len(args); i++ {
		if max < args[i] {
			max = args[i]
		}
	}
	return max
}
