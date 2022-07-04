## gomobile

gomobile 是将 Go 代码库转换成 Android/iOS 库的一种方式.


### 将 Go 编译 Android Jar 包

准备工作:

1. 下载 android-ndk. 网址: https://dl.google.com/android/repository/android-ndk-r23b-linux.zip

```
mkdir -p ~/Documents/android
curl https://dl.google.com/android/repository/android-ndk-r22b-linux.zip -o android-ndk-r23b.zip
unzip android-ndk-r22b.zip && mv android-ndk-r22b ~/Documents/android
```

> 注: 目前虽然已经有了 android-ndk-r23b, android-ndk-r24, 但是编译的时候存在点问题. 需要自己添加些文件连接. 建议
使用 android-ndk-r22b 

2. 下载 android-sdk. (android-sdk 是通过 commandlinetools 工具间接下载的).

网址: https://dl.google.com/android/repository/commandlinetools-linux-8092744_latest.zip

```
# download
curl https://dl.google.com/android/repository/commandlinetools-linux-8092744_latest.zip -o commandlinetools.zip
unzip commandlinetools.zip
mv commandlinetools-linux-7583922_latest/cmdline-tools . && rm -rf commandlinetools-linux-7583922_latest

# sdk
./cmdline-tools/bin/sdkmanager "platform-tools" "platforms;android-23" --sdk_root=./cmdline-tools
mv mv cmdline-tools ~/Documents/android/android-sdk
```

> 注: android-23 当中的 23 是 API 级别. 这个对应的是 Android6.0, 对于 Android10, 需要 android-29. 自行决定使
用哪个API级别的Android版本

3. 安装 gomobile, gobind, 在 go 的版本升级到 go1.16 以上后, 执行命令:

```cgo
GO111MODULE=off  go install golang.org/x/mobile/cmd/gomobile@latest
GO111MODULE=off  go install golang.org/x/mobile/cmd/gobind@latest
```

4. 清理本地 go-build 缓存, 目录是 `~/.cache/go-build`

5. 编译 makefile

// makefile
```makefile
SRC = $(wildcard *.go)

# config
export ANDROID_HOME=~/Documents/android/android-sdk
export ANDROID_NDK_HOME=~/Documents/android/android-ndk-r22b
export TOOL=~/Documents/android/android-ndk-r22b

android:
	GO111MODULE=off \
	gomobile bind -target=android/arm64 -o device.aar -v -x /home/user/go/src/cloud/vdevice/demo

ios:
	GO111MODULE=off \
	gomobile bind -target=ios -o device.framework -v /home/user/go/src/cloud/vdevice/demo

clean:
	rm -rvf libdevice.*
```

// demo.go
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