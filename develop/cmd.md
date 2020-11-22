# Go 使用命令的五种姿势

在 Go 当中使用命令的库是 `os/exec`, exec.Command 函数返回一个 `Cmd` 对象. 根据不同的需求, 可以将命令的执行分为三
种情况

1. 只执行命令, 不获取结果

2. 执行命令, 并获取结果(不区分 stdout 和 stderr)

3. 指向命令, 并获取结果(区分 stdout 和 stderr)


### 第一种: 只执行命令, 不获取结果

直接调用 Cmd 对象的 `Run()` 方法, 返回的只有成功和失败, 获取不到任何输出的结果.


```cgo
func main() {
    cmd := exec.Command("ls", "-l", "/var/log")
    err := cmd.Run()
    if err != nil {
        log.Fatalf("Run() failed with %v", err)
    }
}
```

### 第二种: 执行命令, 并获取结果

执行一个命令就是需要获取输出结果, 此时可以调用 Cmd 的 `CombinedOutPut()` 方法.

```cgo
func main() {
    cmd := exec.Command("ls", "-l", "/var/log")
    out, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("out: %v", out)
        log.Fatalf("CombinedOutPut() failed with %v", err)
    }
}
```

> `CombinedOutput()` 函数, 只返回 out, 并不区分 stdout 和 stderr. 
> 注意: Cmd 还有一个方法 `Output()`, 这个方法也可以获取结果, 但是是它是 stdout 的输出. 使用的时候注意两者的区别.

### 第三种: 执行命令, 并区分 stdout 和 stderr

```cgo
func main() {
    cmd := exec.Command("ls", "-l", "/var/log")
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    err := cmd.Run()
    log.Printf("out:\n%v\nerr:\n%v", stdout.String(), stderr.String())
    if err != nil {
        log.Fatalf("Run() failed with %v", err)
    }
}
```


### 第四种: 多条命令组合, 使用管道

将上一条命令的执行输出结果, 作为下一条命令的参数. 在 Shell 中可以使用管道符 `|` 实现. 

例如: 统计 message 日志当中 ERROR 日志的数量.

```bash
grep ERROR /var/log/messages | wc -l
```

在 Go 当中也是有类似的实现的.

```cgo
func main() {
    c1 := exec.Command("grep", "ERROR", "/var/log/messages")
    c2 := exec.Command("wc", "-l")
    
    c2.Stdin, _ = c1.StdoutPipe()
    c2.Stdout = os.Stdout
   
    _ = c2.Start()
    _ = c1.Run()
    _ = c2.Wait()
}
```

### 第五种: 设置命令级别的环境变量

使用 os 库的 `Setenv()` 函数来设置环境变量, 是作用于整个进程的生命周期.

```cgo
func main() {
    os.Setenv("NAME", "Go")
    
    cmd := exec.Command("echo", os.ExpandEnv("$NAME"))
    out, _ = cmd.CombinedOutput()
    if err != nil {
        log.Fatalf("CombinedOutput() failed with %v", err)
    }
    
    log.Println("out: %s", out)
}
```