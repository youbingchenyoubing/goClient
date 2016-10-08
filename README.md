> 这个程序主要是改编于[Vegeta](https://github.com/tsenart/vegeta),可以自定义协议，进行模拟多个客户端进行攻击服务器。

----------

源码介绍：
=====
lib文件中主要包含了`command.go,attack.go,io.go,parseProtocol.go,csvoperation.go`

`command.go`:主要是用来封装协议，大家可以根据自己的需求来封装要发给服务器的包 
`attack.go`: 主要是用来创建TCP连接（通过协程模拟多个客户端）
`io.go`: 是自己的协议需求的读取文件，所以将代码单独开来。
`parseProtocol.go`：解析服务器发送过来的协议，当然如果是纯粹的攻击，可以不用解析。
`csvoperation.go`： 记录攻击过程中，服务器回应的时间，以及一些状态，方便后期数据分析

> 除了lib文件中的3个go源代码，还有`main.go,attack.go`两个代码，主要是用来解析命令参数和接入lib中的接口


使用方法：
======
```shell
./goClient attack -targets=172.16.1.248 -port=23456 -file=all.txt -offset=45555 -length=40960 -protocol=read -function=false -workers=1000
# targets 攻击目标IP
# port   攻击目标端口
# file   以文件内容攻击，适合批量攻击
# offset 读取文件的偏移量 （如果function为false）这个偏移量没实际意义，由程序随机生成
# length 读取文件长度
# -protocol 攻击协议（目前只有read/write）大家可以根据自己的攻击对象来自定义协议
# workers 协程的个数。每个协程模拟出客户端

```

----------

Author：chenyoubing
Contact: chenyoubing@stu.xmu.edu.cn
