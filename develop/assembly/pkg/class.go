package pkg

type Int int

//go:noescape
func (i Int) Twice() int

//go:noescape
func (i Int) Ptr() Int
