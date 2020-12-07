#include "textflag.h"

// func Add(a, b int) int
TEXT ·Add(SB), $8
    MOVQ a+0(FP), AX
    MOVQ b+8(FP), BX

    MOVQ AX, 0(SP)
    MOVQ BX, 8(SP)
    CALL ·Print2(SB)
    MOVQ 16(SP), CX

    MOVQ CX, 0(SP)
    CALL ·Print1(SB)

    MOVQ a+0(FP), AX
    MOVQ b+8(FP), BX

    ADDQ AX, BX
    MOVQ BX, ret+16(FP)

    RET
