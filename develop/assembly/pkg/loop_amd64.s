#include "textflag.h"

// func Sequence(count, start, step int) int
TEXT ·Sequence(SB), NOSPLIT, $0-32
    MOVQ count+0(FP), AX // count
    MOVQ start+8(FP), BX // result/start
    MOVQ step+16(FP), CX // step

LOOP_BEGIN:
    MOVQ $0, DX // i

LOOP_IF:
    CMPQ DX, AX // i, count
    JL LOOP_BODY // i < count
    JMP LOOP_END

LOOP_BODY:
    ADDQ $1, DX // i+=1
    ADDQ CX, BX // result += step
    JMP LOOP_IF

LOOP_END:
    MOVQ BX, ret+24(FP) // return result
    RET
