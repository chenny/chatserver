package handlers

import (
	"github.com/gansidui/chatserver/dao/cid"
	"github.com/gansidui/chatserver/dao/group/groupinfo"
	"github.com/gansidui/chatserver/dao/group/groupmsg"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/pb"
	proto "github.com/golang/protobuf/proto"
	"log"
	"net"
	"time"
)

// 客户端申请建立讨论组
func HandleClientBuildGroup(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientBuildGroup{}
	packet.Unpack(recPacket, readMsg)
	uuid := ConnMapUuid.Get(conn).(string)
	group_name := readMsg.GetGroupName()

	// 建立讨论组
	ret, group_id := groupinfo.BuildGroup(group_name, uuid)
	tips_msg := "讨论组[" + group_name + "]建立成功"
	if ret {
		tips_msg = "讨论组[" + group_name + "]建立失败"
	}

	// write
	writeMsg := &pb.PbServerNotifyBuildGroup{
		Build:     proto.Bool(ret),
		GroupId:   proto.String(group_id),
		GroupName: proto.String(group_name),
		OwnerUuid: proto.String(uuid),
		TipsMsg:   proto.String(tips_msg),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	SendPbData(conn, packet.PK_ServerNotifyBuildGroup, writeMsg)
}

// 客户端（群主）申请解散讨论组
func HandleClientDisbandGroup(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientDisbandGroup{}
	packet.Unpack(recPacket, readMsg)
	from_id := readMsg.GetFromUuid()
	group_id := readMsg.GetGroupId()
	// timestamp := readMsg.GetTimestamp()

	// 验证from_id确实是该conn，并且是group_id的群主
	group_name, group_owner := groupinfo.GetGroupNameAndOwner(group_id)
	if from_id != ConnMapUuid.Get(conn).(string) || group_owner != from_id {
		CloseConn(conn)
		return
	}

	// 解散讨论组
	ret := groupinfo.DisbandGroup(from_id, group_id)
	tips_msg := "解散讨论组[" + group_name + "]成功"
	if ret {
		tips_msg = "解散讨论组[" + group_name + "]失败"
	}

	// write
	writeMsg := &pb.PbServerNotifyDisbandGroup{
		Disband:   proto.Bool(ret),
		GroupId:   proto.String(group_id),
		GroupName: proto.String(group_name),
		TipsMsg:   proto.String(tips_msg),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	SendPbData(conn, packet.PK_ServerNotifyDisbandGroup, writeMsg)
}

// 客户端申请加入讨论组
func HandleClientJoinGroup(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientJoinGroup{}
	packet.Unpack(recPacket, readMsg)

	from_uuid := readMsg.GetFromUuid()
	group_id := readMsg.GetGroupId()
	// note_msg := readMsg.GetNoteMsg()
	// timestamp := readMsg.GetTimestamp()

	// 加入讨论组
	if ret := groupinfo.JoinGroup(from_uuid, group_id); !ret {
		return
	}

	group_name, _ := groupinfo.GetGroupNameAndOwner(group_id)

	// write
	writeMsg := pb.PbServerNotifyJoinGroup{
		ApplicantUuid: proto.String(from_uuid),
		GroupId:       proto.String(group_id),
		GroupName:     proto.String(group_name),
		Timestamp:     proto.Int64(time.Now().Unix()),
	}

	// 通知所有组员，这个消息不离线存储
	group_members := groupinfo.GetAllUuid(group_id)
	for i := 1; i < len(group_members); i++ {
		SendPbData(UuidMapConn.Get(group_members[i]).(*net.TCPConn), packet.PK_ServerNotifyJoinGroup, writeMsg)
	}
}

// 客户端申请退出讨论组
func HandleClientLeaveGroup(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbClientLeaveGroup{}
	packet.Unpack(recPacket, readMsg)

	from_uuid := readMsg.GetFromUuid()
	group_id := readMsg.GetGroupId()
	// timestamp := readMsg.GetTimestamp()

	if ret := groupinfo.ExitGroup(from_uuid, group_id); !ret {
		return
	}

	group_name, _ := groupinfo.GetGroupNameAndOwner(group_id)

	// write
	writeMsg := &pb.PbServerNotifyLeaveGroup{
		LeaverUuid: proto.String(from_uuid),
		GroupId:    proto.String(group_id),
		GroupName:  proto.String(group_name),
		Timestamp:  proto.Int64(time.Now().Unix()),
	}

	// 通知所有组员，这个消息不离线存储
	group_members := groupinfo.GetAllUuid(group_id)
	for i := 1; i < len(group_members); i++ {
		SendPbData(UuidMapConn.Get(group_members[i]).(*net.TCPConn), packet.PK_ServerNotifyLeaveGroup, writeMsg)
	}
}

// 处理讨论组消息转发
func HandleGroupTextChat(conn *net.TCPConn, recPacket *packet.Packet) {
	// read
	readMsg := &pb.PbGroupTextChat{}
	packet.Unpack(recPacket, readMsg)

	from_uuid := ConnMapUuid.Get(conn).(string)
	group_id := readMsg.GetGroupId()
	txt_msg := readMsg.GetTextMsg()
	// timestamp := readMsg.GetTimestamp()

	// 验证发送者，这样会影响性能，视情况可以把这个验证去掉
	if readMsg.GetFromUuid() != from_uuid || !groupinfo.ExistUuidFromGroup(group_id, from_uuid) {
		CloseConn(conn)
		return
	}

	// write
	writeMsg := &pb.PbGroupTextChat{
		FromUuid:  proto.String(from_uuid),
		GroupId:   proto.String(group_id),
		TextMsg:   proto.String(txt_msg),
		Timestamp: proto.Int64(time.Now().Unix()),
	}

	// 在线消息包
	pac1, err := packet.Pack(packet.PK_GroupTextChat, writeMsg)
	if err != nil {
		log.Printf("%v\r\n", err)
		return
	}

	// 离线消息包
	pac2, err := packet.Pack(packet.PK_ServerResponseGroupOfflineMsg, writeMsg)
	if err != nil {
		log.Printf("%v\r\n", err)
		return
	}

	// 将消息转发给所有组员(除了自己)，不在线则离线存储
	group_members := groupinfo.GetAllUuid(group_id)
	for i := 1; i < len(group_members); i++ {
		if group_members[i] == from_uuid {
			continue
		}

		if cid.UuidCheckOnline(group_members[i]) {
			to_conn := UuidMapConn.Get(group_members[i]).(*net.TCPConn)
			SendByteStream(to_conn, pac1.GetBytes())
		} else {
			groupmsg.AddMsg(group_members[i], string(pac2.GetBytes()))
		}
	}

}

// 客户端请求讨论组离线消息
func HandleClientRequestGroupOfflineMsg(conn *net.TCPConn, recPacket *packet.Packet) {
	readMsg := &pb.PbClientRequestGroupOfflineMsg{}
	packet.Unpack(recPacket, readMsg)

	uuid := readMsg.GetFromUuid()
	// timestamp := readMsg.GetTimestamp()

	// 检查是否有该uuid的离线消息存在，若有，则发送其离线消息
	if offmsgNum := groupmsg.GetMsgNum(uuid); offmsgNum > 0 {
		// 这里比较复杂，后续再优化(可以多个离线消息一起发送)
		// 暂时将所有离线消息单独发送
		msgs := groupmsg.GetMsgs(uuid, offmsgNum)
		groupmsg.DeleteMsgs(uuid, offmsgNum, offmsgNum)
		for i, _ := range msgs {
			SendByteStream(conn, []byte(msgs[i]))
		}
	}
}

// 客户端请求获取讨论组信息
func HandleClientRequestGroupInfo(conn *net.TCPConn, recPacket *packet.Packet) {
	readMsg := &pb.PbClientRequestGroupInfo{}
	packet.Unpack(recPacket, readMsg)

	uuid := readMsg.GetFromUuid()
	// timestamp := readMsg.GetTimestamp()

	groupids := groupinfo.GetAllGroup(uuid)

	var allGroupInfo []*pb.PbGroupInfo

	for _, group_id := range groupids {
		s := groupinfo.GetAllUuid(group_id)
		group_name, owner_uuid, member_uuid := s[0], s[1], s[2:]

		gi := &pb.PbGroupInfo{
			GroupId:    proto.String(group_id),
			GroupName:  proto.String(group_name),
			OwnerUuid:  proto.String(owner_uuid),
			MemberUuid: member_uuid,
		}

		allGroupInfo = append(allGroupInfo, gi)
	}

	// write
	writeMsg := &pb.PbServerResponseGroupInfo{
		FromUuid:     proto.String(uuid),
		AllGroupInfo: allGroupInfo,
		Timestamp:    proto.Int64(time.Now().Unix()),
	}
	SendPbData(conn, packet.PK_ServerResponseGroupInfo, writeMsg)
}
