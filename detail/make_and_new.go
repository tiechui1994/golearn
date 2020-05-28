package main

/**
虽然 make 和 new 都是能够用于初始化数据结构, 但是它们两者能够初始化的结构类型却有着较大的不同;
make 在 Go 语言中只能用于初始化语言中的基本类型:

slice := make([]int, 0, 100)
hash := make(map[int]bool, 10)
channel := make(chan int, 5)

这三者返回了不同类型的数据结构:
slice 是一个包含 data, cap 和 len 的结构体;
hash 是一个指向 hmap 结构体的指针;
channel 是一个指向 hchan 结构体的指针;

----------------------------------------------------------------------------------------------------

而另一个用于初始化数据结构的关键字 new 的作用其实就非常简单, 它只是接收一个类型作为参数然后返回一个指向这个类型的指针:

i := new(int)

var v int
i := &v

上述代码片段中的两种不同初始化方法其实是等价的, 它们都会创建一个指向 int 零值的指针.

----------------------------------------------------------------------------------------------------

GOLANG MAKE & NEW
       +---> slice
       |
make --+---> hash
       |
       +---> channel

new  ------> pointer

----------------------------------------------------------------------------------------------------

原理:

GOLANG MAKE TYPECHECK:

OMAKE --> OMAKESLICE, OMAKEMAP, OMAKECHAN

在编译期间的 "类型检查" 阶段, Go语言其实就将代表make关键字的OMAKE节点根据参数类型的不同转换成了, OMAKESLICE,
OMAKEMAP 和 OMAKECHAN三种不同类型的节点, 这些节点最终会调用不同的运行时函数来初始化数据结构.


内置函数new会在编译机器的SSA代码生成阶段经过callnew函数的处理, 如果请求创建的类型大小是0, 那么就会返回一个表示空指针
的 zerobase 变量, 在遇到其他情况下会将关键字转换成 newobject:

func callnew(t *types.Type) *Node {
    if t.NotInHeap() {
        yyerror("%v is go:notinheap; heap allocation disallowed", t)
    }
    dowidth(t)

    if t.Size() == 0 {
        z := newname(Runtimepkg.Lookup("zerobase"))
        z.SetClass(PEXTERN)
        z.Type = t
        return typecheck(nod(OADDR, z, nil), ctxExpr)
    }

    fn := syslook("newobject")
    fn = substArgTypes(fn, t)
    v := mkcall1(fn, types.NewPtr(t), nil, typename(t))
    v.SetNonNil(true)
    return v
}

注: 哪怕当前变量是使用 var 进行初始化, 在这一阶段可能会被转换成 newobject 的函数调用并在堆上申请内存.

func walkstmt(n *Node) *Node {
    switch n.Op {
    case ODCL:
        v := n.Left
        if v.Class() == PAUTOHEAP {
            if prealloc[v] == nil {
                prealloc[v] = callnew(v.Type)
            }
            nn := nod(OAS, v.Name.Param.Heapaddr, prealloc[v])
            nn.SetColas(true)
            nn = typecheck(nn, ctxStmt)
            return walkstmt(nn)
        }
    case ONEW:
        if n.Esc == EscNone {
            r := temp(n.Type.Elem())
            r = nod(OAS, r, nil)
            r = typecheck(r, ctxStmt)
            init.Append(r)
            r = nod(OADDR, r.Left, nil)
            r = typecheck(r, ctxExpr)
            n = r
        } else {
            n = callnew(n.Type.Elem())
        }
    }
}

newobject 函数的工作就是获取传入类型的大小并调用 mallocgc 在堆上申请一片大小合适的内存空间并返回指向这片内存
空间的指针.
**/
