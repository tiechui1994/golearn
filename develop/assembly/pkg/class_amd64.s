#include "textflag.h"

// func (i Int) Twice() int
TEXT ·Int·Twice(SB), NOSPLIT, $0
    MOVQ i+0(FP), AX
    ADDQ AX, AX
    MOVQ AX, ret+8(FP)
    RET

// func (i Int) Ptr() Int
TEXT ·Int·Ptr(SB), NOSPLIT, $0
    MOVQ i+0(FP), AX
    MOVQ AX, ret+8(FP)
    RET
