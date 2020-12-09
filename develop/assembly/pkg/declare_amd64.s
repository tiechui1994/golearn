#include "textflag.h"

// var INT int
GLOBL ·INT(SB), $8
DATA ·INT+0(SB)/8, $0x10


// var ARRAY [2]byte
GLOBL ·ARRAY(SB), $16
DATA ·ARRAY+0(SB)/1, $0x10
DATA ·ARRAY+1(SB)/1, $0x20

// var STRING string
GLOBL ·STRING(SB), NOPTR, $16
DATA  ·STRING+0(SB)/8, $private<>(SB)
DATA  ·STRING+8(SB)/8, $12

GLOBL private<>(SB), NOPTR, $16
DATA private<>+0(SB)/8,$"Hello Wo"      // ...string data...
DATA private<>+8(SB)/8,$"rld!"          // ...string data...

// var SLICE []int
GLOBL ·SLICE(SB), $24
DATA ·SLICE+0(SB)/8, $slice<>(SB)
DATA ·SLICE+8(SB)/8, $4
DATA ·SLICE+16(SB)/8, $6

GLOBL slice<>(SB), NOPTR, $16
DATA slice<>+0(SB)/8, $10
DATA slice<>+8(SB)/8, $20
DATA slice<>+16(SB)/8, $21
DATA slice<>+24(SB)/8, $21

