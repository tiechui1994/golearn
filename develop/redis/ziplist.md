## ziplist

- ziplist 的数据结构:

```
<zlbytes><zltail><zllen><entry1>...<entryN><zlend>
```

zlbytes 4字节, 整个 ziplist 占用的内存字节数.

zltail 4字节, 到达 ziplist 表尾节点的偏移量. 通过这个偏移量, 可以直接找到表尾节点

zllen  2字节, ziplist 当中 entry 的数量. 当这个值小于 2^16-1 时, 这个值就是 ziplist 中节点的数量; 当值是 2^16-1
时, 节点的数量需要遍历才能获得.

entryX 可变长度, ziplist所保存的节点. 

zlend, 1字节, 值是 0xFF, 用于标记 ziplist 的末端.


- entry 的数据结构:

```
<prerawlen><len><data>
```

prerawlen 表示前一个节点的长度, 通过这个值, 可以进行指针计算, 从而跳转到上一个节点. 从而达到倒序遍历的目的.

prerawlen 占用1个字节或5个字节:

1字节: 如果前一个节点长度小于 254 字节, 使用一个字节保存它的值.

5字节: 如果前一个节点长度大于等于 254 字节, 那么第一个字节设置为 254, 后面的4个字节保存实际的长度. 

> 注: 由于 255 是作为 zlend 的特定值, 所以这里使用 254 作为分界线.

len 当中包含了编码类型和数据的长度, 使用变长编码, 分为9种情况:

00xxxxxx, 1个字节, 存储字符数组, 数组的长度小于等于 2^6-1 字节
01xxxxxx xxxxxxxx, 2字节, 存储字符数组, 数组长度小于等于 2^14-1 字节
10000000 xxxxxxxx xxxxxxxx xxxxxxxx xxxxxxxx 5字节, 存储字符数组, 数组长度小于等于 2^32-1 字节

11000000 1 字节, int16类型整数
11010000 1 字节, int32类型整数
11100000 1 字节, int64类型整数
11110000 1 字节, 24 bit有符号整数
11111110 1 字节, 8 bit有符号整数
1111xxxx 1 字节, 4 bit有符号整数 (介于0-12之间)


data: 真实的数据.

例如: `hello` 案例

```
prerawlen: ?
len: 1字节 (00000101)
data: 5字节, 内容是 hello 
```


- 将新节点插入到 ziplist 的末端过程:

1). 记录到达 ziplist 末端所需的偏移量

2). 根据新保存的值, 计算编码方式和长度(len), 以及编码前一个节点的长度所需的空间大小(prerawlen), 然后进行内存重新
分配.

3) 设置新节点的各项属性 prerawlen, len, data

4) 更新 ziplist 的相关属性, zlbytes, zltail, zllen

```

```

