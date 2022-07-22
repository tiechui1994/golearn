# WebSocket 协议

## 数据帧

在WebSocket协议中，数据使用帧序列来传输. 

客户端必须掩码(mask)发送到服务器的所有帧`(注意: 不管WebSocket协议是否运行在TLS至上，掩码都要做)`. 当收到
一个没有掩码的帧时, 服务器必须关闭连接. 这种状况下, 服务器可能发送一个1002(协议错误)的Close帧.

服务器必须不掩码发送到客户端的所有帧. 如果客户端检测到掩码的帧, 它必须关闭连接. 这种状况下, 客户端可能发送一个
1002(协议错误)的Close帧.


基本帧协议定义了带有操作码(opcode)的帧类型, 负载长度

数据帧可以被客户端或者服务器在打开阶段握手完成之后和端点发送Close帧之前的任何时候传输.

![image](/images/http_websocket.png)

FIN: 1bit, 指示这个是消息的最后片段。第一个片段可能也是最后的片段.

RSV1, RSV2, RSV3: 3bit, 必须是0, 除非一个扩展协商为非零值定义含义. 如果收到一个非零值且没有协商的扩展定义
这个非零值的含义, 接收端点必须失败WebSokcket连接.

opcode: 4bit, 定义了"负载数据"的解释. 如果是一个未知的操作码, 接收端必须失败WebSocket连接.
```
0x0 代表一个继续帧
0x1 代表一个文本帧
0x2 代表一个二进制帧
0x3-7 保留用于未来的非控制帧

0x8 代表连接关闭
0x9 代表ping
0xA 代表pong
0xB-F 保留用于未来的控制帧
```

MASK: 1bit, 定义是否"负载数据"是掩码的. 如果设置是1, 表示掩码(mask)

payload len: 7bit, 7+16bit, 或者 7+64bit. 负载数据的长度, 以字节为单位. 
```
如果是0-125, 这是负载长度.

如果是126, 之后的2字节解释为一个16位的无符号整数的负载长度.

如果是127, 之后的8字节解释位一个64位的无符号整数的负载长度.

注: 多字节长度数量以网络字节顺序表示. 在所有情况下, 最小数量的字节必须用于编码长度. 例如, 一个124字节长的字符串
不能被编码位序列126,0,124. 

负载长度 = "扩展数据"长度 + "应用数据"长度.  "扩展数据"长度可能是0, 这种状况下, 负载长度是"应用数据"长度.
说明: 扩展数据, 指的是扩展协议, websocket扩展项部分. 应用数据, 指的是网络传输的数据
```

masking-key: 0|32bit, 客户端发送到服务器的所有帧通过一个包含在帧中的32位值来掩码. 如果mask位设置为1, 则
该字段存在; 如果mask位设置为0, 则该字段缺失.


payload data: (x+y)*8 bit, "负载数据"="扩展数据"+"应用数据". 其中, 扩展数据x字节, 应用数据y字节
```
extension data: x字节, 如果没有协商使用扩展的话, 扩展数据为0字节. 所有的扩展都必须声明扩展数据的长度,
或者可以计算出扩展数据的长度. 此外, 扩展如何使用必须在握手阶段就协商好. 如果扩展数据存在, 那么负载数据的
长度必须将将扩展数据的长度包含在内.


application data: y字节, 任意的"应用数据", 在扩展数据之后(如果存在扩展数据), 占据了数据帧剩余的位置. 
"应用数据"的长度= 负载长度 - "扩展数据"长度
```


格式:
```
ws-frame                = frame-fin            ; 1位长度
                          frame-rsv1           ; 1位长度
                          frame-rsv2           ; 1位长度
                          frame-rsv3           ; 1位长度
                          frame-opcode         ; 4位长度
                          frame-masked         ; 1位长度
                          frame-payload-length ; 7、 7+16、; 或者7+64 位长度
                          [frame-masking-key]  ; 32位长度
                          frame-payload-data   ; n*8位长度; n>=0


frame-fin               = 0x0 ; 这条消息后续还有更多的帧
                          0x1 ; 这条消息的最终帧
                              ; 1位长度


frame-rsv1              = 0x0 | 0x1
                              ; 1位长度，必须是0，除非协商其他

frame-rsv2              = 0x0 | 0x1
                              ; 1位长度，必须是0，除非协商其他

frame-rsv3              = 0x0 | 0x1
                              ; 1位长度，必须是0，除非协商其他


frame-opcode            = 0x0   ; 帧继续
                          0x1   ; 文本帧
                          0x2   ; 二进制帧
                          0x3-7 ; 保留
                          0x8   ; 连接关闭
                          0x9   ; ping
                          0xA   ; pong
                          0B-F  ; 保留用于未来的控制帧 
                                ; 4位长度


frame-masked            = 0x0   ; 帧没有掩码，没有frame-masking-key
                          0x1   ; 帧被掩码， 存在frame-masking-key 
                                ; 1位长度


frame-payload-length    = ( 0x00-7D ) ; 7 位长度(0-0x7D)
                          ( 0x0000-FFFF ) ; 16位长度 (0x7F-0xFFFF)
                          ( 0x0000000000000000-7FFFFFFFFFFFFFFF ) 64位长度 (0x10000-0x7FFFFFFFFFFFFFFF)


frame-masking-key       = 0x20( 0x00-FF ) ; 仅当frame-masked 是1时存在
                                ; 32位长度



frame-payload-data      = (frame-masked-extension-data
                           frame-masked-application-data)
                                ; 当frame-masked是1
                        
                          (frame-unmasked-extension-data
                           frame-unmasked-application-data)
                                ; 当frame-masked是0
                                
frame-masked-extension-data     = *( 0x00-FF ) ; 保留用于未来扩展
                                   ; n*8 位长度，n >= 0

frame-masked-application-data   = *( 0x00-FF ) ; 数据
                                   ; n*8 位长度，n >= 0

frame-unmasked-extension-data   = *( 0x00-FF ) ; 保留用于未来扩展
                                   ; n*8 位长度，n >= 0

frame-unmasked-application-data = *( 0x00-FF ) ; 数据
                                   ; n*8 位长度，n >= 0
```


## 掩码算法

masking-key是客户端挑出来的32位的随机数. 掩码操作不会影响数据负载长度, 掩码,反掩码操作都采用如下算法:

首先, 假设
- origin-octet-i: 为原始数据的第i字节
-transformed-octet-i: 为转换后的数据的第i字节
- j: mod 4 的结果
- masking-key-octet-i: masking-key的第i字节

算法:

j = i MOD 4

transformed-octet-i = original-octet-i XOR masking-key-octet-j