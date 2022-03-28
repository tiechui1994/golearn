## C 常用函数对比

### 内存分配

malloc, calloc, realloc, memset, free

- malloc

```
#include <stdlib.h>
#include <malloc.h>
void* malloc (size_t size)
```

- calloc

```
#include <stdlib.h>
#include <malloc.h>
void* calloc(size_t n, size_t size)
```

分配 n 个长度为 size 的连续空间.

- realloc

```
#include <stdlib.h>
#include <malloc.h>
void* realloc(void* mem_address, size_t newsize);
```

对已有内存的变量重新分配新的内存大小(可大,可小). 先判断 mem_address 是否有足够的连续空间, 如果有, 扩大 mem_address
指向的地址, 并且将 mem_address 返回. 如果空间不足, 先按照 newsize 指定大小分配空间, 然后将原有数据从头到尾拷贝到新
分配的内存区域, 然后释放原来 mem_address 的内存区域.

- memset

```
#include <stdlib.h>
#include <malloc.h>
void* memset(void* addr, int c, size_t n);
```

复制字符 c 到参数 addr 所指定的内存区域的前 n 个长度.

### 内存拷贝

memcpy, memmove


- memcpy

```
#include <string.h>
void *memcpy(void* dest, const void* src, size_t n);
```

从 src 存储区复制 n 个字节到 dest.

- memmove

```
#include <string.h>
void *memmove(void* dest, const void* src, size_t n);
```

从 src 存储区复制 n 个字节到 dest. 

**memcpy vs memmove**

memmove() 更加灵活, 当 src 和 dest 所指定的内存区域重叠时, memmove() 可以正确处理, 但是执行效率上会比 memcpy() 略
慢些.


### C 中关于 static 的作用

C 中, static 的作用有三条: 一是隐藏功能, 二是保持持久性功能, 三是默认初始化为0

隐藏功能, 对于 static 修饰的函数和全局变量而言.

保持持久性功能, 对于 static 修饰的局部变量而言.

默认初始化为 0, 全局和局部的 static 修饰的变量而言.

1. 当同时编译多个文件时, 所有未加 static 前缀的全局变量和函数都具有全局可见性.

```cgo
// a.c
#include <stdio.h>

char a = 'A';

void message() {
    printf("hello\n");
}


// main.c
#include <stdio.h>

int main() {
    extern char a; // extern, declared before used.
    printf("%c ", a);
    message();
    
    return 0;
}
```

输出的结果是 `A hello`. 由于在 a.c 当中的 `全局变量 a` 和 `函数 message` 未加 static 前缀, 因此都具有全局可见性,
其它的源文件也能访问. 因此, 在 main.c 当中是可以获取到a的值, 以及执行 message 函数.

如果加了 static, 就会对其它源文件因此. 例如, 如果 a.c 当中的 `全局变量 a` 和 `函数 message` 在定义时加上 static 前
缀, 则在 mian.c 当中就无法访问到这些变量. 利用这一特性, 可以在不同的文件当中定义同名函数和同名变量, 而不必担心命名冲突.

2. 保持变量内存的持久性. 存储在静态数据区的变量在程序运行时旧完成了初始化, 也是唯一的一次初始化. 有两者变量存储在静态存储
区: 全局变量和static变量. 

```cgo
// main.c

#include <stdio.h>

int fun() {
    static int count = 10; // 该赋值只会在程序启动时执行一次, 之后再也不会执行了.
    return count--; // 输出的值依次递减, 10, 9, ....
}

int count=1;
int main() {
    for (; count <=10; count++) {
        printf("%d   %d\n", count, fun())
    }
    
    return 0;
}
```

3. 默认初始化为0. 全局变量也具有这一属性, 因为全局变量也存储在静态数据区. 在静态数据区, 内存所有的字节默认值都是 0x00.
