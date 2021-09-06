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

