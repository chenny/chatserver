package main

import (
	"flag"
	"fmt"
	"github.com/chenny/chatserver/config"
	"github.com/chenny/chatserver/handlers"
	"github.com/chenny/chatserver/packet"
	"github.com/chenny/chatserver/pb"
	"github.com/chenny/chatserver/utils/convert"
	proto "github.com/golang/protobuf/proto"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	prefix string // uuid的前缀为随机生成，方便多个客户端测试，避免将其他客户端踢下线
	total  int    // uuid数量, uuid名字为 1---total
)

func getUuid(uuid int) string {
	//return strconv.Itoa(uuid)
	return prefix + strconv.Itoa(uuid)
}

// 发送心跳包
func ping(conn *net.TCPConn) {
	ticker := time.NewTicker(60 * time.Second)
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

// 处理收发数据包
func handlePackets(uuid int, conn *net.TCPConn, receivePackets <-chan *packet.Packet, chStop <-chan bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Panic: %v\r\n")
		}
	}()
	for {
		select {
		case <-chStop:
			return

		// 消息包处理
		case p := <-receivePackets:
			if p.Type == packet.PK_ServerAcceptLogin { // 登陆回复
				// read
				readMsg := &pb.PbServerAcceptLogin{}
				packet.Unpack(p, readMsg)
				if readMsg.GetLogin() == true {
					log.Printf("[%v]: [%v]---[%v]\r\n", getUuid(uuid), readMsg.GetTipsMsg(), convert.TimestampToTimeString(readMsg.GetTimestamp()))
				}

				// write，随机向10个人发送消息
				for i := 0; i < 10; i++ {
					rand.Seed(time.Now().UnixNano())
					to_uuid := rand.Intn(total) + 1 // [1, total]
					if to_uuid == uuid {
						continue
					}

					writeMsg := &pb.PbC2CTextChat{
						FromUuid:  proto.String(getUuid(uuid)),
						ToUuid:    proto.String(getUuid(to_uuid)),
						TextMsg:   proto.String(strings.Repeat("hello,世界！！！", 2)),
						Timestamp: proto.Int64(time.Now().Unix()),
					}
					handlers.SendPbData(conn, packet.PK_C2CTextChat, writeMsg)
				}

			} else if p.Type == packet.PK_C2CTextChat { // 普通消息
				// read
				readMsg := &pb.PbC2CTextChat{}
				packet.Unpack(p, readMsg)
				from_uuid := readMsg.GetFromUuid()
				to_uuid := readMsg.GetToUuid()
				txt_msg := readMsg.GetTextMsg()
				timestamp := readMsg.GetTimestamp()
				if to_uuid != getUuid(uuid) {
					log.Printf("[%v]收到[%v]发来的不属于自己的包,该包应该属于[%v]\r\n", getUuid(uuid), from_uuid, to_uuid)
				} else {
					log.Printf("[%v]：[%v]收到来自[%v]的消息: [%v]", convert.TimestampToTimeString(timestamp), getUuid(uuid), from_uuid, txt_msg)

					// write, 回复时在原基础上加点消息，控制长度范围
					var to_txt_msg string
					var add_txt string = " 你好 hello world"

					if len(txt_msg)+len(add_txt) <= 64 {
						to_txt_msg = txt_msg + add_txt
					} else {
						to_txt_msg = txt_msg
					}

					writeMsg := &pb.PbC2CTextChat{
						FromUuid:  proto.String(getUuid(uuid)),
						ToUuid:    proto.String(from_uuid),
						TextMsg:   proto.String(to_txt_msg),
						Timestamp: proto.Int64(time.Now().Unix()),
					}
					handlers.SendPbData(conn, packet.PK_C2CTextChat, writeMsg)
				}

			} else {
				log.Printf("[%v]收到未知包\r\n", getUuid(uuid))
			}
		}
	}
}

// 模拟客户端(uuid)
func testClient(uuid int) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("uuid: [%v] Panic: %v\r\n", getUuid(uuid), e)
		}
	}()

	// 连接服务器
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", config.Addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("[%v] DialTCP失败: %v\r\n", getUuid(uuid), err)
		return
	}

	// 发送登陆请求
	writeLoginMsg := &pb.PbClientLogin{
		Uuid:      proto.String(getUuid(uuid)),
		Version:   proto.Float32(3.14),
		Timestamp: proto.Int64(time.Now().Unix()),
	}
	err = handlers.SendPbData(conn, packet.PK_ClientLogin, writeLoginMsg)
	if err != nil {
		log.Printf("[%v] 发送登陆包失败: %v\r\n", getUuid(uuid), err)
		return
	}

	// 下面这些处理和server.go中的一样
	receivePackets := make(chan *packet.Packet, 100) // 接收到的包
	chStop := make(chan bool)                        // 通知停止消息处理

	defer func() {
		conn.Close()
		chStop <- true
	}()

	// 发送心跳包
	go ping(conn)

	// 处理接受到的包
	go handlePackets(uuid, conn, receivePackets, chStop)

	// 包长(4) + 类型(4) + 包体(len(pacData))
	var (
		bLen   []byte = make([]byte, 4)
		bType  []byte = make([]byte, 4)
		pacLen uint32
	)

	for {
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

		receivePackets <- &packet.Packet{
			Len:  pacLen,
			Type: convert.BytesToUint32(bType),
			Data: pacData,
		}
	}
}

func main() {
	for i := 1; i <= 2; i++ {
		time.Sleep(50 * time.Millisecond)
		go testClient(i)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)
}

func init() {
	flag.IntVar(&total, "t", 10, "uuid的总数")
	flag.Parse()
	// 读取配置文件
	err := config.ReadIniFile("../config.ini")
	if err != nil {
		log.Fatal(err, "\r\n")
	}

	//
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(10)
	y := rand.Intn(10)
	prefix = "[" + strconv.Itoa(x) + strconv.Itoa(y) + "]--> "
}
