package handlers

import (
	proto "code.google.com/p/goprotobuf/proto"
	"github.com/gansidui/chatserver/dao/c2c/c2cmsg"
	"github.com/gansidui/chatserver/dao/cid"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/pb"
	"github.com/gansidui/chatserver/report"
	"log"
	"net"
	"time"
)

// 处理C2C消息转发
func HandleC2CTextChat(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbC2CTextChat{}
	packet.Unpack(recPacket, readMsg)

	from_uuid := ConnMapUuid.Get(conn).(string)
	to_uuid := readMsg.GetToUuid()
	txt_msg := readMsg.GetTextMsg()
	// timestamp := readMsg.GetTimestamp()

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
		Timestamp: proto.Int64(time.Now().Unix()),
	}

	// 在线消息包
	pac1, err := packet.Pack(packet.PK_C2CTextChat, writeMsg)
	if err != nil {
		log.Printf("%v\r\n", err)
		return
	}

	// 离线消息包
	pac2, err := packet.Pack(packet.PK_ServerResponseC2COfflineMsg, writeMsg)
	if err != nil {
		log.Printf("%v\r\n", err)
		return
	}

	// 若 to_uuid 在线，则转发该消息，发送失败 或者 to_uuid不在线 则保存为离线消息
	if cid.UuidCheckOnline(to_uuid) {
		to_conn := UuidMapConn.Get(to_uuid).(*net.TCPConn)
		if SendByteStream(to_conn, pac1.GetBytes()) != nil {
			c2cmsg.AddMsg(to_uuid, string(pac2.GetBytes()))
			report.AddCount(report.OfflineMsg, 1)
		} else {
			report.AddCount(report.OnlineMsg, 1)
		}
	} else {
		c2cmsg.AddMsg(to_uuid, string(pac2.GetBytes()))
		report.AddCount(report.OfflineMsg, 1)
	}
}

// 客户端请求C2C离线消息
func HandleClientRequestC2COfflineMsg(conn *net.TCPConn, recPacket *packet.Packet) {
	readMsg := &pb.PbClientRequestC2COfflineMsg{}
	packet.Unpack(recPacket, readMsg)

	uuid := readMsg.GetFromUuid()

	// 检查是否有该uuid的离线消息存在，若有，则发送其离线消息
	if offmsgNum := c2cmsg.GetMsgNum(uuid); offmsgNum > 0 {
		// 这里比较复杂，后续再优化(可以多个离线消息一起发送)
		// 暂时将所有离线消息单独发送
		msgs := c2cmsg.GetMsgs(uuid, offmsgNum)
		c2cmsg.DeleteMsgs(uuid, offmsgNum, offmsgNum)
		for i, _ := range msgs {
			SendByteStream(conn, []byte(msgs[i]))
		}
	}
}
