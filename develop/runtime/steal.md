## Go 调度盗取算法


偷取算法的核心的原理:

```cgo
如果 p 和 q 互为质数(p, q的最大公因数是1), 并且 p < q 

那么, 从任何一个随机位置 offset 开始, 每次 offset 位置, 并设置新的 offset=(offset+p)%q, 经过 q 次一定能够访问完
q 当中所有的元素.  
```


使用 Go 表示:

```cgo
offset := uint32(random()) % nprocs
coprime := 选取一个数 coprime, 满足 coprime 与 nprocs 互为质数

for i := 0; i < nprocs; i++ {
    p := allp[offset]
    从p的运行队列偷取goroutine
    if 偷取成功 {
        break
    }
    
    offset = (offset + coprime) % nprocs
}
```

offset 表示每次开始偷取选取的随机位置.

coprime 表示随机选取与 nprocs 互为质数的数.


现假设nprocs为8, 也就是一共有8个p. 

如果第一次随机选择的 offset=6, coprime=5(5与8互质, 满足算法要求), 则从 allp 切片中偷取的下标顺序为 `6,3,0,5,2,7,4,1`
计算过程:

```
6, (6+5)%8=3, (3+5)%8=0, (0+5)%8=5, (5+5)%8=2, (2+5)%8=7, (7+5)%8=4, (4+5)%8=1
```


两个数的最大公约数: 辗转相除法

```cgo
求解 a, b 的最大公约数:

for b != 0 {
    a, b = b, a%b 
}

return a
```

如果两个数的最大公约数是 1, 则这两个数互为质数.