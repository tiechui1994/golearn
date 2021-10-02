"".main STEXT size=171 args=0x0 locals=0x58
	0x0000 00000 (iface_struct.go:5)	TEXT	"".main(SB), ABIInternal, $88-0
	0x0000 00000 (iface_struct.go:5)	MOVQ	(TLS), CX
	0x0009 00009 (iface_struct.go:5)	CMPQ	SP, 16(CX)
	0x000d 00013 (iface_struct.go:5)	JLS	161
	0x0013 00019 (iface_struct.go:5)	SUBQ	$88, SP
	0x0017 00023 (iface_struct.go:5)	MOVQ	BP, 80(SP)
	0x001c 00028 (iface_struct.go:5)	LEAQ	80(SP), BP
	0x0021 00033 (iface_struct.go:6)	XORPS	X0, X0
	0x0024 00036 (iface_struct.go:6)	MOVUPS	X0, ""..autotmp_1+64(SP)
	0x0029 00041 (iface_struct.go:6)	LEAQ	go.string."halfrost"(SB), AX
	0x0030 00048 (iface_struct.go:6)	MOVQ	AX, ""..autotmp_1+64(SP)
	0x0035 00053 (iface_struct.go:6)	MOVQ	$8, ""..autotmp_1+72(SP)
	0x003e 00062 (iface_struct.go:6)	MOVQ	AX, (SP)
	0x0042 00066 (iface_struct.go:6)	MOVQ	$8, 8(SP)
	0x004b 00075 (iface_struct.go:6)	CALL	runtime.convTstring(SB)
	0x0050 00080 (iface_struct.go:6)	MOVQ	16(SP), AX
	0x0055 00085 (iface_struct.go:6)	MOVQ	AX, ""..autotmp_2+40(SP)
	0x005a 00090 (iface_struct.go:6)	LEAQ	go.itab."".Student,"".Person(SB), CX
	0x0061 00097 (iface_struct.go:6)	MOVQ	CX, "".s+48(SP)
	0x0066 00102 (iface_struct.go:6)	MOVQ	AX, "".s+56(SP)
	0x006b 00107 (iface_struct.go:7)	MOVQ	"".s+48(SP), AX
	0x0070 00112 (iface_struct.go:7)	TESTB	AL, (AX)
	0x0072 00114 (iface_struct.go:7)	MOVQ	32(AX), AX
	0x0076 00118 (iface_struct.go:7)	MOVQ	"".s+56(SP), CX
	0x007b 00123 (iface_struct.go:7)	MOVQ	CX, (SP)
	0x007f 00127 (iface_struct.go:7)	LEAQ	go.string."everyone"(SB), CX
	0x0086 00134 (iface_struct.go:7)	MOVQ	CX, 8(SP)
	0x008b 00139 (iface_struct.go:7)	MOVQ	$8, 16(SP)
	0x0094 00148 (iface_struct.go:7)	CALL	AX
	0x0096 00150 (iface_struct.go:8)	MOVQ	80(SP), BP
	0x009b 00155 (iface_struct.go:8)	ADDQ	$88, SP
	0x009f 00159 (iface_struct.go:8)	NOP
	0x00a0 00160 (iface_struct.go:8)	RET
	0x00a1 00161 (iface_struct.go:8)	NOP
	0x00a1 00161 (iface_struct.go:5)	CALL	runtime.morestack_noctxt(SB)
	0x00a6 00166 (iface_struct.go:5)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 0f 86 8e  dH..%....H;a....
	0x0010 00 00 00 48 83 ec 58 48 89 6c 24 50 48 8d 6c 24  ...H..XH.l$PH.l$
	0x0020 50 0f 57 c0 0f 11 44 24 40 48 8d 05 00 00 00 00  P.W...D$@H......
	0x0030 48 89 44 24 40 48 c7 44 24 48 08 00 00 00 48 89  H.D$@H.D$H....H.
	0x0040 04 24 48 c7 44 24 08 08 00 00 00 e8 00 00 00 00  .$H.D$..........
	0x0050 48 8b 44 24 10 48 89 44 24 28 48 8d 0d 00 00 00  H.D$.H.D$(H.....
	0x0060 00 48 89 4c 24 30 48 89 44 24 38 48 8b 44 24 30  .H.L$0H.D$8H.D$0
	0x0070 84 00 48 8b 40 20 48 8b 4c 24 38 48 89 0c 24 48  ..H.@ H.L$8H..$H
	0x0080 8d 0d 00 00 00 00 48 89 4c 24 08 48 c7 44 24 10  ......H.L$.H.D$.
	0x0090 08 00 00 00 ff d0 48 8b 6c 24 50 48 83 c4 58 90  ......H.l$PH..X.
	0x00a0 c3 e8 00 00 00 00 e9 55 ff ff ff                 .......U...
	rel 5+4 t=17 TLS+0
	rel 44+4 t=16 go.string."halfrost"+0
	rel 76+4 t=8 runtime.convTstring+0
	rel 93+4 t=16 go.itab."".Student,"".Person+0
	rel 130+4 t=16 go.string."everyone"+0
	rel 148+0 t=11 +0
	rel 162+4 t=8 runtime.morestack_noctxt+0
"".Student.sayHello STEXT size=421 args=0x30 locals=0xa0
	0x0000 00000 (iface_struct.go:20)	TEXT	"".Student.sayHello(SB), ABIInternal, $160-48
	0x0000 00000 (iface_struct.go:20)	MOVQ	(TLS), CX
	0x0009 00009 (iface_struct.go:20)	LEAQ	-32(SP), AX
	0x000e 00014 (iface_struct.go:20)	CMPQ	AX, 16(CX)
	0x0012 00018 (iface_struct.go:20)	JLS	406
	0x0018 00024 (iface_struct.go:20)	SUBQ	$160, SP
	0x001f 00031 (iface_struct.go:20)	MOVQ	BP, 152(SP)
	0x0027 00039 (iface_struct.go:20)	LEAQ	152(SP), BP
	0x002f 00047 (iface_struct.go:20)	XORPS	X0, X0
	0x0032 00050 (iface_struct.go:20)	MOVUPS	X0, "".~r1+200(SP)
	0x003a 00058 (iface_struct.go:21)	XORPS	X0, X0
	0x003d 00061 (iface_struct.go:21)	MOVUPS	X0, ""..autotmp_3+120(SP)
	0x0042 00066 (iface_struct.go:21)	MOVUPS	X0, ""..autotmp_3+136(SP)
	0x004a 00074 (iface_struct.go:21)	LEAQ	""..autotmp_3+120(SP), AX
	0x004f 00079 (iface_struct.go:21)	MOVQ	AX, ""..autotmp_6+72(SP)
	0x0054 00084 (iface_struct.go:21)	MOVQ	"".s+168(SP), AX
	0x005c 00092 (iface_struct.go:21)	MOVQ	"".s+176(SP), CX
	0x0064 00100 (iface_struct.go:21)	MOVQ	AX, (SP)
	0x0068 00104 (iface_struct.go:21)	MOVQ	CX, 8(SP)
	0x006d 00109 (iface_struct.go:21)	CALL	runtime.convTstring(SB)
	0x0072 00114 (iface_struct.go:21)	MOVQ	16(SP), AX
	0x0077 00119 (iface_struct.go:21)	MOVQ	AX, ""..autotmp_7+64(SP)
	0x007c 00124 (iface_struct.go:21)	MOVQ	""..autotmp_6+72(SP), CX
	0x0081 00129 (iface_struct.go:21)	TESTB	AL, (CX)
	0x0083 00131 (iface_struct.go:21)	LEAQ	type.string(SB), DX
	0x008a 00138 (iface_struct.go:21)	MOVQ	DX, (CX)
	0x008d 00141 (iface_struct.go:21)	LEAQ	8(CX), DI
	0x0091 00145 (iface_struct.go:21)	CMPL	runtime.writeBarrier(SB), $0
	0x0098 00152 (iface_struct.go:21)	JEQ	159
	0x009a 00154 (iface_struct.go:21)	JMP	396
	0x009f 00159 (iface_struct.go:21)	MOVQ	AX, 8(CX)
	0x00a3 00163 (iface_struct.go:21)	JMP	165
	0x00a5 00165 (iface_struct.go:21)	MOVQ	"".name+184(SP), AX
	0x00ad 00173 (iface_struct.go:21)	MOVQ	"".name+192(SP), CX
	0x00b5 00181 (iface_struct.go:21)	MOVQ	AX, (SP)
	0x00b9 00185 (iface_struct.go:21)	MOVQ	CX, 8(SP)
	0x00be 00190 (iface_struct.go:21)	NOP
	0x00c0 00192 (iface_struct.go:21)	CALL	runtime.convTstring(SB)
	0x00c5 00197 (iface_struct.go:21)	MOVQ	16(SP), AX
	0x00ca 00202 (iface_struct.go:21)	MOVQ	AX, ""..autotmp_8+56(SP)
	0x00cf 00207 (iface_struct.go:21)	MOVQ	""..autotmp_6+72(SP), CX
	0x00d4 00212 (iface_struct.go:21)	TESTB	AL, (CX)
	0x00d6 00214 (iface_struct.go:21)	LEAQ	type.string(SB), DX
	0x00dd 00221 (iface_struct.go:21)	MOVQ	DX, 16(CX)
	0x00e1 00225 (iface_struct.go:21)	LEAQ	24(CX), DI
	0x00e5 00229 (iface_struct.go:21)	CMPL	runtime.writeBarrier(SB), $0
	0x00ec 00236 (iface_struct.go:21)	JEQ	243
	0x00ee 00238 (iface_struct.go:21)	JMP	386
	0x00f3 00243 (iface_struct.go:21)	MOVQ	AX, 24(CX)
	0x00f7 00247 (iface_struct.go:21)	JMP	249
	0x00f9 00249 (iface_struct.go:21)	MOVQ	""..autotmp_6+72(SP), AX
	0x00fe 00254 (iface_struct.go:21)	TESTB	AL, (AX)
	0x0100 00256 (iface_struct.go:21)	JMP	258
	0x0102 00258 (iface_struct.go:21)	MOVQ	AX, ""..autotmp_5+96(SP)
	0x0107 00263 (iface_struct.go:21)	MOVQ	$2, ""..autotmp_5+104(SP)
	0x0110 00272 (iface_struct.go:21)	MOVQ	$2, ""..autotmp_5+112(SP)
	0x0119 00281 (iface_struct.go:21)	LEAQ	go.string."%v: Hello %v, nice to meet you.\n"(SB), AX
	0x0120 00288 (iface_struct.go:21)	MOVQ	AX, (SP)
	0x0124 00292 (iface_struct.go:21)	MOVQ	$32, 8(SP)
	0x012d 00301 (iface_struct.go:21)	MOVQ	""..autotmp_5+96(SP), AX
	0x0132 00306 (iface_struct.go:21)	MOVQ	AX, 16(SP)
	0x0137 00311 (iface_struct.go:21)	MOVQ	$2, 24(SP)
	0x0140 00320 (iface_struct.go:21)	MOVQ	$2, 32(SP)
	0x0149 00329 (iface_struct.go:21)	CALL	fmt.Sprintf(SB)
	0x014e 00334 (iface_struct.go:21)	MOVQ	40(SP), AX
	0x0153 00339 (iface_struct.go:21)	MOVQ	48(SP), CX
	0x0158 00344 (iface_struct.go:21)	MOVQ	AX, ""..autotmp_4+80(SP)
	0x015d 00349 (iface_struct.go:21)	MOVQ	CX, ""..autotmp_4+88(SP)
	0x0162 00354 (iface_struct.go:21)	MOVQ	AX, "".~r1+200(SP)
	0x016a 00362 (iface_struct.go:21)	MOVQ	CX, "".~r1+208(SP)
	0x0172 00370 (iface_struct.go:21)	MOVQ	152(SP), BP
	0x017a 00378 (iface_struct.go:21)	ADDQ	$160, SP
	0x0181 00385 (iface_struct.go:21)	RET
	0x0182 00386 (iface_struct.go:21)	CALL	runtime.gcWriteBarrier(SB)
	0x0187 00391 (iface_struct.go:21)	JMP	249
	0x018c 00396 (iface_struct.go:21)	CALL	runtime.gcWriteBarrier(SB)
	0x0191 00401 (iface_struct.go:21)	JMP	165
	0x0196 00406 (iface_struct.go:21)	NOP
	0x0196 00406 (iface_struct.go:20)	CALL	runtime.morestack_noctxt(SB)
	0x019b 00411 (iface_struct.go:20)	NOP
	0x01a0 00416 (iface_struct.go:20)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 8d 44 24 e0 48 3b  dH..%....H.D$.H;
	0x0010 41 10 0f 86 7e 01 00 00 48 81 ec a0 00 00 00 48  A...~...H......H
	0x0020 89 ac 24 98 00 00 00 48 8d ac 24 98 00 00 00 0f  ..$....H..$.....
	0x0030 57 c0 0f 11 84 24 c8 00 00 00 0f 57 c0 0f 11 44  W....$.....W...D
	0x0040 24 78 0f 11 84 24 88 00 00 00 48 8d 44 24 78 48  $x...$....H.D$xH
	0x0050 89 44 24 48 48 8b 84 24 a8 00 00 00 48 8b 8c 24  .D$HH..$....H..$
	0x0060 b0 00 00 00 48 89 04 24 48 89 4c 24 08 e8 00 00  ....H..$H.L$....
	0x0070 00 00 48 8b 44 24 10 48 89 44 24 40 48 8b 4c 24  ..H.D$.H.D$@H.L$
	0x0080 48 84 01 48 8d 15 00 00 00 00 48 89 11 48 8d 79  H..H......H..H.y
	0x0090 08 83 3d 00 00 00 00 00 74 05 e9 ed 00 00 00 48  ..=.....t......H
	0x00a0 89 41 08 eb 00 48 8b 84 24 b8 00 00 00 48 8b 8c  .A...H..$....H..
	0x00b0 24 c0 00 00 00 48 89 04 24 48 89 4c 24 08 66 90  $....H..$H.L$.f.
	0x00c0 e8 00 00 00 00 48 8b 44 24 10 48 89 44 24 38 48  .....H.D$.H.D$8H
	0x00d0 8b 4c 24 48 84 01 48 8d 15 00 00 00 00 48 89 51  .L$H..H......H.Q
	0x00e0 10 48 8d 79 18 83 3d 00 00 00 00 00 74 05 e9 8f  .H.y..=.....t...
	0x00f0 00 00 00 48 89 41 18 eb 00 48 8b 44 24 48 84 00  ...H.A...H.D$H..
	0x0100 eb 00 48 89 44 24 60 48 c7 44 24 68 02 00 00 00  ..H.D$`H.D$h....
	0x0110 48 c7 44 24 70 02 00 00 00 48 8d 05 00 00 00 00  H.D$p....H......
	0x0120 48 89 04 24 48 c7 44 24 08 20 00 00 00 48 8b 44  H..$H.D$. ...H.D
	0x0130 24 60 48 89 44 24 10 48 c7 44 24 18 02 00 00 00  $`H.D$.H.D$.....
	0x0140 48 c7 44 24 20 02 00 00 00 e8 00 00 00 00 48 8b  H.D$ .........H.
	0x0150 44 24 28 48 8b 4c 24 30 48 89 44 24 50 48 89 4c  D$(H.L$0H.D$PH.L
	0x0160 24 58 48 89 84 24 c8 00 00 00 48 89 8c 24 d0 00  $XH..$....H..$..
	0x0170 00 00 48 8b ac 24 98 00 00 00 48 81 c4 a0 00 00  ..H..$....H.....
	0x0180 00 c3 e8 00 00 00 00 e9 6d ff ff ff e8 00 00 00  ........m.......
	0x0190 00 e9 0f ff ff ff e8 00 00 00 00 0f 1f 44 00 00  .............D..
	0x01a0 e9 5b fe ff ff                                   .[...
	rel 5+4 t=17 TLS+0
	rel 110+4 t=8 runtime.convTstring+0
	rel 134+4 t=16 type.string+0
	rel 147+4 t=16 runtime.writeBarrier+-1
	rel 193+4 t=8 runtime.convTstring+0
	rel 217+4 t=16 type.string+0
	rel 231+4 t=16 runtime.writeBarrier+-1
	rel 284+4 t=16 go.string."%v: Hello %v, nice to meet you.\n"+0
	rel 330+4 t=8 fmt.Sprintf+0
	rel 387+4 t=8 runtime.gcWriteBarrier+0
	rel 397+4 t=8 runtime.gcWriteBarrier+0
	rel 407+4 t=8 runtime.morestack_noctxt+0
type..eq.[2]interface {} STEXT dupok size=236 args=0x18 locals=0x50
	0x0000 00000 (<autogenerated>:1)	TEXT	type..eq.[2]interface {}(SB), DUPOK|ABIInternal, $80-24
	0x0000 00000 (<autogenerated>:1)	MOVQ	(TLS), CX
	0x0009 00009 (<autogenerated>:1)	CMPQ	SP, 16(CX)
	0x000d 00013 (<autogenerated>:1)	JLS	226
	0x0013 00019 (<autogenerated>:1)	SUBQ	$80, SP
	0x0017 00023 (<autogenerated>:1)	MOVQ	BP, 72(SP)
	0x001c 00028 (<autogenerated>:1)	LEAQ	72(SP), BP
	0x0021 00033 (<autogenerated>:1)	MOVB	$0, "".r+104(SP)
	0x0026 00038 (<autogenerated>:1)	MOVQ	$0, ""..autotmp_3+32(SP)
	0x002f 00047 (<autogenerated>:1)	JMP	49
	0x0031 00049 (<autogenerated>:1)	CMPQ	""..autotmp_3+32(SP), $2
	0x0037 00055 (<autogenerated>:1)	JLT	62
	0x0039 00057 (<autogenerated>:1)	JMP	211
	0x003e 00062 (<autogenerated>:1)	MOVQ	""..autotmp_3+32(SP), AX
	0x0043 00067 (<autogenerated>:1)	SHLQ	$4, AX
	0x0047 00071 (<autogenerated>:1)	ADDQ	"".q+96(SP), AX
	0x004c 00076 (<autogenerated>:1)	MOVQ	(AX), CX
	0x004f 00079 (<autogenerated>:1)	MOVQ	8(AX), AX
	0x0053 00083 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_4+56(SP)
	0x0058 00088 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_4+64(SP)
	0x005d 00093 (<autogenerated>:1)	MOVQ	""..autotmp_3+32(SP), AX
	0x0062 00098 (<autogenerated>:1)	SHLQ	$4, AX
	0x0066 00102 (<autogenerated>:1)	ADDQ	"".p+88(SP), AX
	0x006b 00107 (<autogenerated>:1)	MOVQ	(AX), CX
	0x006e 00110 (<autogenerated>:1)	MOVQ	8(AX), AX
	0x0072 00114 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_5+40(SP)
	0x0077 00119 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_5+48(SP)
	0x007c 00124 (<autogenerated>:1)	NOP
	0x0080 00128 (<autogenerated>:1)	CMPQ	""..autotmp_4+56(SP), CX
	0x0085 00133 (<autogenerated>:1)	JEQ	137
	0x0087 00135 (<autogenerated>:1)	JMP	209
	0x0089 00137 (<autogenerated>:1)	MOVQ	CX, (SP)
	0x008d 00141 (<autogenerated>:1)	MOVQ	AX, 8(SP)
	0x0092 00146 (<autogenerated>:1)	MOVQ	""..autotmp_4+64(SP), AX
	0x0097 00151 (<autogenerated>:1)	MOVQ	AX, 16(SP)
	0x009c 00156 (<autogenerated>:1)	NOP
	0x00a0 00160 (<autogenerated>:1)	CALL	runtime.efaceeq(SB)
	0x00a5 00165 (<autogenerated>:1)	CMPB	24(SP), $0
	0x00aa 00170 (<autogenerated>:1)	JNE	174
	0x00ac 00172 (<autogenerated>:1)	JMP	197
	0x00ae 00174 (<autogenerated>:1)	JMP	176
	0x00b0 00176 (<autogenerated>:1)	MOVQ	""..autotmp_3+32(SP), AX
	0x00b5 00181 (<autogenerated>:1)	INCQ	AX
	0x00b8 00184 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_3+32(SP)
	0x00bd 00189 (<autogenerated>:1)	NOP
	0x00c0 00192 (<autogenerated>:1)	JMP	49
	0x00c5 00197 (<autogenerated>:1)	JMP	199
	0x00c7 00199 (<autogenerated>:1)	MOVQ	72(SP), BP
	0x00cc 00204 (<autogenerated>:1)	ADDQ	$80, SP
	0x00d0 00208 (<autogenerated>:1)	RET
	0x00d1 00209 (<autogenerated>:1)	JMP	199
	0x00d3 00211 (<autogenerated>:1)	MOVB	$1, "".r+104(SP)
	0x00d8 00216 (<autogenerated>:1)	MOVQ	72(SP), BP
	0x00dd 00221 (<autogenerated>:1)	ADDQ	$80, SP
	0x00e1 00225 (<autogenerated>:1)	RET
	0x00e2 00226 (<autogenerated>:1)	NOP
	0x00e2 00226 (<autogenerated>:1)	CALL	runtime.morestack_noctxt(SB)
	0x00e7 00231 (<autogenerated>:1)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 0f 86 cf  dH..%....H;a....
	0x0010 00 00 00 48 83 ec 50 48 89 6c 24 48 48 8d 6c 24  ...H..PH.l$HH.l$
	0x0020 48 c6 44 24 68 00 48 c7 44 24 20 00 00 00 00 eb  H.D$h.H.D$ .....
	0x0030 00 48 83 7c 24 20 02 7c 05 e9 95 00 00 00 48 8b  .H.|$ .|......H.
	0x0040 44 24 20 48 c1 e0 04 48 03 44 24 60 48 8b 08 48  D$ H...H.D$`H..H
	0x0050 8b 40 08 48 89 4c 24 38 48 89 44 24 40 48 8b 44  .@.H.L$8H.D$@H.D
	0x0060 24 20 48 c1 e0 04 48 03 44 24 58 48 8b 08 48 8b  $ H...H.D$XH..H.
	0x0070 40 08 48 89 4c 24 28 48 89 44 24 30 0f 1f 40 00  @.H.L$(H.D$0..@.
	0x0080 48 39 4c 24 38 74 02 eb 48 48 89 0c 24 48 89 44  H9L$8t..HH..$H.D
	0x0090 24 08 48 8b 44 24 40 48 89 44 24 10 0f 1f 40 00  $.H.D$@H.D$...@.
	0x00a0 e8 00 00 00 00 80 7c 24 18 00 75 02 eb 17 eb 00  ......|$..u.....
	0x00b0 48 8b 44 24 20 48 ff c0 48 89 44 24 20 0f 1f 00  H.D$ H..H.D$ ...
	0x00c0 e9 6c ff ff ff eb 00 48 8b 6c 24 48 48 83 c4 50  .l.....H.l$HH..P
	0x00d0 c3 eb f4 c6 44 24 68 01 48 8b 6c 24 48 48 83 c4  ....D$h.H.l$HH..
	0x00e0 50 c3 e8 00 00 00 00 e9 14 ff ff ff              P...........
	rel 5+4 t=17 TLS+0
	rel 161+4 t=8 runtime.efaceeq+0
	rel 227+4 t=8 runtime.morestack_noctxt+0
"".Student.sayGoodbye STEXT size=421 args=0x30 locals=0xa0
	0x0000 00000 (iface_struct.go:25)	TEXT	"".Student.sayGoodbye(SB), ABIInternal, $160-48
	0x0000 00000 (iface_struct.go:25)	MOVQ	(TLS), CX
	0x0009 00009 (iface_struct.go:25)	LEAQ	-32(SP), AX
	0x000e 00014 (iface_struct.go:25)	CMPQ	AX, 16(CX)
	0x0012 00018 (iface_struct.go:25)	JLS	406
	0x0018 00024 (iface_struct.go:25)	SUBQ	$160, SP
	0x001f 00031 (iface_struct.go:25)	MOVQ	BP, 152(SP)
	0x0027 00039 (iface_struct.go:25)	LEAQ	152(SP), BP
	0x002f 00047 (iface_struct.go:25)	XORPS	X0, X0
	0x0032 00050 (iface_struct.go:25)	MOVUPS	X0, "".~r1+200(SP)
	0x003a 00058 (iface_struct.go:26)	XORPS	X0, X0
	0x003d 00061 (iface_struct.go:26)	MOVUPS	X0, ""..autotmp_3+120(SP)
	0x0042 00066 (iface_struct.go:26)	MOVUPS	X0, ""..autotmp_3+136(SP)
	0x004a 00074 (iface_struct.go:26)	LEAQ	""..autotmp_3+120(SP), AX
	0x004f 00079 (iface_struct.go:26)	MOVQ	AX, ""..autotmp_6+72(SP)
	0x0054 00084 (iface_struct.go:26)	MOVQ	"".s+168(SP), AX
	0x005c 00092 (iface_struct.go:26)	MOVQ	"".s+176(SP), CX
	0x0064 00100 (iface_struct.go:26)	MOVQ	AX, (SP)
	0x0068 00104 (iface_struct.go:26)	MOVQ	CX, 8(SP)
	0x006d 00109 (iface_struct.go:26)	CALL	runtime.convTstring(SB)
	0x0072 00114 (iface_struct.go:26)	MOVQ	16(SP), AX
	0x0077 00119 (iface_struct.go:26)	MOVQ	AX, ""..autotmp_7+64(SP)
	0x007c 00124 (iface_struct.go:26)	MOVQ	""..autotmp_6+72(SP), CX
	0x0081 00129 (iface_struct.go:26)	TESTB	AL, (CX)
	0x0083 00131 (iface_struct.go:26)	LEAQ	type.string(SB), DX
	0x008a 00138 (iface_struct.go:26)	MOVQ	DX, (CX)
	0x008d 00141 (iface_struct.go:26)	LEAQ	8(CX), DI
	0x0091 00145 (iface_struct.go:26)	CMPL	runtime.writeBarrier(SB), $0
	0x0098 00152 (iface_struct.go:26)	JEQ	159
	0x009a 00154 (iface_struct.go:26)	JMP	396
	0x009f 00159 (iface_struct.go:26)	MOVQ	AX, 8(CX)
	0x00a3 00163 (iface_struct.go:26)	JMP	165
	0x00a5 00165 (iface_struct.go:26)	MOVQ	"".name+184(SP), AX
	0x00ad 00173 (iface_struct.go:26)	MOVQ	"".name+192(SP), CX
	0x00b5 00181 (iface_struct.go:26)	MOVQ	AX, (SP)
	0x00b9 00185 (iface_struct.go:26)	MOVQ	CX, 8(SP)
	0x00be 00190 (iface_struct.go:26)	NOP
	0x00c0 00192 (iface_struct.go:26)	CALL	runtime.convTstring(SB)
	0x00c5 00197 (iface_struct.go:26)	MOVQ	16(SP), AX
	0x00ca 00202 (iface_struct.go:26)	MOVQ	AX, ""..autotmp_8+56(SP)
	0x00cf 00207 (iface_struct.go:26)	MOVQ	""..autotmp_6+72(SP), CX
	0x00d4 00212 (iface_struct.go:26)	TESTB	AL, (CX)
	0x00d6 00214 (iface_struct.go:26)	LEAQ	type.string(SB), DX
	0x00dd 00221 (iface_struct.go:26)	MOVQ	DX, 16(CX)
	0x00e1 00225 (iface_struct.go:26)	LEAQ	24(CX), DI
	0x00e5 00229 (iface_struct.go:26)	CMPL	runtime.writeBarrier(SB), $0
	0x00ec 00236 (iface_struct.go:26)	JEQ	243
	0x00ee 00238 (iface_struct.go:26)	JMP	386
	0x00f3 00243 (iface_struct.go:26)	MOVQ	AX, 24(CX)
	0x00f7 00247 (iface_struct.go:26)	JMP	249
	0x00f9 00249 (iface_struct.go:26)	MOVQ	""..autotmp_6+72(SP), AX
	0x00fe 00254 (iface_struct.go:26)	TESTB	AL, (AX)
	0x0100 00256 (iface_struct.go:26)	JMP	258
	0x0102 00258 (iface_struct.go:26)	MOVQ	AX, ""..autotmp_5+96(SP)
	0x0107 00263 (iface_struct.go:26)	MOVQ	$2, ""..autotmp_5+104(SP)
	0x0110 00272 (iface_struct.go:26)	MOVQ	$2, ""..autotmp_5+112(SP)
	0x0119 00281 (iface_struct.go:26)	LEAQ	go.string."%v: Hi %v, see you next time.\n"(SB), AX
	0x0120 00288 (iface_struct.go:26)	MOVQ	AX, (SP)
	0x0124 00292 (iface_struct.go:26)	MOVQ	$30, 8(SP)
	0x012d 00301 (iface_struct.go:26)	MOVQ	""..autotmp_5+96(SP), AX
	0x0132 00306 (iface_struct.go:26)	MOVQ	AX, 16(SP)
	0x0137 00311 (iface_struct.go:26)	MOVQ	$2, 24(SP)
	0x0140 00320 (iface_struct.go:26)	MOVQ	$2, 32(SP)
	0x0149 00329 (iface_struct.go:26)	CALL	fmt.Sprintf(SB)
	0x014e 00334 (iface_struct.go:26)	MOVQ	40(SP), AX
	0x0153 00339 (iface_struct.go:26)	MOVQ	48(SP), CX
	0x0158 00344 (iface_struct.go:26)	MOVQ	AX, ""..autotmp_4+80(SP)
	0x015d 00349 (iface_struct.go:26)	MOVQ	CX, ""..autotmp_4+88(SP)
	0x0162 00354 (iface_struct.go:26)	MOVQ	AX, "".~r1+200(SP)
	0x016a 00362 (iface_struct.go:26)	MOVQ	CX, "".~r1+208(SP)
	0x0172 00370 (iface_struct.go:26)	MOVQ	152(SP), BP
	0x017a 00378 (iface_struct.go:26)	ADDQ	$160, SP
	0x0181 00385 (iface_struct.go:26)	RET
	0x0182 00386 (iface_struct.go:26)	CALL	runtime.gcWriteBarrier(SB)
	0x0187 00391 (iface_struct.go:26)	JMP	249
	0x018c 00396 (iface_struct.go:26)	CALL	runtime.gcWriteBarrier(SB)
	0x0191 00401 (iface_struct.go:26)	JMP	165
	0x0196 00406 (iface_struct.go:26)	NOP
	0x0196 00406 (iface_struct.go:25)	CALL	runtime.morestack_noctxt(SB)
	0x019b 00411 (iface_struct.go:25)	NOP
	0x01a0 00416 (iface_struct.go:25)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 8d 44 24 e0 48 3b  dH..%....H.D$.H;
	0x0010 41 10 0f 86 7e 01 00 00 48 81 ec a0 00 00 00 48  A...~...H......H
	0x0020 89 ac 24 98 00 00 00 48 8d ac 24 98 00 00 00 0f  ..$....H..$.....
	0x0030 57 c0 0f 11 84 24 c8 00 00 00 0f 57 c0 0f 11 44  W....$.....W...D
	0x0040 24 78 0f 11 84 24 88 00 00 00 48 8d 44 24 78 48  $x...$....H.D$xH
	0x0050 89 44 24 48 48 8b 84 24 a8 00 00 00 48 8b 8c 24  .D$HH..$....H..$
	0x0060 b0 00 00 00 48 89 04 24 48 89 4c 24 08 e8 00 00  ....H..$H.L$....
	0x0070 00 00 48 8b 44 24 10 48 89 44 24 40 48 8b 4c 24  ..H.D$.H.D$@H.L$
	0x0080 48 84 01 48 8d 15 00 00 00 00 48 89 11 48 8d 79  H..H......H..H.y
	0x0090 08 83 3d 00 00 00 00 00 74 05 e9 ed 00 00 00 48  ..=.....t......H
	0x00a0 89 41 08 eb 00 48 8b 84 24 b8 00 00 00 48 8b 8c  .A...H..$....H..
	0x00b0 24 c0 00 00 00 48 89 04 24 48 89 4c 24 08 66 90  $....H..$H.L$.f.
	0x00c0 e8 00 00 00 00 48 8b 44 24 10 48 89 44 24 38 48  .....H.D$.H.D$8H
	0x00d0 8b 4c 24 48 84 01 48 8d 15 00 00 00 00 48 89 51  .L$H..H......H.Q
	0x00e0 10 48 8d 79 18 83 3d 00 00 00 00 00 74 05 e9 8f  .H.y..=.....t...
	0x00f0 00 00 00 48 89 41 18 eb 00 48 8b 44 24 48 84 00  ...H.A...H.D$H..
	0x0100 eb 00 48 89 44 24 60 48 c7 44 24 68 02 00 00 00  ..H.D$`H.D$h....
	0x0110 48 c7 44 24 70 02 00 00 00 48 8d 05 00 00 00 00  H.D$p....H......
	0x0120 48 89 04 24 48 c7 44 24 08 1e 00 00 00 48 8b 44  H..$H.D$.....H.D
	0x0130 24 60 48 89 44 24 10 48 c7 44 24 18 02 00 00 00  $`H.D$.H.D$.....
	0x0140 48 c7 44 24 20 02 00 00 00 e8 00 00 00 00 48 8b  H.D$ .........H.
	0x0150 44 24 28 48 8b 4c 24 30 48 89 44 24 50 48 89 4c  D$(H.L$0H.D$PH.L
	0x0160 24 58 48 89 84 24 c8 00 00 00 48 89 8c 24 d0 00  $XH..$....H..$..
	0x0170 00 00 48 8b ac 24 98 00 00 00 48 81 c4 a0 00 00  ..H..$....H.....
	0x0180 00 c3 e8 00 00 00 00 e9 6d ff ff ff e8 00 00 00  ........m.......
	0x0190 00 e9 0f ff ff ff e8 00 00 00 00 0f 1f 44 00 00  .............D..
	0x01a0 e9 5b fe ff ff                                   .[...
	rel 5+4 t=17 TLS+0
	rel 110+4 t=8 runtime.convTstring+0
	rel 134+4 t=16 type.string+0
	rel 147+4 t=16 runtime.writeBarrier+-1
	rel 193+4 t=8 runtime.convTstring+0
	rel 217+4 t=16 type.string+0
	rel 231+4 t=16 runtime.writeBarrier+-1
	rel 284+4 t=16 go.string."%v: Hi %v, see you next time.\n"+0
	rel 330+4 t=8 fmt.Sprintf+0
	rel 387+4 t=8 runtime.gcWriteBarrier+0
	rel 397+4 t=8 runtime.gcWriteBarrier+0
	rel 407+4 t=8 runtime.morestack_noctxt+0
"".Person.sayGoodbye STEXT dupok size=154 args=0x30 locals=0x40
	0x0000 00000 (<autogenerated>:1)	TEXT	"".Person.sayGoodbye(SB), DUPOK|WRAPPER|ABIInternal, $64-48
	0x0000 00000 (<autogenerated>:1)	MOVQ	(TLS), CX
	0x0009 00009 (<autogenerated>:1)	CMPQ	SP, 16(CX)
	0x000d 00013 (<autogenerated>:1)	JLS	129
	0x000f 00015 (<autogenerated>:1)	SUBQ	$64, SP
	0x0013 00019 (<autogenerated>:1)	MOVQ	BP, 56(SP)
	0x0018 00024 (<autogenerated>:1)	LEAQ	56(SP), BP
	0x001d 00029 (<autogenerated>:1)	MOVQ	32(CX), BX
	0x0021 00033 (<autogenerated>:1)	TESTQ	BX, BX
	0x0024 00036 (<autogenerated>:1)	JNE	139
	0x0026 00038 (<autogenerated>:1)	NOP
	0x0026 00038 (<autogenerated>:1)	XORPS	X0, X0
	0x0029 00041 (<autogenerated>:1)	MOVUPS	X0, "".~r1+104(SP)
	0x002e 00046 (<autogenerated>:1)	MOVQ	""..this+72(SP), AX
	0x0033 00051 (<autogenerated>:1)	TESTB	AL, (AX)
	0x0035 00053 (<autogenerated>:1)	MOVQ	""..this+80(SP), CX
	0x003a 00058 (<autogenerated>:1)	MOVQ	24(AX), AX
	0x003e 00062 (<autogenerated>:1)	MOVQ	CX, (SP)
	0x0042 00066 (<autogenerated>:1)	MOVQ	"".name+88(SP), CX
	0x0047 00071 (<autogenerated>:1)	MOVQ	"".name+96(SP), DX
	0x004c 00076 (<autogenerated>:1)	MOVQ	CX, 8(SP)
	0x0051 00081 (<autogenerated>:1)	MOVQ	DX, 16(SP)
	0x0056 00086 (<autogenerated>:1)	CALL	AX
	0x0058 00088 (<autogenerated>:1)	MOVQ	32(SP), AX
	0x005d 00093 (<autogenerated>:1)	MOVQ	24(SP), CX
	0x0062 00098 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_3+40(SP)
	0x0067 00103 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_3+48(SP)
	0x006c 00108 (<autogenerated>:1)	MOVQ	CX, "".~r1+104(SP)
	0x0071 00113 (<autogenerated>:1)	MOVQ	AX, "".~r1+112(SP)
	0x0076 00118 (<autogenerated>:1)	MOVQ	56(SP), BP
	0x007b 00123 (<autogenerated>:1)	ADDQ	$64, SP
	0x007f 00127 (<autogenerated>:1)	NOP
	0x0080 00128 (<autogenerated>:1)	RET
	0x0081 00129 (<autogenerated>:1)	NOP
	0x0081 00129 (<autogenerated>:1)	CALL	runtime.morestack_noctxt(SB)
	0x0086 00134 (<autogenerated>:1)	JMP	0
	0x008b 00139 (<autogenerated>:1)	LEAQ	72(SP), DI
	0x0090 00144 (<autogenerated>:1)	CMPQ	(BX), DI
	0x0093 00147 (<autogenerated>:1)	JNE	38
	0x0095 00149 (<autogenerated>:1)	MOVQ	SP, (BX)
	0x0098 00152 (<autogenerated>:1)	JMP	38
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 72 48  dH..%....H;a.vrH
	0x0010 83 ec 40 48 89 6c 24 38 48 8d 6c 24 38 48 8b 59  ..@H.l$8H.l$8H.Y
	0x0020 20 48 85 db 75 65 0f 57 c0 0f 11 44 24 68 48 8b   H..ue.W...D$hH.
	0x0030 44 24 48 84 00 48 8b 4c 24 50 48 8b 40 18 48 89  D$H..H.L$PH.@.H.
	0x0040 0c 24 48 8b 4c 24 58 48 8b 54 24 60 48 89 4c 24  .$H.L$XH.T$`H.L$
	0x0050 08 48 89 54 24 10 ff d0 48 8b 44 24 20 48 8b 4c  .H.T$...H.D$ H.L
	0x0060 24 18 48 89 4c 24 28 48 89 44 24 30 48 89 4c 24  $.H.L$(H.D$0H.L$
	0x0070 68 48 89 44 24 70 48 8b 6c 24 38 48 83 c4 40 90  hH.D$pH.l$8H..@.
	0x0080 c3 e8 00 00 00 00 e9 75 ff ff ff 48 8d 7c 24 48  .......u...H.|$H
	0x0090 48 39 3b 75 91 48 89 23 eb 8c                    H9;u.H.#..
	rel 5+4 t=17 TLS+0
	rel 86+0 t=11 +0
	rel 130+4 t=8 runtime.morestack_noctxt+0
"".Person.sayHello STEXT dupok size=154 args=0x30 locals=0x40
	0x0000 00000 (<autogenerated>:1)	TEXT	"".Person.sayHello(SB), DUPOK|WRAPPER|ABIInternal, $64-48
	0x0000 00000 (<autogenerated>:1)	MOVQ	(TLS), CX
	0x0009 00009 (<autogenerated>:1)	CMPQ	SP, 16(CX)
	0x000d 00013 (<autogenerated>:1)	JLS	129
	0x000f 00015 (<autogenerated>:1)	SUBQ	$64, SP
	0x0013 00019 (<autogenerated>:1)	MOVQ	BP, 56(SP)
	0x0018 00024 (<autogenerated>:1)	LEAQ	56(SP), BP
	0x001d 00029 (<autogenerated>:1)	MOVQ	32(CX), BX
	0x0021 00033 (<autogenerated>:1)	TESTQ	BX, BX
	0x0024 00036 (<autogenerated>:1)	JNE	139
	0x0026 00038 (<autogenerated>:1)	NOP
	0x0026 00038 (<autogenerated>:1)	XORPS	X0, X0
	0x0029 00041 (<autogenerated>:1)	MOVUPS	X0, "".~r1+104(SP)
	0x002e 00046 (<autogenerated>:1)	MOVQ	""..this+72(SP), AX
	0x0033 00051 (<autogenerated>:1)	TESTB	AL, (AX)
	0x0035 00053 (<autogenerated>:1)	MOVQ	""..this+80(SP), CX
	0x003a 00058 (<autogenerated>:1)	MOVQ	32(AX), AX
	0x003e 00062 (<autogenerated>:1)	MOVQ	CX, (SP)
	0x0042 00066 (<autogenerated>:1)	MOVQ	"".name+88(SP), CX
	0x0047 00071 (<autogenerated>:1)	MOVQ	"".name+96(SP), DX
	0x004c 00076 (<autogenerated>:1)	MOVQ	CX, 8(SP)
	0x0051 00081 (<autogenerated>:1)	MOVQ	DX, 16(SP)
	0x0056 00086 (<autogenerated>:1)	CALL	AX
	0x0058 00088 (<autogenerated>:1)	MOVQ	32(SP), AX
	0x005d 00093 (<autogenerated>:1)	MOVQ	24(SP), CX
	0x0062 00098 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_3+40(SP)
	0x0067 00103 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_3+48(SP)
	0x006c 00108 (<autogenerated>:1)	MOVQ	CX, "".~r1+104(SP)
	0x0071 00113 (<autogenerated>:1)	MOVQ	AX, "".~r1+112(SP)
	0x0076 00118 (<autogenerated>:1)	MOVQ	56(SP), BP
	0x007b 00123 (<autogenerated>:1)	ADDQ	$64, SP
	0x007f 00127 (<autogenerated>:1)	NOP
	0x0080 00128 (<autogenerated>:1)	RET
	0x0081 00129 (<autogenerated>:1)	NOP
	0x0081 00129 (<autogenerated>:1)	CALL	runtime.morestack_noctxt(SB)
	0x0086 00134 (<autogenerated>:1)	JMP	0
	0x008b 00139 (<autogenerated>:1)	LEAQ	72(SP), DI
	0x0090 00144 (<autogenerated>:1)	CMPQ	(BX), DI
	0x0093 00147 (<autogenerated>:1)	JNE	38
	0x0095 00149 (<autogenerated>:1)	MOVQ	SP, (BX)
	0x0098 00152 (<autogenerated>:1)	JMP	38
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 72 48  dH..%....H;a.vrH
	0x0010 83 ec 40 48 89 6c 24 38 48 8d 6c 24 38 48 8b 59  ..@H.l$8H.l$8H.Y
	0x0020 20 48 85 db 75 65 0f 57 c0 0f 11 44 24 68 48 8b   H..ue.W...D$hH.
	0x0030 44 24 48 84 00 48 8b 4c 24 50 48 8b 40 20 48 89  D$H..H.L$PH.@ H.
	0x0040 0c 24 48 8b 4c 24 58 48 8b 54 24 60 48 89 4c 24  .$H.L$XH.T$`H.L$
	0x0050 08 48 89 54 24 10 ff d0 48 8b 44 24 20 48 8b 4c  .H.T$...H.D$ H.L
	0x0060 24 18 48 89 4c 24 28 48 89 44 24 30 48 89 4c 24  $.H.L$(H.D$0H.L$
	0x0070 68 48 89 44 24 70 48 8b 6c 24 38 48 83 c4 40 90  hH.D$pH.l$8H..@.
	0x0080 c3 e8 00 00 00 00 e9 75 ff ff ff 48 8d 7c 24 48  .......u...H.|$H
	0x0090 48 39 3b 75 91 48 89 23 eb 8c                    H9;u.H.#..
	rel 5+4 t=17 TLS+0
	rel 86+0 t=11 +0
	rel 130+4 t=8 runtime.morestack_noctxt+0
"".(*Student).sayGoodbye STEXT dupok size=209 args=0x28 locals=0x58
	0x0000 00000 (<autogenerated>:1)	TEXT	"".(*Student).sayGoodbye(SB), DUPOK|WRAPPER|ABIInternal, $88-40
	0x0000 00000 (<autogenerated>:1)	MOVQ	(TLS), CX
	0x0009 00009 (<autogenerated>:1)	CMPQ	SP, 16(CX)
	0x000d 00013 (<autogenerated>:1)	JLS	173
	0x0013 00019 (<autogenerated>:1)	SUBQ	$88, SP
	0x0017 00023 (<autogenerated>:1)	MOVQ	BP, 80(SP)
	0x001c 00028 (<autogenerated>:1)	LEAQ	80(SP), BP
	0x0021 00033 (<autogenerated>:1)	MOVQ	32(CX), BX
	0x0025 00037 (<autogenerated>:1)	TESTQ	BX, BX
	0x0028 00040 (<autogenerated>:1)	JNE	183
	0x002e 00046 (<autogenerated>:1)	NOP
	0x002e 00046 (<autogenerated>:1)	XORPS	X0, X0
	0x0031 00049 (<autogenerated>:1)	MOVUPS	X0, "".~r1+120(SP)
	0x0036 00054 (<autogenerated>:1)	CMPQ	""..this+96(SP), $0
	0x003c 00060 (<autogenerated>:1)	JNE	66
	0x003e 00062 (<autogenerated>:1)	NOP
	0x0040 00064 (<autogenerated>:1)	JMP	167
	0x0042 00066 (<autogenerated>:1)	MOVQ	""..this+96(SP), AX
	0x0047 00071 (<autogenerated>:1)	TESTB	AL, (AX)
	0x0049 00073 (<autogenerated>:1)	MOVQ	(AX), CX
	0x004c 00076 (<autogenerated>:1)	MOVQ	8(AX), AX
	0x0050 00080 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_4+48(SP)
	0x0055 00085 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_4+56(SP)
	0x005a 00090 (<autogenerated>:1)	MOVQ	CX, (SP)
	0x005e 00094 (<autogenerated>:1)	MOVQ	AX, 8(SP)
	0x0063 00099 (<autogenerated>:1)	MOVQ	"".name+104(SP), AX
	0x0068 00104 (<autogenerated>:1)	MOVQ	"".name+112(SP), CX
	0x006d 00109 (<autogenerated>:1)	MOVQ	AX, 16(SP)
	0x0072 00114 (<autogenerated>:1)	MOVQ	CX, 24(SP)
	0x0077 00119 (<autogenerated>:1)	CALL	"".Student.sayGoodbye(SB)
	0x007c 00124 (<autogenerated>:1)	MOVQ	32(SP), AX
	0x0081 00129 (<autogenerated>:1)	MOVQ	40(SP), CX
	0x0086 00134 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_3+64(SP)
	0x008b 00139 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_3+72(SP)
	0x0090 00144 (<autogenerated>:1)	MOVQ	AX, "".~r1+120(SP)
	0x0095 00149 (<autogenerated>:1)	MOVQ	CX, "".~r1+128(SP)
	0x009d 00157 (<autogenerated>:1)	MOVQ	80(SP), BP
	0x00a2 00162 (<autogenerated>:1)	ADDQ	$88, SP
	0x00a6 00166 (<autogenerated>:1)	RET
	0x00a7 00167 (<autogenerated>:1)	CALL	runtime.panicwrap(SB)
	0x00ac 00172 (<autogenerated>:1)	XCHGL	AX, AX
	0x00ad 00173 (<autogenerated>:1)	NOP
	0x00ad 00173 (<autogenerated>:1)	CALL	runtime.morestack_noctxt(SB)
	0x00b2 00178 (<autogenerated>:1)	JMP	0
	0x00b7 00183 (<autogenerated>:1)	LEAQ	96(SP), DI
	0x00bc 00188 (<autogenerated>:1)	NOP
	0x00c0 00192 (<autogenerated>:1)	CMPQ	(BX), DI
	0x00c3 00195 (<autogenerated>:1)	JNE	46
	0x00c9 00201 (<autogenerated>:1)	MOVQ	SP, (BX)
	0x00cc 00204 (<autogenerated>:1)	JMP	46
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 0f 86 9a  dH..%....H;a....
	0x0010 00 00 00 48 83 ec 58 48 89 6c 24 50 48 8d 6c 24  ...H..XH.l$PH.l$
	0x0020 50 48 8b 59 20 48 85 db 0f 85 89 00 00 00 0f 57  PH.Y H.........W
	0x0030 c0 0f 11 44 24 78 48 83 7c 24 60 00 75 04 66 90  ...D$xH.|$`.u.f.
	0x0040 eb 65 48 8b 44 24 60 84 00 48 8b 08 48 8b 40 08  .eH.D$`..H..H.@.
	0x0050 48 89 4c 24 30 48 89 44 24 38 48 89 0c 24 48 89  H.L$0H.D$8H..$H.
	0x0060 44 24 08 48 8b 44 24 68 48 8b 4c 24 70 48 89 44  D$.H.D$hH.L$pH.D
	0x0070 24 10 48 89 4c 24 18 e8 00 00 00 00 48 8b 44 24  $.H.L$......H.D$
	0x0080 20 48 8b 4c 24 28 48 89 44 24 40 48 89 4c 24 48   H.L$(H.D$@H.L$H
	0x0090 48 89 44 24 78 48 89 8c 24 80 00 00 00 48 8b 6c  H.D$xH..$....H.l
	0x00a0 24 50 48 83 c4 58 c3 e8 00 00 00 00 90 e8 00 00  $PH..X..........
	0x00b0 00 00 e9 49 ff ff ff 48 8d 7c 24 60 0f 1f 40 00  ...I...H.|$`..@.
	0x00c0 48 39 3b 0f 85 65 ff ff ff 48 89 23 e9 5d ff ff  H9;..e...H.#.]..
	0x00d0 ff                                               .
	rel 5+4 t=17 TLS+0
	rel 120+4 t=8 "".Student.sayGoodbye+0
	rel 168+4 t=8 runtime.panicwrap+0
	rel 174+4 t=8 runtime.morestack_noctxt+0
"".(*Student).sayHello STEXT dupok size=209 args=0x28 locals=0x58
	0x0000 00000 (<autogenerated>:1)	TEXT	"".(*Student).sayHello(SB), DUPOK|WRAPPER|ABIInternal, $88-40
	0x0000 00000 (<autogenerated>:1)	MOVQ	(TLS), CX
	0x0009 00009 (<autogenerated>:1)	CMPQ	SP, 16(CX)
	0x000d 00013 (<autogenerated>:1)	JLS	173
	0x0013 00019 (<autogenerated>:1)	SUBQ	$88, SP
	0x0017 00023 (<autogenerated>:1)	MOVQ	BP, 80(SP)
	0x001c 00028 (<autogenerated>:1)	LEAQ	80(SP), BP
	0x0021 00033 (<autogenerated>:1)	MOVQ	32(CX), BX
	0x0025 00037 (<autogenerated>:1)	TESTQ	BX, BX
	0x0028 00040 (<autogenerated>:1)	JNE	183
	0x002e 00046 (<autogenerated>:1)	NOP
	0x002e 00046 (<autogenerated>:1)	XORPS	X0, X0
	0x0031 00049 (<autogenerated>:1)	MOVUPS	X0, "".~r1+120(SP)
	0x0036 00054 (<autogenerated>:1)	CMPQ	""..this+96(SP), $0
	0x003c 00060 (<autogenerated>:1)	JNE	66
	0x003e 00062 (<autogenerated>:1)	NOP
	0x0040 00064 (<autogenerated>:1)	JMP	167
	0x0042 00066 (<autogenerated>:1)	MOVQ	""..this+96(SP), AX
	0x0047 00071 (<autogenerated>:1)	TESTB	AL, (AX)
	0x0049 00073 (<autogenerated>:1)	MOVQ	(AX), CX
	0x004c 00076 (<autogenerated>:1)	MOVQ	8(AX), AX
	0x0050 00080 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_4+48(SP)
	0x0055 00085 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_4+56(SP)
	0x005a 00090 (<autogenerated>:1)	MOVQ	CX, (SP)
	0x005e 00094 (<autogenerated>:1)	MOVQ	AX, 8(SP)
	0x0063 00099 (<autogenerated>:1)	MOVQ	"".name+104(SP), AX
	0x0068 00104 (<autogenerated>:1)	MOVQ	"".name+112(SP), CX
	0x006d 00109 (<autogenerated>:1)	MOVQ	AX, 16(SP)
	0x0072 00114 (<autogenerated>:1)	MOVQ	CX, 24(SP)
	0x0077 00119 (<autogenerated>:1)	CALL	"".Student.sayHello(SB)
	0x007c 00124 (<autogenerated>:1)	MOVQ	32(SP), AX
	0x0081 00129 (<autogenerated>:1)	MOVQ	40(SP), CX
	0x0086 00134 (<autogenerated>:1)	MOVQ	AX, ""..autotmp_3+64(SP)
	0x008b 00139 (<autogenerated>:1)	MOVQ	CX, ""..autotmp_3+72(SP)
	0x0090 00144 (<autogenerated>:1)	MOVQ	AX, "".~r1+120(SP)
	0x0095 00149 (<autogenerated>:1)	MOVQ	CX, "".~r1+128(SP)
	0x009d 00157 (<autogenerated>:1)	MOVQ	80(SP), BP
	0x00a2 00162 (<autogenerated>:1)	ADDQ	$88, SP
	0x00a6 00166 (<autogenerated>:1)	RET
	0x00a7 00167 (<autogenerated>:1)	CALL	runtime.panicwrap(SB)
	0x00ac 00172 (<autogenerated>:1)	XCHGL	AX, AX
	0x00ad 00173 (<autogenerated>:1)	NOP
	0x00ad 00173 (<autogenerated>:1)	CALL	runtime.morestack_noctxt(SB)
	0x00b2 00178 (<autogenerated>:1)	JMP	0
	0x00b7 00183 (<autogenerated>:1)	LEAQ	96(SP), DI
	0x00bc 00188 (<autogenerated>:1)	NOP
	0x00c0 00192 (<autogenerated>:1)	CMPQ	(BX), DI
	0x00c3 00195 (<autogenerated>:1)	JNE	46
	0x00c9 00201 (<autogenerated>:1)	MOVQ	SP, (BX)
	0x00cc 00204 (<autogenerated>:1)	JMP	46
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 0f 86 9a  dH..%....H;a....
	0x0010 00 00 00 48 83 ec 58 48 89 6c 24 50 48 8d 6c 24  ...H..XH.l$PH.l$
	0x0020 50 48 8b 59 20 48 85 db 0f 85 89 00 00 00 0f 57  PH.Y H.........W
	0x0030 c0 0f 11 44 24 78 48 83 7c 24 60 00 75 04 66 90  ...D$xH.|$`.u.f.
	0x0040 eb 65 48 8b 44 24 60 84 00 48 8b 08 48 8b 40 08  .eH.D$`..H..H.@.
	0x0050 48 89 4c 24 30 48 89 44 24 38 48 89 0c 24 48 89  H.L$0H.D$8H..$H.
	0x0060 44 24 08 48 8b 44 24 68 48 8b 4c 24 70 48 89 44  D$.H.D$hH.L$pH.D
	0x0070 24 10 48 89 4c 24 18 e8 00 00 00 00 48 8b 44 24  $.H.L$......H.D$
	0x0080 20 48 8b 4c 24 28 48 89 44 24 40 48 89 4c 24 48   H.L$(H.D$@H.L$H
	0x0090 48 89 44 24 78 48 89 8c 24 80 00 00 00 48 8b 6c  H.D$xH..$....H.l
	0x00a0 24 50 48 83 c4 58 c3 e8 00 00 00 00 90 e8 00 00  $PH..X..........
	0x00b0 00 00 e9 49 ff ff ff 48 8d 7c 24 60 0f 1f 40 00  ...I...H.|$`..@.
	0x00c0 48 39 3b 0f 85 65 ff ff ff 48 89 23 e9 5d ff ff  H9;..e...H.#.]..
	0x00d0 ff                                               .
	rel 5+4 t=17 TLS+0
	rel 120+4 t=8 "".Student.sayHello+0
	rel 168+4 t=8 runtime.panicwrap+0
	rel 174+4 t=8 runtime.morestack_noctxt+0
go.cuinfo.packagename. SDWARFINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
go.string."halfrost" SRODATA dupok size=8
	0x0000 68 61 6c 66 72 6f 73 74                          halfrost
go.string."everyone" SRODATA dupok size=8
	0x0000 65 76 65 72 79 6f 6e 65                          everyone
go.string."%v: Hello %v, nice to meet you.\n" SRODATA dupok size=32
	0x0000 25 76 3a 20 48 65 6c 6c 6f 20 25 76 2c 20 6e 69  %v: Hello %v, ni
	0x0010 63 65 20 74 6f 20 6d 65 65 74 20 79 6f 75 2e 0a  ce to meet you..
runtime.nilinterequal·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.nilinterequal+0
runtime.memequal64·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.memequal64+0
runtime.gcbits.01 SRODATA dupok size=1
	0x0000 01                                               .
type..namedata.*interface {}- SRODATA dupok size=16
	0x0000 00 00 0d 2a 69 6e 74 65 72 66 61 63 65 20 7b 7d  ...*interface {}
type.*interface {} SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 4f 0f 96 9d 08 08 08 36 00 00 00 00 00 00 00 00  O......6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*interface {}-+0
	rel 48+8 t=1 type.interface {}+0
runtime.gcbits.02 SRODATA dupok size=1
	0x0000 02                                               .
type.interface {} SRODATA dupok size=80
	0x0000 10 00 00 00 00 00 00 00 10 00 00 00 00 00 00 00  ................
	0x0010 e7 57 a0 18 02 08 08 14 00 00 00 00 00 00 00 00  .W..............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 24+8 t=1 runtime.nilinterequal·f+0
	rel 32+8 t=1 runtime.gcbits.02+0
	rel 40+4 t=5 type..namedata.*interface {}-+0
	rel 44+4 t=6 type.*interface {}+0
	rel 56+8 t=1 type.interface {}+80
type..namedata.*[]interface {}- SRODATA dupok size=18
	0x0000 00 00 0f 2a 5b 5d 69 6e 74 65 72 66 61 63 65 20  ...*[]interface 
	0x0010 7b 7d                                            {}
type.*[]interface {} SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 f3 04 9a e7 08 08 08 36 00 00 00 00 00 00 00 00  .......6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[]interface {}-+0
	rel 48+8 t=1 type.[]interface {}+0
type.[]interface {} SRODATA dupok size=56
	0x0000 18 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 70 93 ea 2f 02 08 08 17 00 00 00 00 00 00 00 00  p../............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[]interface {}-+0
	rel 44+4 t=6 type.*[]interface {}+0
	rel 48+8 t=1 type.interface {}+0
type..eqfunc.[2]interface {} SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type..eq.[2]interface {}+0
type..namedata.*[2]interface {}- SRODATA dupok size=19
	0x0000 00 00 10 2a 5b 32 5d 69 6e 74 65 72 66 61 63 65  ...*[2]interface
	0x0010 20 7b 7d                                          {}
type.*[2]interface {} SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 be 73 2d 71 08 08 08 36 00 00 00 00 00 00 00 00  .s-q...6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[2]interface {}-+0
	rel 48+8 t=1 type.[2]interface {}+0
runtime.gcbits.0a SRODATA dupok size=1
	0x0000 0a                                               .
type.[2]interface {} SRODATA dupok size=72
	0x0000 20 00 00 00 00 00 00 00 20 00 00 00 00 00 00 00   ....... .......
	0x0010 2c 59 a4 f1 02 08 08 11 00 00 00 00 00 00 00 00  ,Y..............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 02 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 type..eqfunc.[2]interface {}+0
	rel 32+8 t=1 runtime.gcbits.0a+0
	rel 40+4 t=5 type..namedata.*[2]interface {}-+0
	rel 44+4 t=6 type.*[2]interface {}+0
	rel 48+8 t=1 type.interface {}+0
	rel 56+8 t=1 type.[]interface {}+0
go.string."%v: Hi %v, see you next time.\n" SRODATA dupok size=30
	0x0000 25 76 3a 20 48 69 20 25 76 2c 20 73 65 65 20 79  %v: Hi %v, see y
	0x0010 6f 75 20 6e 65 78 74 20 74 69 6d 65 2e 0a        ou next time..
""..inittask SNOPTRDATA size=32
	0x0000 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 24+8 t=1 fmt..inittask+0
type..namedata.*func(string) string- SRODATA dupok size=23
	0x0000 00 00 14 2a 66 75 6e 63 28 73 74 72 69 6e 67 29  ...*func(string)
	0x0010 20 73 74 72 69 6e 67                              string
type.*func(string) string SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 9e c0 3b da 08 08 08 36 00 00 00 00 00 00 00 00  ..;....6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*func(string) string-+0
	rel 48+8 t=1 type.func(string) string+0
type.func(string) string SRODATA dupok size=72
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 4d fc a8 e7 02 08 08 33 00 00 00 00 00 00 00 00  M......3........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 01 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 00 00 00 00 00 00 00 00                          ........
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*func(string) string-+0
	rel 44+4 t=6 type.*func(string) string+0
	rel 56+8 t=1 type.string+0
	rel 64+8 t=1 type.string+0
runtime.interequal·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.interequal+0
type..namedata.*main.Person. SRODATA dupok size=15
	0x0000 01 00 0c 2a 6d 61 69 6e 2e 50 65 72 73 6f 6e     ...*main.Person
type.*"".Person SRODATA size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 91 18 70 14 08 08 08 36 00 00 00 00 00 00 00 00  ..p....6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*main.Person.+0
	rel 48+8 t=1 type."".Person+0
type..namedata.sayGoodbye- SRODATA dupok size=13
	0x0000 00 00 0a 73 61 79 47 6f 6f 64 62 79 65           ...sayGoodbye
type..namedata.sayHello- SRODATA dupok size=11
	0x0000 00 00 08 73 61 79 48 65 6c 6c 6f                 ...sayHello
type."".Person SRODATA size=112
	0x0000 10 00 00 00 00 00 00 00 10 00 00 00 00 00 00 00  ................
	0x0010 fd cf 91 06 07 08 08 14 00 00 00 00 00 00 00 00  ................
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 02 00 00 00 00 00 00 00 02 00 00 00 00 00 00 00  ................
	0x0050 00 00 00 00 00 00 00 00 20 00 00 00 00 00 00 00  ........ .......
	0x0060 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 24+8 t=1 runtime.interequal·f+0
	rel 32+8 t=1 runtime.gcbits.02+0
	rel 40+4 t=5 type..namedata.*main.Person.+0
	rel 44+4 t=5 type.*"".Person+0
	rel 48+8 t=1 type..importpath."".+0
	rel 56+8 t=1 type."".Person+96
	rel 80+4 t=5 type..importpath."".+0
	rel 96+4 t=5 type..namedata.sayGoodbye-+0
	rel 100+4 t=5 type.func(string) string+0
	rel 104+4 t=5 type..namedata.sayHello-+0
	rel 108+4 t=5 type.func(string) string+0
runtime.strequal·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.strequal+0
type..namedata.*main.Student. SRODATA dupok size=16
	0x0000 01 00 0d 2a 6d 61 69 6e 2e 53 74 75 64 65 6e 74  ...*main.Student
type..namedata.*func(*main.Student, string) string- SRODATA dupok size=38
	0x0000 00 00 23 2a 66 75 6e 63 28 2a 6d 61 69 6e 2e 53  ..#*func(*main.S
	0x0010 74 75 64 65 6e 74 2c 20 73 74 72 69 6e 67 29 20  tudent, string) 
	0x0020 73 74 72 69 6e 67                                string
type.*func(*"".Student, string) string SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 34 96 e5 08 08 08 08 36 00 00 00 00 00 00 00 00  4......6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*func(*main.Student, string) string-+0
	rel 48+8 t=1 type.func(*"".Student, string) string+0
type.func(*"".Student, string) string SRODATA dupok size=80
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 4d a8 19 06 02 08 08 33 00 00 00 00 00 00 00 00  M......3........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 02 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*func(*main.Student, string) string-+0
	rel 44+4 t=6 type.*func(*"".Student, string) string+0
	rel 56+8 t=1 type.*"".Student+0
	rel 64+8 t=1 type.string+0
	rel 72+8 t=1 type.string+0
type.*"".Student SRODATA size=104
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 0c 31 79 12 09 08 08 36 00 00 00 00 00 00 00 00  .1y....6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 02 00 00 00  ................
	0x0040 10 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0050 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0060 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*main.Student.+0
	rel 48+8 t=1 type."".Student+0
	rel 56+4 t=5 type..importpath."".+0
	rel 72+4 t=5 type..namedata.sayGoodbye-+0
	rel 76+4 t=25 type.func(string) string+0
	rel 80+4 t=25 "".(*Student).sayGoodbye+0
	rel 84+4 t=25 "".(*Student).sayGoodbye+0
	rel 88+4 t=5 type..namedata.sayHello-+0
	rel 92+4 t=25 type.func(string) string+0
	rel 96+4 t=25 "".(*Student).sayHello+0
	rel 100+4 t=25 "".(*Student).sayHello+0
type..namedata.*func(main.Student, string) string- SRODATA dupok size=37
	0x0000 00 00 22 2a 66 75 6e 63 28 6d 61 69 6e 2e 53 74  .."*func(main.St
	0x0010 75 64 65 6e 74 2c 20 73 74 72 69 6e 67 29 20 73  udent, string) s
	0x0020 74 72 69 6e 67                                   tring
type.*func("".Student, string) string SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 78 b3 48 66 08 08 08 36 00 00 00 00 00 00 00 00  x.Hf...6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*func(main.Student, string) string-+0
	rel 48+8 t=1 type.func("".Student, string) string+0
type.func("".Student, string) string SRODATA dupok size=80
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 0d 56 0a 12 02 08 08 33 00 00 00 00 00 00 00 00  .V.....3........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 02 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*func(main.Student, string) string-+0
	rel 44+4 t=6 type.*func("".Student, string) string+0
	rel 56+8 t=1 type."".Student+0
	rel 64+8 t=1 type.string+0
	rel 72+8 t=1 type.string+0
type..namedata.name- SRODATA dupok size=7
	0x0000 00 00 04 6e 61 6d 65                             ...name
type."".Student SRODATA size=152
	0x0000 10 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 da 9f 20 d4 07 08 08 19 00 00 00 00 00 00 00 00  .. .............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 01 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00  ................
	0x0050 00 00 00 00 02 00 00 00 28 00 00 00 00 00 00 00  ........(.......
	0x0060 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0070 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0080 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0090 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.strequal·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*main.Student.+0
	rel 44+4 t=5 type.*"".Student+0
	rel 48+8 t=1 type..importpath."".+0
	rel 56+8 t=1 type."".Student+96
	rel 80+4 t=5 type..importpath."".+0
	rel 96+8 t=1 type..namedata.name-+0
	rel 104+8 t=1 type.string+0
	rel 120+4 t=5 type..namedata.sayGoodbye-+0
	rel 124+4 t=25 type.func(string) string+0
	rel 128+4 t=25 "".(*Student).sayGoodbye+0
	rel 132+4 t=25 "".Student.sayGoodbye+0
	rel 136+4 t=5 type..namedata.sayHello-+0
	rel 140+4 t=25 type.func(string) string+0
	rel 144+4 t=25 "".(*Student).sayHello+0
	rel 148+4 t=25 "".Student.sayHello+0
go.itab."".Student,"".Person SRODATA dupok size=40
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 da 9f 20 d4 00 00 00 00 00 00 00 00 00 00 00 00  .. .............
	0x0020 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 type."".Person+0
	rel 8+8 t=1 type."".Student+0
	rel 24+8 t=1 "".(*Student).sayGoodbye+0
	rel 32+8 t=1 "".(*Student).sayHello+0
go.itablink."".Student,"".Person SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 go.itab."".Student,"".Person+0
type..importpath.fmt. SRODATA dupok size=6
	0x0000 00 00 03 66 6d 74                                ...fmt
"".Student.sayHello.stkobj SRODATA size=24
	0x0000 01 00 00 00 00 00 00 00 e0 ff ff ff ff ff ff ff  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 16+8 t=1 type.[2]interface {}+0
"".Student.sayGoodbye.stkobj SRODATA size=24
	0x0000 01 00 00 00 00 00 00 00 e0 ff ff ff ff ff ff ff  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
	rel 16+8 t=1 type.[2]interface {}+0