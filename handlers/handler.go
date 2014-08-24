package handlers

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/dao/c2c"
	"github.com/gansidui/chatserver/dao/cid"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/pb"
	"github.com/gansidui/chatserver/report"
	"github.com/gansidui/chatserver/utils/convert"
	"github.com/gansidui/chatserver/utils/safemap"
	"log"
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

// 处理登陆
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

	// fmt.Println("uuid:", readMsg.GetUuid())
	// fmt.Println("version:", readMsg.GetVersion())
	// fmt.Println("timestamp:", convert.TimestampToTimeString(readMsg.GetTimestamp()))

	// write
	writeMsg := &pb.PbServerAcceptLogin{
		Login:     proto.Bool(true),
		TipsMsg:   proto.String("登陆成功"),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	SendPbData(conn, packet.PK_ServerAcceptLogin, writeMsg)

	// 检查是否有该uuid的离线消息存在，若有，则发送其离线消息
	if offmsgNum := c2c.GetMsgNum(uuid); offmsgNum > 0 {
		// 这里比较复杂，后续再优化(可以多个离线消息一起发送)
		// 得到所有离线消息
		// fmt.Println("转发离线消息")
		msgs := c2c.GetMsgs(uuid, offmsgNum)
		c2c.DeleteMsgs(uuid, offmsgNum, offmsgNum)
		for i, _ := range msgs {
			SendByteStream(conn, []byte(msgs[i]))
		}
	}
}

// 处理下线
func HandleClientLogout(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientLogout{}
	packet.Unpack(recPacket, readMsg)

	fmt.Println("logout:", readMsg.GetLogout())
	fmt.Println("timestamp:", convert.TimestampToTimeString(readMsg.GetTimestamp()))

	CloseConn(conn)
}

// 处理心跳
func HandleClientPing(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	// readMsg := &pb.PbClientPing{}
	// packet.Unpack(recPacket, readMsg)

	// fmt.Println("ping:", readMsg.GetPing())
	// fmt.Println("timestamp:", convert.TimestampToTimeString(readMsg.GetTimestamp()))
}

// 处理客户端之间的消息转发
func HandleC2CTextChat(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbC2CTextChat{}
	packet.Unpack(recPacket, readMsg)

	from_uuid := ConnMapUuid.Get(conn).(string)
	to_uuid := readMsg.GetToUuid()
	txt_msg := readMsg.GetTextMsg()
	timestamp := readMsg.GetTimestamp()

	// 验证发送者的真实性以及发送对象是否存，若消息伪造，则断开该连接
	if readMsg.GetFromUuid() != from_uuid || !cid.UuidCheckExist(to_uuid) {
		CloseConn(conn)
		return
	}

	// write
	writeMsg := &pb.PbC2CTextChat{
		FromUuid:  proto.String(from_uuid),
		ToUuid:    proto.String(to_uuid),
		TextMsg:   proto.String(txt_msg),
		Timestamp: proto.Int64(timestamp),
	}
	pac, err := packet.Pack(packet.PK_C2CTextChat, writeMsg)
	if err != nil {
		log.Printf("%v\r\n", err)
		return
	}

	// 若 to_uuid 在线，则转发该消息，发送失败 或者 to_uuid不在线 则保存为离线消息
	if cid.UuidCheckOnline(to_uuid) {
		// fmt.Println("在线消息转发")
		to_conn := UuidMapConn.Get(to_uuid).(*net.TCPConn)
		if SendByteStream(to_conn, pac.GetBytes()) != nil {
			// fmt.Println("发送失败转离线消息保存")
			c2c.AddMsg(to_uuid, string(pac.GetBytes()))
			report.AddCount(report.OfflineMsg, 1)
		} else {
			report.AddCount(report.OnlineMsg, 1)
		}
	} else {
		// fmt.Println("不在线转离线消息保存")
		c2c.AddMsg(to_uuid, string(pac.GetBytes()))
		report.AddCount(report.OfflineMsg, 1)
	}
}
