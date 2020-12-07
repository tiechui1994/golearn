#include "textflag.h"

// func Add(A, B int) int
// 局部变量大小为 16
TEXT ·Add(SB), NOSPLIT, $16
    MOVQ A+0(FP), AX
    MOVQ B+8(FP), BX

    MOVQ AX, 0(SP) // 局部变量
    MOVQ BX, 8(SP) // 局部变量
    CALL ·Print2(SB)
    MOVQ 16(SP), CX

    MOVQ CX, 0(SP)  // 局部变量
    CALL ·Print1(SB)

    MOVQ A+0(FP), AX
    MOVQ B+8(FP), BX

    ADDQ AX, BX
    MOVQ BX, RET+16(FP)

    RET

