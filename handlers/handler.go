package handlers

import (
	"fmt"
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/dao/cid"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/pb"
	"github.com/gansidui/chatserver/report"
	"github.com/gansidui/chatserver/utils/safemap"
	proto "github.com/golang/protobuf/proto"
	"net"
	"time"
)

var (
	// 映射conn到uuid, uuid --> conn
	UuidMapConn *safemap.SafeMap = safemap.NewSafeMap()
	// 映射uuid到conn, conn --> uuid
	ConnMapUuid *safemap.SafeMap = safemap.NewSafeMap()
	// ConnMapLoginStatus 映射conn的登陆状态,conn->loginstatus
	// loginstatus为nil表示conn已经登陆, 否则 loginstatus表示conn连接服务器时的时间
	// 用于判断登陆是否超时，控制恶意连接
	ConnMapLoginStatus *safemap.SafeMap = safemap.NewSafeMap()
)

// 关闭conn
func CloseConn(conn *net.TCPConn) {
	conn.Close()
	ConnMapLoginStatus.Delete(conn)
	uuid := ConnMapUuid.Get(conn)
	UuidMapConn.Delete(uuid)
	ConnMapUuid.Delete(conn)
	if uuid != nil {
		cid.UuidOffLine(uuid.(string))
	}
	report.AddCount(report.OnlineUser, -1)
}

// 初始化conn
func InitConn(conn *net.TCPConn, uuid string) {
	ConnMapLoginStatus.Set(conn, nil)
	UuidMapConn.Set(uuid, conn)
	ConnMapUuid.Set(conn, uuid)
	cid.UuidOnLine(uuid)
	report.AddCount(report.OnlineUser, 1)
}

// 发送字节流
func SendByteStream(conn *net.TCPConn, buf []byte) error {
	conn.SetWriteDeadline(time.Now().Add(time.Duration(config.WriteTimeout) * time.Second))
	n, err := conn.Write(buf)
	if n != len(buf) || err != nil {
		return fmt.Errorf("Write to %v failed, Error: %v", ConnMapUuid.Get(conn).(string), err)
	}
	return nil
}

// 发送protobuf结构数据
func SendPbData(conn *net.TCPConn, dataType uint32, pb interface{}) error {
	pac, err := packet.Pack(dataType, pb)
	if err != nil {
		return err
	}
	return SendByteStream(conn, pac.GetBytes())
}

// 客户端登陆
func HandleClientLogin(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientLogin{}
	packet.Unpack(recPacket, readMsg)

	uuid := readMsg.GetUuid()

	// 检测uuid合法性
	if cid.UuidCheckExist(uuid) {
		// 如果已经在线则关闭以前的conn
		if cid.UuidCheckOnline(uuid) {
			co := UuidMapConn.Get(uuid)
			if co != nil {
				CloseConn(co.(*net.TCPConn))
			}
		}
		// 上线conn
		InitConn(conn, uuid)
	} else {
		CloseConn(conn)
	}

	fmt.Println("recPacket=:", recPacket)
	fmt.Println("readMsg=:", readMsg)
	fmt.Println("uuid:", readMsg.GetUuid())
	// fmt.Println("version:", readMsg.GetVersion())
	// fmt.Println("timestamp:", convert.TimestampToTimeString(readMsg.GetTimestamp()))

	// write
	writeMsg := &pb.PbServerAcceptLogin{
		Login:     proto.Bool(true),
		TipsMsg:   proto.String("登陆成功"),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	SendPbData(conn, packet.PK_ServerAcceptLogin, writeMsg)
}

// 客户端下线
func HandleClientLogout(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientLogout{}
	packet.Unpack(recPacket, readMsg)

	// fmt.Println("logout:", readMsg.GetLogout())
	// fmt.Println("timestamp:", convert.TimestampToTimeString(readMsg.GetTimestamp()))
	CloseConn(conn)
}

// 心跳包
func HandleClientPing(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	// readMsg := &pb.PbClientPing{}
	// packet.Unpack(recPacket, readMsg)

	// fmt.Println("ping:", readMsg.GetPing())
	// fmt.Println("timestamp:", convert.TimestampToTimeString(readMsg.GetTimestamp()))
}
