#include "textflag.h"

// func If(ok int, a, b int) int
TEXT ·If(SB), NOSPLIT, $0-32
    MOVQ ok+0(FP), CX
    MOVQ a+8(FP), AX
    MOVQ b+16(FP), BX

    CMPQ CX, $0
    JZ FALSE
    MOVQ AX, ret+24(FP) // return a
    RET

FALSE:
    MOVQ AX, ret+24(FP)
    MOVQ AX, 0(SP)
    CALL ·Print(SB)
    RET
