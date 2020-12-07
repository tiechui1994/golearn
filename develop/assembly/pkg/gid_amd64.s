#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"


// func getg() int64
TEXT ·getg(SB), NOSPLIT, $8
    MOVQ (TLS), AX
    MOVQ g_goid(AX), AX
    MOVQ AX, ret+0(FP)
    RET
