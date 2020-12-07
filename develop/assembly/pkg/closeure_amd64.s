#include "textflag.h"

TEXT ·ptrToFunc(SB), NOSPLIT, $0
    MOVQ ptr+0(FP), AX // AX = ptr
    MOVQ AX, ret+8(FP) // return AX
    RET

TEXT ·asmFunTwiceClosureAddr(SB), NOSPLIT, $0
    LEAQ ·asmFunTwiceClosureBody(SB), AX // AX=·asmFunTwiceClosureAddr(SB)
    MOVQ AX, ret+0(FP)
    RET

TEXT ·asmFunTwiceClosureBody(SB), NOSPLIT|NEEDCTXT, $0
    MOVQ 8(DX), AX     // 获取 X 的值
    ADDQ AX, AX        // AX += AX
    MOVQ AX, 8(DX)     // 将 X 保存到 DX 当中
    MOVQ AX, ret+0(FP) // return
    RET
