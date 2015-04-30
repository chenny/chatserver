###Golang实现的高性能聊天服务器
fork : https://github.com/gansidui/chatserver

###特点

完全异步处理

protobuf序列化，AES加密数据

redis存储所有数据

支持讨论组

支持离线消息过期


###数据包构造

（1）字节序：大端模式

（2）数据包组成：包长 + 类型 + 包体

 包长：4字节，uint32，整个数据包的长度

 类型：4字节，uint32

 包体：字节数组，[]byte

 包长和类型用明文传输，包体由结构体采用protobuf序列化后再进行AES加密得到。


###通信协议

####登陆

客户端登陆包：PK_ClientLogin

服务器回复是否登录成功：PK_ServerAcceptLogin

客户端在一定时间内没有登录成功，服务器会主动断开连接。

（1）客户端在连接服务器后需要在一定时间内发送登陆包。

（2）若客户端与服务器已经存在一条连接，则断开存在的连接，开始新的这个连接，
也就是只能登陆一个客户端。


####下线

客户端主动下线数据包：PK_ClientLogout

（1）主动下线，客户端向服务器发送PK_ClientLogout包，服务器在收到PK_ClientLogout包后断开连接。

（2）被动下线，一般是在网络掉线的情况下，这种情况依赖于服务器的超时机制。


####心跳

客户端心跳包：PK_ClientPing

客户端需要定时向服务器发送PK_ClientPing包，维持在线状态。

服务器设置的是读超时，所以在收到PK_ClientPing包会选择忽略。


####C2C、讨论组聊天以及离线消息请求

详见：https://github.com/chenny/chatserver/blob/master/pb/pb.proto


###使用方法
首先根据config.ini配置并运行多个redis-server实例

再运行服务器程序 go run main.go

client_test文件夹下为多个测试程序


###TODO
（1）过载保护：例如conn数量控制，针对单个conn进行流量控制......

（2）讨论组消息缓存优化，目前的策略是将收到的讨论组消息立即挨个发送给各个组成员，每次都要从redis获取一遍讨论组成员列表，......

（3）信息监控

......