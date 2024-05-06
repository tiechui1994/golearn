# go defer 原理

执行 defer 语句之后, 是注册记录一个稍后执行的函数. 把函数名和参数确定下来, 不会立即调用, 而是等到从当前函数 return 
出去的时候.

例子:

```cgo
func main() {
    var a int
    defer func(i int) {
        println("defer=", i)
    }(a)

    println("main=", a)
}
```

通过汇编, 其代码如下: `go tool compile -N -l -S main.go`, go1.16

```
"".main STEXT size=197 args=0x0 locals=0x70 funcid=0x0
    0x0000 00000 (main.go:3)    TEXT    "".main(SB), ABIInternal, $112-0
    0x0000 00000 (main.go:3)    MOVQ    (TLS), CX
    0x0009 00009 (main.go:3)    CMPQ    SP, 16(CX)
    0x000d 00013 (main.go:3)    JLS     182
    0x0013 00019 (main.go:3)    SUBQ    $112, SP
    0x0017 00023 (main.go:3)    MOVQ    BP, 104(SP)
    0x001c 00028 (main.go:3)    LEAQ    104(SP), BP
    0x0021 00033 (main.go:4)    MOVQ    $0, "".a+16(SP)          # a 变量初始化
    0x002a 00042 (main.go:5)    MOVL    $8, ""..autotmp_1+24(SP) # _defer.siz 
    0x0032 00050 (main.go:5)    LEAQ    "".main.func1·f(SB), AX
    0x0039 00057 (main.go:5)    MOVQ    AX, ""..autotmp_1+48(SP) # _defer.fn
    0x003e 00062 (main.go:5)    MOVQ    "".a+16(SP), AX
    0x0043 00067 (main.go:5)    MOVQ    AX, ""..autotmp_1+96(SP) # 拷贝 fn 参数到 _defer
    0x0048 00072 (main.go:5)    LEAQ    ""..autotmp_1+24(SP), AX
    0x004d 00077 (main.go:5)    MOVQ    AX, (SP)                 # 调整栈 72(96-24)
    0x0051 00081 (main.go:5)    CALL    runtime.deferprocStack(SB) # 注册 _defer 到 g._defer
    0x0056 00086 (main.go:5)    TESTL   AX, AX
    0x0058 00088 (main.go:5)    JNE     166 # 注册失败
    0x005a 00090 (main.go:5)    JMP     96  # 注册成功 AX=0
    0x005c 00092 (main.go:5)    NOP
    0x0060 00096 (main.go:9)    CALL    runtime.printlock(SB)
    0x0065 00101 (main.go:9)    LEAQ    go.string."main= "(SB), AX
    0x006c 00108 (main.go:9)    MOVQ    AX, (SP)
    0x0070 00112 (main.go:9)    MOVQ    $6, 8(SP)
    0x0079 00121 (main.go:9)    CALL    runtime.printstring(SB)
    0x007e 00126 (main.go:9)    MOVQ    "".a+16(SP), AX
    0x0083 00131 (main.go:9)    MOVQ    AX, (SP)
    0x0087 00135 (main.go:9)    CALL    runtime.printint(SB)
    0x008c 00140 (main.go:9)    CALL    runtime.printnl(SB)
    0x0091 00145 (main.go:9)    CALL    runtime.printunlock(SB)
    0x0096 00150 (main.go:10)    XCHGL   AX, AX
    0x0097 00151 (main.go:10)    CALL    runtime.deferreturn(SB) # 执行 deferreturn
    0x009c 00156 (main.go:10)    MOVQ    104(SP), BP
    0x00a1 00161 (main.go:10)    ADDQ    $112, SP
    0x00a5 00165 (main.go:10)    RET
    0x00a6 00166 (main.go:5)    XCHGL   AX, AX
    0x00a7 00167 (main.go:5)    CALL    runtime.deferreturn(SB)
    0x00ac 00172 (main.go:5)    MOVQ    104(SP), BP
    0x00b1 00177 (main.go:5)    ADDQ    $112, SP
    0x00b5 00181 (main.go:5)    RET
    0x00b6 00182 (main.go:5)    NOP
    0x00b6 00182 (main.go:3)    CALL    runtime.morestack_noctxt(SB)
    0x00bb 00187 (main.go:3)    NOP
    0x00c0 00192 (main.go:3)    JMP     0
```

关于 _defer 的数据结构:

```
type _defer struct {
    siz     int32 // 函数 fn 的 arguments 与 results 大小
    started bool  
    heap    bool  // 是否在堆上分配, 调用 deferprocStack 值是 false
    // openDefer indicates that this _defer is for a frame with open-coded
    // defers. We have only one defer record for the entire frame (which may
    // currently have 0, 1, or more defers active).
    openDefer bool
    sp        uintptr  // sp at time of defer
    pc        uintptr  // pc at time of defer
    fn        *funcval // can be nil for open-coded defers
    _panic    *_panic  // panic that is running defer
    link      *_defer  // 链表结构

    // If openDefer is true, the fields below record values about the stack
    // frame and associated function that has the open-coded defer(s). sp
    // above will be the sp for the frame, and pc will be address of the
    // deferreturn call in the function.
    fd   unsafe.Pointer // funcdata for the function associated with the frame
    varp uintptr        // value of varp for the stack frame
    // framepc is the current pc associated with the stack frame. Together,
    // with sp above (which is the sp associated with the stack frame),
    // framepc/sp can be used as pc/sp pair to continue a stack trace via
    // gentraceback().
    framepc uintptr
}
```

_defer 结构只是一个 header, 紧跟 header 后面的是参数和返回值, 大小由 _defer.siz 指定. **这块内存的值所在 defer
注册的时候就已经填充好了**, _defer 的内存布局:

![image](/images/develop_defer_mem.png)

### 栈上分配 defer (deferprocStack)

// runtime.deferprocStack, 将 defer 挂载到 g._defer 上

```cgo
// 入参 _defer 是在栈上已经预先分配
func deferprocStack(d *_defer) {
    gp := getg()
    if gp.m.curg != gp {
        // go code on the system stack can't defer
        throw("defer on system stack")
    }
    if goexperiment.RegabiDefer && d.siz != 0 {
        throw("defer with non-empty frame")
    }
    // siz, fn 提前已经设置好了, 剩下其他参数初始化
    d.started = false
    d.heap = false  // 表示为栈内存
    d.openDefer = false
    d.sp = getcallersp() // 获取 caller rsp 寄存器值
    d.pc = getcallerpc() // 获取 caller rip 寄存器值, 这里的值就是调用 deferprocStack 的下一条指令地址
    d.framepc = 0
    d.varp = 0

    // 将 _defer 挂载到 g 上
    *(*uintptr)(unsafe.Pointer(&d._panic)) = 0
    *(*uintptr)(unsafe.Pointer(&d.fd)) = 0
    *(*uintptr)(unsafe.Pointer(&d.link)) = uintptr(unsafe.Pointer(gp._defer))
    *(*uintptr)(unsafe.Pointer(&gp._defer)) = uintptr(unsafe.Pointer(d))
  
    // return0 汇编实现, 将 AX 设置为 0
    return0()
}
```

小结:

1) 由于是栈上分配内存, 因此在调用 `deferprocStack` 之前, 编译器已经将 _defer 结构的内存已经分配好, 并且填充了 siz,
参数, 返回值

2) _defer.heap 标识此结构体是分配在栈上, 保存 caller 的 rsp, rip 寄存器值保存到 _defer 当中

3) _defer 作为一个节点挂载到链表, 链表头是 g._defer, 在一个 goroutine 当中大部分有多次函数调用, 所以这个链表会挂
接一个调用栈上的 _defer 结构, 执行的时候按照 rsp 来过滤区分;

![image](/images/develop_defer_link.png)

### 执行 defer 链 (deferreturn)

编译器遇到 defer 语句, 会插入两种函数:

1. 分配函数: `deferprocStack` 或 `deferproc`

2. 执行函数: `deferreturn` (N+1, N 表示 defer 个数, 执行 `deferprocStack` 或 `deferproc` 的跳转函数, 1 是
最终执行 defer 链的执行函数)

// deferreturn 执行 defer 链. 注意的是 deferreturn 没有任何返回值与参数, 无需修改栈帧
```cgo
func deferreturn() {
    gp := getg()
    d := gp._defer
    if d == nil {
        return
    }
    
    // 通过 _defer.sp 与 sp 比较确定栈帧是否一致
    sp := getcallersp()
    if d.sp != sp {
        return
    }
    
    // 特殊的 defer
    if d.openDefer {
        done := runOpenDeferFrame(gp, d)
        if !done {
            throw("unfinished open-coded defers in deferreturn")
        }
        gp._defer = d.link
        freedefer(d)
        return
    }

    // 参数移动, 保证栈是 nosplit
    argp := getcallersp() + sys.MinFrameSize
    switch d.siz {
    case 0:
        // Do nothing.
    case sys.PtrSize:
        // deferArgs(d) 获取 d 的第一个参数指针
        *(*uintptr)(unsafe.Pointer(argp)) = *(*uintptr)(deferArgs(d))
    default:
        memmove(unsafe.Pointer(argp), deferArgs(d), uintptr(d.siz))
    }
    fn := d.fn
    d.fn = nil
    gp._defer = d.link // 修改链表头
    freedefer(d) // 释放 _defer
    // 确保 fn 是非空指针
    _ = fn.fn
    
    // jmpdefer 汇编实现函数跳转
    // argp 是 caller 的栈 rsp
    jmpdefer(fn, argp)
}
```


// 跳转函数 jmpdefer
```
// func jmpdefer(fv *funcval, argp uintptr)
// argp is a caller SP.
// called from deferreturn.
// 1. pop the caller
// 2. sub 5 bytes from the callers return
// 3. jmp to the argument
TEXT runtime·jmpdefer(SB), NOSPLIT, $0-16
    MOVQ    fv+0(FP), DX    // 取出延迟回调函数 fn 地址
    MOVQ    argp+8(FP), BX    // caller 函数的 rsp
    LEAQ    -8(BX), SP        // rsp 的值设置成 caller rsp 值往下 8 字节(即 'CALL runtime.deferreturn(SB)', 
                            // 压入栈的值, 即 rip)
    MOVQ    -8(SP), BP        // 恢复 BP 寄存器地址
    SUBQ    $5, (SP)        // 修改此时 rsp 栈顶的值, 这个值又恢复成了 'CALL runtime.deferreturn(SB)' 的值(
                            // 这个值也是函数返回地址, 相当于又一次去执行 deferreturn, 非常的巧妙)
    MOVQ    0(DX), BX
    JMP    BX    // but first run the deferred function
```

上述代码实现了 `caller -> [ deferreturn -> jmpdefer -> defered func -> deferreturn ] 的递归循环`, 直到 g._defer
为空结束.

![image](/images/develop_defer_call.png)


这个相当于编译器手动管理了函数栈帧, 通过修改栈上的值, 让延迟调用函数执行完调用 ret 的时候, 重新跳转到 deferreturn 
函数进行循环执行.(这是一种 hack 手段, 修改函数压栈的值, 跳转到一些 hack 的指令上去执行代码)

