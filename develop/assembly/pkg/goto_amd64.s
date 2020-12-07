#include "textflag.h"

// func If(ok bool, a, b int) int
TEXT ·If(SB), NOSPLIT, $8-32
    MOVQ ok+0(FP), CX // ok
    MOVQ a+8(FP), AX  // a
    MOVQ b+16(FP), BX // b

    CMPQ CX, $0
    JZ FALSE // 等于 0 则跳转 ( CX == 0)

    MOVQ AX, 0(SP)
    CALL ·Print1(SB)

    MOVQ a+8(FP), AX // AX=a,  恢复 AX 的值. 这是一个局部变量
    MOVQ AX, ret+24(FP) // return a
    RET

FALSE:
    MOVQ BX, 0(SP)
    CALL ·Print1(SB)

    MOVQ b+16(FP), BX   // BX=b, 恢复 BX 的值, 这是一个局部变量
    MOVQ BX, ret+24(FP) // return b
    RET
