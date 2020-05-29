# 编译

## 条件编译

条件编译 (condition compile) 命令指定预处理器依据特定的条件来判断保留或删除某段源代码. 例如, 可以使用条件编译
让源代码适用不同的目标系统, 而不需要管理源代码的各种不同版本.

条件编译区域以 **`#if`, `#ifdef`, `#ifndef`** 等命令开头, 以 **`#endif`** 命令结尾. 条件编译区域可以有任
意数量的 **`#elif`** 命令, 但最多一个 **`#else`** 命令.

```c
#if 表达式1
  [ 组1]
[#elif 表达式2
  [ 组2]]
...
[#elif 表达式n
  [ 组n ]]
[#else
  [ 组n+1 ]]
#endif
```

> 预处理器会依次计算条件表达式, 直到发现结果非0的条件表达式. 预处理器会保留对应组内的源代码, 以供后续处理.
> 如果找不到值为true的表达式, 并且该条件编译区域中包含 `#else` 命令, 则保留 `#else` 命令组内的代码.
 

- `#if` 和 `#elif` 命令 

**作为 `#if` 或 `#elif` 命令条件的表达式, 必须是整数常量预处理表达式.** 这与 `普通的常量表达式` 不同, 主要
区别在于:

1) 不能在 `#if` 或 `#elif` 表达式中使用类型转换运算符.

2) 可以使用预处理运算符 `defined`

3) 在预处理器展开所有宏, 并且计算完所有 `defined` 表达式之后, 会使用字符 `o` 替换掉表达式中所有其他标识符或
关键字.

4) 表达式中所有带符号值都具有 `intmax_t` 类型, 并且所有无符号值都具有 `uintmax_t` 类型. 字符常量也会受该规
则的影响. `intmax_t` 和 `uintmax_t` 定义在头文件 `stdint.h` 中.


- defined 预算符

一元运算符号 `defined` 可以出现在 `#if` 或 `#elif` 命令的条件中.

```
defined identifier

defined (identifier)
``` 

> 如果指定 `identifier` 是一个宏名称 (即, 它已经被 `#define` 命令定义, 并且未被 `#undef` 命令取消定义), 则
> defined 表示会生成值 1, 否则, defined 表达式会生成值 0.

```c
#if defined(__unix__) && defined(__GNUC__)
/* ... */
#endif
```

大多数编译器会提供预定义宏.


## gcc(g++) 编译选项

- `-Wall`, 生成所有警告信息.

---

- `-E`, 产生预处理阶段的输出.

```
gcc -E main.c > main.i
```

- `-S`, 产生汇编阶段的代码

```
gcc -S main.c > main.s
```

- `-c`, 只产生编译的代码(没有链接link)

```
gcc -c main.c
```

> 上面代码产生 `main.o`, 包括机器级别的代码或者编译的代码


- `-save-temps`, 产生所有的中间步骤的文件

```
gcc -save-temps main.c
``` 

> 上面的代码将产生文件 `main.i`, `main.s`, `main.o`, `a.out`, 其中 `a.out` 是可执行文件 

---


- `-l`, 指定链接共享库.

```
gcc -Wall main.c -o main -l CPPfile
```

> 上面的代码会链接 `libCPPfile.so`, 产生可执行文件 `main`.


- `-fPIC`, 产生位置无关的代码. 当产生 `共享库` 的时候, 应该创建位置无关的代码, 这会让共享库使用任意的地
址而不是固定的地址, 要实现这个功能, 需要使用 `-fPIC` 参数. 概念上就是可执行程序装载它们的时候, 它们可以放
置可执行程序的内存里的任何地方.

```
gcc -c -Wall -Werror -fPIC Cfile.c
gcc -shared -o libCfile.so Cfile.o
```

- `-V`, 打印所有的执行命令. (打印出 gcc 编译一个文件的时候所有的步骤) 


- `-ansi`, 支持 ISO C89 程序.

- `-funsigned-char`, `char` 类型被看作为 `unsigned char` 类型.

- `-fsigned-char`, `char` 类型被看作为 `signed char` 类型.


- `-D`, 可以使用编译时的宏

```c
#include <stdio.h>

int main() {
#ifdef MY_MACRD
    printf("\n Macro defined \n");
#endif
    char c = -10;
    printf("\n char [%d] \n", c);
    return 0;
}
```

```
$ gcc -Wall -D MY_MACRD main.c -o main
$ ./main
```


- `-Werror`, 将所有的警告转换为错误信息.


- `-I`, 指定头文件的文件夹

```
gcc -I /home/user/include input.c
```

- `-L`, 表示要链接的库所在目录


- `-std`, 指定支持的 C++/C 的标准

```
gcc -std=c++11 main.cpp
```

> 标准如 `c++11`, `c++14`, `c90`, `c89` 等.


- `-static`, 生成静态连接的文件. 静态编译文件(把动态库的函数和其他依赖都编译进最终文件)

```
gcc main.c -static -o main -l pthread
```

- `-shared`, 使用动态库链接.

- `-static-libstdc++`, 静态链接 `libstdc++`. 如果没有使用 `-static`, 默认使用 `libstdc++` 共
享库, 而 `-static-libstdc++` 可以指定使用 `libstdc++` 静态库.


- `-Wl,options`, 把参数(options) 传递给链接库ld, 如果options中间有逗号, 就将options分成多个选项,
然后传递给链接程序.

## GCC 编译器的环境变量

- PATH

在 `PATH` 中找到可执行文件程序的路径


- C_INCLUDE_PATH

gcc 找到头文件的路径  


- CPLUS_INCLUDE_PATH

g++ 找到头文件的路径


- LD_LIBRARY_PATH

动态链接库的路径


- LIBRARY_PATH

静态链接库的路径

库文件在链接(静态库和共享库) 和 运行 (仅限于使用共享库的程序) 时被使用, 其搜索路径是在系统中进行设置的. 一般 Linux 
系统把 `/lib` 和 `/usr/lib` 两个目录作为默认的库搜索路径, 所以使用这两个目录的库时不需要设置搜索路径即可使用. 对
于处于默认库搜索路径之外的库, 需要将库的位置添加到库的搜索路径之中. 

设置库文件的搜索路径有以下两种方式:

方式一: 在环境变量 LD_LIBRARY_PATH 中指明动态库的搜索路径

方式二: 在/etc/ld.so.conf 文件当中添加库的搜索路径