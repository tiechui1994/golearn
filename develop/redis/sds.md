## Redis - sds 

sds(Simple Dymamic String), 特点:

- 可动态扩展内存. sds 标示的字符串其内容可以修改, 也可以追加. 

- 二进制安全. sds 能存储任意二进制数据, 而不仅仅是可打印字符串.

- 与传统的C语言字符串兼容.


sds 数据结构定义:

C 语言当中, 字符串都是以 '\0' (0x00, NULL结束符) 字符结尾的字符数组来存储的, 通常的表达为字符指针(char*), 它不允许
字节0出现在字符串中间, 因此它不能表示任意二进制数据.

sds 定义
```cgo
typedef char* sds;

typedef __attribute__((__packed__)) sdshdr5 {
    unsigned char flags;
    char buf[];
}

typedef __attribute__((__packed__)) sdshdr8 {
    uint8_t len;
    uint8_t alloc;
    unsigned char flags;
    char buf[];
}

typedef __attribute__((__packed__)) sdshdr16 {
    uint16_t len;
    uint16_t alloc;
    unsigned char flags;
    char buf[];
}

typedef __attribute__((__packed__)) sdshdr32 {
    uint32_t len;
    uint32_t alloc;
    unsigned char flags;
    char buf[];
}

typedef __attribute__((__packed__)) sdshdr64 {
    uint64_t len;
    uint64_t alloc;
    unsigned char flags;
    char buf[];
}
```

sds 和 char* 是不等同. sds的二进制安全, 可以存储任意二进制数据, 不能像C语言字符串那样以字符 '\0' 来标识字符串的结束,
因此它必然有个长度字段. 实际上, sds 还包含一个 header 结构(dshdr5, sdshdr8, sdshdr16, sdshdr32, sdshdr64, 一
共有5种). 之所以有5种 header, 是为了让不同长度的字符串可以可以使用不同大小的header, 节省内存.

一个完整的 sds 字符串的完成结构, 由在内存地址上前后相邻的两部分组成:

- 一个 header. 通常包含len(真正的长度, 不包含NULL结束符在内), alloc(字符串最大容量, 不包含最后那个多余的字节) 和 flags. 
sdshdr5比较特殊.

- 一个字符数组. 这个字符数组的长度等于最大容量+1. 真正有效的字符串数据, 其长度通常小于最大容量. 在真正的字符串数据之后,
是空余未用的字节(一般使用 0x00 填充). 在真正的字符串之后, 还有一个 NULL(0x00) 结束符.

> flags 只使用了三个bit, 表示当前的 SDS 的类型(0, 1, 2, 3, 4). 对于 sdshdr5, flags 的高5位表示 sds 的长度, 最
多保存32个字节. sdshdr5 是不支持动态扩展.
>
> 在各个 header 定义当中使用了 `__attribute__((packed))`, 是为了让编译器以紧凑模式分配内存. 如果没有这个熟悉, 编
译器可能会为 struct 的字段做优化对齐, 在其中填充空字节. 如果那样的话, 就不能保证 header 和 sds 的数据是部分是紧紧前后
相邻, 也不能按照固定向低地址方向偏移1字节来获取 flags 字段了.
>
> 在各个 header 定义中最后有一个 char buf[]. 这只是起到一个标记作用, 表示flags后面是一个字符数组. **而程序在为header
分配内存时, 它并不占用内存空间. 在计算  sizeof(struct sdshdr16)的值, 结果是5**.

sds 与 string 的关系:

setbit 和 getrange 是先根据 key 获取整个整个 sds 字符串, 然后从字符串选取或修改指定的部分. 

当存储的值是数字的时候, 它还支持 incr, decr 操作, 此时它的内部结构不再是 sds, 这种状况下 setbit 和 getrange 是根据
数值的二进制进行操作的.(在 robj 当中会有详细的说明)
