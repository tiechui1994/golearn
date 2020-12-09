#include "textflag.h"

// var INT int
GLOBL ·INT(SB), $8
DATA ·INT+0(SB)/8,$0x3725

// var ARRAY [2]byte
GLOBL ·ARRAY(SB), $16
DATA ·ARRAY+0(SB)/1, $0x10
DATA ·ARRAY+1(SB)/1, $0x20

// var STRING string
GLOBL ·STRING(SB), NOPTR, $16
DATA  ·STRING+0(SB)/8, $text<>(SB)
DATA  ·STRING+8(SB)/8, $6

GLOBL text<>(SB),$16
DATA text<>+0(SB)/8,$"Hello Wo"      // ...string data...
DATA text<>+8(SB)/8,$"rld!"          // ...string data...
