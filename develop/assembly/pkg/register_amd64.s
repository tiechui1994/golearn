#include "textflag.h"

//func GetRegister() (rsp, sp, fp uint64)
TEXT ·GetRegister(SB), NOSPLIT, $24
    LEAQ (SP), AX
    LEAQ a+0(SP), BX
    LEAQ b+0(FP), CX
    MOVQ AX, rsp+0(FP)
    MOVQ BX, sp+8(FP)
    MOVQ CX, fp+24(FP)
    RET
