package main

import (
	proto "code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/handlers"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/pb"
	"github.com/gansidui/chatserver/utils/convert"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	i_uuid string
	u_uuid string
)

// 发送心跳包
func ping(conn *net.TCPConn) {
	ticker := time.NewTicker(10 * time.Second)
	for _ = range ticker.C {
		//write
		writePingMsg := &pb.PbClientPing{
			Ping:      proto.Bool(true),
			Timestamp: proto.Int64(time.Now().Unix()),
		}
		err := handlers.SendPbData(conn, packet.PK_ClientPing, writePingMsg)
		if err != nil {
			return
		}
		fmt.Println(conn.RemoteAddr().String(), "ping.")
	}
}

func testBB(i_uuid string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", config.Addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("%v DialTCP失败: %v\r\n", i_uuid, err)
		return
	}
	defer conn.Close()

	// 登陆
	// write
	writeLoginMsg := &pb.PbClientLogin{
		Uuid:      proto.String(i_uuid),
		Version:   proto.Float32(3.14),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	err = handlers.SendPbData(conn, packet.PK_ClientLogin, writeLoginMsg)
	if err != nil {
		log.Printf("%v 发送登陆包失败: %v\r\n", i_uuid, err)
		return
	}

	var (
		bLen   []byte = make([]byte, 4)
		bType  []byte = make([]byte, 4)
		pacLen uint32
	)

	// read
	if n, err := io.ReadFull(conn, bLen); err != nil && n != 4 {
		log.Printf("Read pacLen failed: %v\r\n", err)
		return
	}
	if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
		log.Printf("Read pacType failed: %v\r\n", err)
		return
	}
	if pacLen = convert.BytesToUint32(bLen); pacLen > uint32(2048) {
		log.Printf("pacLen larger than maxPacLen\r\n")
		return
	}
	pacData := make([]byte, pacLen-8)
	if n, err := io.ReadFull(conn, pacData); err != nil && n != int(pacLen) {
		log.Printf("Read pacData failed: %v\r\n", err)
		return
	}
	pac := &packet.Packet{
		Len:  pacLen,
		Type: convert.BytesToUint32(bType),
		Data: pacData,
	}

	readAccepLoginMsg := &pb.PbServerAcceptLogin{}
	err = packet.Unpack(pac, readAccepLoginMsg)
	if err != nil {
		log.Printf("%v Unpack error: %v\r\n", i_uuid, err)
		return
	}
	fmt.Println(readAccepLoginMsg.GetLogin())
	fmt.Println(readAccepLoginMsg.GetTipsMsg())
	fmt.Println(convert.TimestampToTimeString(readAccepLoginMsg.GetTimestamp()))

	// 定时发送心跳包
	go ping(conn)

	// 先向对方发送消息
	// write
	go func() {
		writeC2CMsg := &pb.PbC2CTextChat{
			FromUuid:  proto.String(i_uuid),
			ToUuid:    proto.String(u_uuid),
			TextMsg:   proto.String("hi, 我的uuid是： " + i_uuid),
			Timestamp: proto.Int64(time.Now().Unix()),
		}
		err := handlers.SendPbData(conn, packet.PK_C2CTextChat, writeC2CMsg)
		if err != nil {
			log.Printf("%v 发送消息失败: %v\r\n", i_uuid, err)
			return
		}
	}()

	// 死循环，接收消息和发送消息
	for {
		fmt.Println("坐等消息到来...")
		// read
		if n, err := io.ReadFull(conn, bLen); err != nil && n != 4 {
			log.Printf("Read pacLen failed: %v\r\n", err)
			return
		}
		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			log.Printf("Read pacType failed: %v\r\n", err)
			return
		}
		if pacLen = convert.BytesToUint32(bLen); pacLen > uint32(2048) {
			log.Printf("pacLen larger than maxPacLen\r\n")
			return
		}
		pacData := make([]byte, pacLen-8)
		if n, err := io.ReadFull(conn, pacData); err != nil && n != int(pacLen) {
			log.Printf("Read pacData failed: %v\r\n", err)
			return
		}
		pac := &packet.Packet{
			Len:  pacLen,
			Type: convert.BytesToUint32(bType),
			Data: pacData,
		}
		readC2CMsg := &pb.PbC2CTextChat{}
		err = packet.Unpack(pac, readC2CMsg)
		if err != nil {
			log.Printf("%v 读取到的消息Unpack error: %v\r\n", i_uuid, err)
			return
		}

		from_uuid := readC2CMsg.GetFromUuid()
		to_uuid := readC2CMsg.GetToUuid()
		txt_msg := readC2CMsg.GetTextMsg()
		timestamp := readC2CMsg.GetTimestamp()

		fmt.Println("from_uuid:", from_uuid)
		fmt.Println("to_uuid:", to_uuid)
		fmt.Println("txt_msg:", txt_msg)
		fmt.Println("timestamp:", convert.TimestampToTimeString(timestamp))

		time.Sleep(5 * time.Second)

		// write
		writeC2CMsg := &pb.PbC2CTextChat{
			FromUuid:  proto.String(to_uuid),
			ToUuid:    proto.String(from_uuid),
			TextMsg:   proto.String(txt_msg + "我是 " + i_uuid),
			Timestamp: proto.Int64(timestamp),
		}
		err = handlers.SendPbData(conn, packet.PK_C2CTextChat, writeC2CMsg)
		if err != nil {
			log.Printf("%v 回复消息失败: %v\r\n", i_uuid, err)
			return
		}
	}
}

func main() {
	// for i := 0; i < 1; i++ {
	// 	time.Sleep(50 * time.Millisecond)
	// 	go testBB(i)
	// }
	fmt.Println("my name is ", i_uuid)
	fmt.Println("your name is ", u_uuid)
	go testBB(i_uuid)
	fmt.Println("sleep...")
	time.Sleep(360000 * time.Second)
}

func init() {
	flag.StringVar(&i_uuid, "i", "1", "自己的uuid")
	flag.StringVar(&u_uuid, "u", "2", "对方的uuid")
	flag.Parse()
	// 读取配置文件
	err := config.ReadIniFile("../config.ini")
	if err != nil {
		log.Fatal(err, "\r\n")
	}
	// setLogOutput("./log.txt")
}

func setLogOutput(filepath string) {
	// 为log添加短文件名，方便查看行数
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	logfile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	// 注意不要关闭logfile
	if err != nil {
		log.Printf("%v\r\n", err)
	}
	log.SetOutput(logfile)
}
