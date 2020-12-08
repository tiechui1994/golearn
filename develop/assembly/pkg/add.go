package pkg

//go:noescape
func AsmAdd(a, b int) int

//go:noescape
func Add(a, b int) (int, int)

//go:noescape
func add(a, b int) int

//go:noescape
func zero()
