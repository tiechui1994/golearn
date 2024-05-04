## Go Mobile

GoMobile 是将 Go 代码库转换成 Android/iOS 库的一种方式.

### 将 Go 编译 Android Jar 包

准备工作:

1. 下载 android-ndk. 网址: https://developer.android.com/ndk/downloads

```
mkdir -p ${HOME}/android
curl https://dl.google.com/android/repository/android-ndk-r24-linux.zip -o android-ndk-r24.zip
unzip android-ndk-r24.zip && mv android-ndk-r24 ${HOME}/android
```

2. 下载 android-sdk. (android-sdk 是通过 commandlinetools 工具间接下载的).

网址: https://developer.android.com/studio, 选择 `Command line tools` 当中的下载项

```
mkdir -p  ${HOME}/android/android-sdk

# Download Tools
curl https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip -o commandlinetools.zip
unzip commandlinetools.zip
mv cmdline-tools ${HOME}/android/android-sdk

# Download Android SDK
${HOME}/android/android-sdk/cmdline-tools/bin/sdkmanager "platform-tools" "platforms;android-23" --sdk_root=${HOME}/android/android-sdk/cmdline-tools
```

> 注: android-23 当中的 23 是 API 级别. 这个对应的是 Android6.0, 对于 Android10, 需要 android-29. 自行决定使用哪个API级别的
Android 版本

3. 安装 gomobile, gobind, 在 go 的版本升级到 go1.16 以上后, 执行命令:

```cgo
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
```

4. 清理本地 go-build 缓存, 目录是 `~/.cache/go-build`

5. 编译 makefile

// 编译 demo/makefile
```makefile
# config
export ANDROID_HOME=${HOME}/android/android-sdk/cmdline-tools
export ANDROID_NDK_HOME=${HOME}/android/android-ndk-r24
export TOOL=${HOME}/android/android-ndk-r24

android: depend
    gomobile bind -target=android/arm64 -androidapi=23 -o device.aar -v -x ${HOME}/demo

ios: depend
    gomobile bind -target=ios -o device.framework -v ${HOME}/demo

depend:
    cd ${HOME}/demo
    gomobile init
    go get golang.org/x/mobile/bind

clean:
    rm -rvf libdevice.*
```

> `androidapi` 是 API 版本, target 是CPU架构


// 源代码 demo/demo.go
```go
// demo.go
package demo

import (
    "fmt"
    "io/ioutil"
    "path/filepath"
    "time"
)

var done = make(chan struct{})

func Start(dir string) {
    fmt.Println("dir", dir)

    file := filepath.Join(dir, "test.log")
    err := ioutil.WriteFile(file, []byte(time.Now().String()), 0666)
    if err != nil {
        fmt.Printf("Writefile:%v\n", err)
    }
    select {
    case <-done:
    case <-time.After(10 * time.Minute):
    }
}

func Stop(mac string) {
    fmt.Println("stop mac", mac)
    close(done)
}
```

### 将 Go 编译 Android 可执行程序

准备工作:

1. 下载 android-ndk. 网址: https://developer.android.com/ndk/downloads

```
mkdir -p ${HOME}/android
curl https://dl.google.com/android/repository/android-ndk-r24-linux.zip -o android-ndk-r24.zip
unzip android-ndk-r24.zip && mv android-ndk-r24 ${HOME}/android
```

2. 下载 android-sdk

```
export ARCH=arm
export NDK_ROOT=${HOME}/android/ndk-toolchain/${ARCH}
python ${HOME}/android/android-ndk-r24/build/tools/make_standalone_toolchain.py --arch $ARCH --api 22 --install-dir $NDK_ROOT
```

> arch 指定目标编译架构 {arm,arm64,x86,x86_64}, api, 指定 Android API版本

3. 编译

```
export CC=${HOME}/android/ndk-toolchain/arm/bin/arm-linux-androideabi-gcc
export GOOS=android
export GOARCH=arm
export GOARM=7
export CGO_ENABLED=1

go build -x main.go
```

> 这里的 main.go 只是一个简单的 Go 程序.
