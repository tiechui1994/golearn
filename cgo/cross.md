## Go 交叉编译

> 说明: 以下的交叉编译主机是在 x86_64 Ubuntu 16.04 平台下进行的.

Go 交叉编译涉及的编译参数:

- `GOARCH`, 目标平台的 CPU 架构. 常用的值 `amd64`, `arm64`, `i386`, `armhf`

- `GOOS`, 目标平台, 常用的值 `linux`, `windows`, `drawin` (macOS)

- `GOARM`, 只有 `GOARCH` 是 `arm64` 才有效, 表示 `arm` 的版本, 只能是 5, 6, 7 其中之一

- `CGO_ENABLED`, 是否支持 CGO 交叉汇编, 值只能是 `0`, `1`, 默认情况下是 `0`, 启用交叉汇编比较麻烦

- `CC`, 当支持交叉汇编时(即 `CGO_ENABLED=1`), 编译目标文件使用的 `c` 编译器. 

- `CXX`, 当支持交叉汇编时(即 `CGO_ENABLED=1`), 编译目标文件使用的 `c++` 编译器. 

- `AR`, 当支持交叉汇编时(即 `CGO_ENABLED=1`), 编译目标文件使用的创建库文件命令.


### 不支持 CGO

不支持 CGO, 即 CGO_ENABLED=0, 在这种状况下, 进行交叉汇编是比较简单的. 只需要设置 `GOARCH`, `GOOS`, `GOARM` (只
有是 arm 架构的平台) 即可.

交叉汇编 windows 系统 amd64 架构的目标文件:

```bash
GOOS=windows GOARCH=amd64 go build -o xxx *.go
```

交叉汇编 drawin 系统 amd64 架构的目标文件:

```bash
GOOS=drawin GOARCH=amd64 go build -o xxx *.go
```

交叉汇编 linux 系统 arm64 架构的目标文件:

```bash
GOOS=linux GOARCH=arm64 GOARM=7 go build -o xxx *.go
```

其他架构的汇编可以进行类比.


### 支持 CGO 

支持 CGO, 即 CGO_ENABLED=1, 在这种状况下, 进行交叉汇编有点复杂. 除了设置必要的参数`GOARCH`,`GOOS`,`GOARM`(只有
是arm架构的平台),`CGO_ENABLED`之外, 还需要设置`CC`,`CXX`,`AR`参数.

这里主要介绍一下, 交叉汇编 arm 架构下的目标文件.

首先, arm 架构目前主要分为四种, 分别是`armhf`(arm hard float, 硬件浮点),`arm64`(64位的arm默认就是hf的).`eabi`
(embedded applicaion binary interface, 嵌入式二进制接口)和`armel`(arm eabi little endian, 软件浮点).

下面是 arm 交叉汇编工具 (Ubuntu下):

| tool | armhf(arm32)            | arm64                 | eabi                  | 
| ---- | ---                     | ---                   | ---                   |
| gcc  | gcc-arm-linux-gnueabihf | gcc-aarch64-linux-gnu | gcc-arm-linux-gnueabi | 
| g++  | g++-arm-linux-gnueabihf | g++-aarch64-linux-gnu | g++-arm-linux-gnueabi |

在进行交叉汇编之前需要安装各个平台的工具.


交叉汇编 linux 系统 arm64 架构的目标文件:

```makefile
arm64:
    GOOS=linux \
    GOARCH=arm64 \
    GOARM=7 \
    CGO_ENABLED=1 \
    CC=aarch64-linux-gnu-gcc \
    CXX=aarch64-linux-gnu-g++ \
    AR=aarch64-linux-gnu-ar \
    go build -o xxx *.go
```


交叉汇编 linux 系统 armhf 架构的目标文件:

```makefile
arm64:
    GOOS=linux \
    GOARCH=arm64 \
    GOARM=7 \
    CGO_ENABLED=1 \
    CC=arm-linux-gnueabihf-gcc \
    CXX=arm-linux-gnueabihf-g++ \
    AR=arm-linux-gnueabihf-ar \
    go build -o xxx *.go
```

> arm 交叉汇编下载地址: http://releases.linaro.org/components/toolchain/binaries, 选择 `aarch64-linux-gnu`
> `arm-linux-gnueabi`,`arm-linux-gnueabihf`目录下的文件作为交叉编译工具.

至于针对其他的汇编平台, 可以类比, 但是注意选择交叉编译工具.


