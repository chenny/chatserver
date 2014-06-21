package main

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gansidui/chatserver/pb"
	"github.com/gansidui/chatserver/utils/aes"
	"net"
	"os"
	"time"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	fmt.Printf("listen %v\n", listener.Addr().String())

	for {
		conn, err := listener.AcceptTCP()
		if err == nil {

			fmt.Printf("accept %v\n", conn.RemoteAddr().String())

			// read
			buf := make([]byte, 1024)
			readSize, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("read error: %v\n", err)
			}
			fmt.Printf("read size == %v\n", readSize)

			// 先aes解密，再protobuf反序列化
			decrypted, _ := aes.AesDecrypt(buf[:readSize])
			clientLogin := &pb.PbClientLogin{}
			err = proto.Unmarshal(decrypted, clientLogin)
			if err != nil {
				fmt.Printf("unmarshal error: %v\n", err)
			}

			fmt.Println(clientLogin.GetUuid())
			fmt.Println(TimestampToTimeString(clientLogin.GetTimestamp()))
			fmt.Println(clientLogin.GetVersion())

			///////////////////////////////////////////////////////////////
			// 序列化
			s := "hi 你好, hello, 我是服务器，who are you ?，你是客户端吗？"
			reply := &pb.PbClientLogin{
				Uuid:      proto.String(s),
				Version:   proto.Float32(9.18),
				Timestamp: proto.Int64(time.Now().Unix()),
			}
			writeData, err := proto.Marshal(reply)
			if err != nil {
				fmt.Printf("marshal error: %v\n", err)
			}

			// aes加密后，write
			encrypted, _ := aes.AesEncrypt(writeData)
			writeSize, err := conn.Write(encrypted)
			if err != nil {
				fmt.Printf("write error: %v\n", err)
			}
			fmt.Printf("send size == %v\n", writeSize)
			fmt.Println(s)
		}
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Printf("check error: %v\n", err)
		os.Exit(1)
	}
}

func TimestampToTimeString(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
