#include "textflag.h"

// func AsmAdd(A, B int) int
// 局部变量大小为 16
TEXT ·AsmAdd(SB), NOSPLIT, $16
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

    CALL ·zero(SB)

    RET


TEXT ·Add(SB), NOSPLIT, $24-32
    MOVQ $0, r1+16(FP)
    MOVQ $0, r2+24(FP)

    LEAQ (SP), AX
    MOVQ AX, 0(SP)
    LEAQ auto+0(SP), AX
    MOVQ AX, 8(SP)
    LEAQ auto+0(FP), AX
    MOVQ AX, 16(SP)
    CALL ·Print3(SB)

    MOVQ a+0(FP), AX
    MOVQ AX, 0(SP)
    MOVQ b+8(FP), AX
    MOVQ AX, 8(SP)
    CALL ·Print2(SB)

    MOVQ $3, tmp-8(SP) // tep = 3
    MOVQ a+0(FP), AX   // AX = a
    ADDQ tmp-8(SP), AX // AX += tmp
    MOVQ AX, tmp-8(SP) // tmp = AX

    MOVQ AX, 0(SP)
    CALL ·Print1(SB)

    MOVQ a+0(FP), AX
    MOVQ AX, 0(SP)
    MOVQ b+8(FP), AX
    MOVQ AX, 8(SP)
    CALL ·add(SB)
    MOVQ 16(SP), AX
    MOVQ AX, r1+16(FP)

    MOVQ a+0(FP), AX
    SUBQ b+8(FP), AX
    MOVQ AX, r2+24(FP)
    RET


TEXT ·add(SB), NOSPLIT, $8-24
    MOVQ $0, r1+16(FP)

    MOVQ a+0(FP), AX
    ADDQ b+8(FP), AX
    MOVQ AX, r1+16(FP)
    RET


TEXT ·zero(SB), NOSPLIT, $24
    LEAQ (SP), AX
    MOVQ AX, 0(SP)
    LEAQ auto+0(SP), AX
    MOVQ AX, 8(SP)
    LEAQ auto+0(FP), AX
    MOVQ AX, 16(SP)
    CALL ·Print3(SB)
    RET
