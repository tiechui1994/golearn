package pkg

//go:noescape
func GetRegister() (rsp, sp, fp uint64)
