#include "textflag.h"

// func Sum(n int) int
TEXT ·Sum(SB), $8
    MOVQ n+0(FP), AX
    MOVQ ret+8(FP), BX

    CMPQ AX, $0
    JG STEP  // AX > 0: goto STEP
    JMP RETURN // goto RETURN

STEP:
   SUBQ $1, AX  // AX-=1

   MOVQ AX, 0(SP)  // AX 入参数栈
   CALL ·Sum(SB)
   MOVQ 8(SP), BX  // BX=Sum(AX-1), 获取返回值

   MOVQ n+0(FP), AX   // AX=n
   ADDQ AX, BX        // BX+=AX
   MOVQ BX, ret+8(FP) // return BX
   RET

RETURN:
    MOVQ $0, ret+8(FP) // return 0
    RET
