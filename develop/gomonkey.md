# GoMonkey 实现原理

打桩函数的原理是将原始函数指针指向的内容修改为跳转指令, 当执行原始函数时, 会先执行跳转指令(也就跳转到设置的函数地址
位置, 从而达到 Hook 的效果)

### 代码实现

// 核心函数, 使用 double 来 Hook target.
// 注: target 是函数指针, double 是当前函数实现
```
func (this *Patches) ApplyCore(target, double reflect.Value) *Patches {
    this.check(target, double)
    // 要细品这里 assTarget 的含义.
    // *(*uintptr)(getPointer(target)) .text 段当中的地址, 编译的机码位置.
    // (getPointer(target))            .rodata 段当中的地址. 
    assTarget := *(*uintptr)(getPointer(target)) 
    original := replace(assTarget, uintptr(getPointer(double)))
    if _, ok := this.originals[assTarget]; !ok {
        this.originals[assTarget] = original
    }
    this.valueHolders[double] = double
    return this
}
```

> 关于内存的分段的介绍, 参考 **内存区段**


// 获取函数指针指向底层的地址

```
type funcValue struct {
    _ uintptr
    p unsafe.Pointer
}

func getPointer(v reflect.Value) unsafe.Pointer {
    return (*funcValue)(unsafe.Pointer(&v)).p
}
```

// 指针指向数据替换操作.
```
func replace(target, double uintptr) []byte {
    code := buildJmpDirective(double)        // 需要替换的数据 
    bytes := entryAddress(target, len(code)) // 获取原生 target 指向的数据
    original := make([]byte, len(bytes))
    copy(original, bytes)
    modifyBinary(target, code)               // 将原生 target 指向的数据进行替换
    return original
}

// 构建跳跃指令
func buildJmpDirective(double uintptr) []byte {
    d0 := byte(double)
    d1 := byte(double >> 8)
    d2 := byte(double >> 16)
    d3 := byte(double >> 24)
    d4 := byte(double >> 32)
    d5 := byte(double >> 40)
    d6 := byte(double >> 48)
    d7 := byte(double >> 56)
    
    return []byte{
        0x48, 0xBA, d0, d1, d2, d3, d4, d5, d6, d7, // MOV rdx, double
        0xFF, 0x22,                                 // JMP [rdx]
    }
}

// 获取指针 p 指向的真实数据
func entryAddress(p uintptr, l int) []byte {
    return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: p, Len: l, Cap: l}))
}

// 修改内存 
func modifyBinary(target uintptr, bytes []byte) {
    function := entryAddress(target, len(bytes))
    err := mprotectCrossPage(target, len(bytes), syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC)
    if err != nil {
        panic(err)
    }
    copy(function, bytes)
    err = mprotectCrossPage(target, len(bytes), syscall.PROT_READ|syscall.PROT_EXEC)
    if err != nil {
        panic(err)
    }
}
func mprotectCrossPage(addr uintptr, length int, prot int) error {
    pageSize := syscall.Getpagesize()
    for p := pageStart(addr); p < addr+uintptr(length); p += uintptr(pageSize) {
        page := entryAddress(p, pageSize)
        if err := syscall.Mprotect(page, prot); err != nil {
            return err
        }
    }
    return nil
}
```

### 内存区段

系统内的程序分段, 具体细分又可以分:

- text, 存储机器码, 运行前已经确定(编译时确定), 通常为只读, 可以直接在 ROM 或 Flash 中执行, 无需加载到 RAM

- rodata, 存储常量数据, 只读数据, 存储在 ROM 中. 

- data, 存储已经初始化的全局变量, 属于静态内存分配. (注意: 初始化为 0 的全局变量还是被保持在 bss 中)

- bss, 存储没有初值的全局变量或默认为0的全局变量, 属于静态内存分配. bss 不占执行文件空间(无需加入程序之中, 只要链接
时寻址到 RAM 即可), 但占程序运行时的内存空间

- stack, 存储参数和局部变量. 由系统进行申请和释放, 属于静态内存分配. 

stack 的特点是先进先出, 可用于保存/恢复调用现场.

- heap, 存储程序运行过程中被动态分配的内存, 由用户申请和释放. 

heap 申请时是分配虚拟内存, 当真正存储数据时才分配物理内存;
