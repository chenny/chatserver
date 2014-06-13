package main

import (
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/dao"
	"github.com/gansidui/chatserver/handlers"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/server"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var svr *server.Server

func init() {
	// 读取配置文件
	err := config.ReadIniFile("./config.ini")
	checkError(err)

	// 设置cpu数量和日志目录
	runtime.GOMAXPROCS(config.NumCpu)
	setLogOutput(config.LogFile)

	// 初始化dao
	dao.IdMsgInit(config.IdToMsgDB)
	dao.OfflineMsgInit(config.OfflineMsgidsDB)
	dao.UuidInit(config.UuidDB)

	// 服务器初始化
	svr = server.NewServer()
	svr.SetAcceptTimeout(time.Duration(config.AcceptTimeout) * time.Second)
	svr.SetReadTimeout(time.Duration(config.ReadTimeout) * time.Second)
	svr.SetWriteTimeout(time.Duration(config.WriteTimeout) * time.Second)

	// 消息处理函数绑定
	svr.BindMsgHandler(packet.PK_ClientLogin, handlers.HandleClientLogin)
	svr.BindMsgHandler(packet.PK_ClientLogout, handlers.HandleClientLogout)
	svr.BindMsgHandler(packet.PK_ClientPing, handlers.HandleClientPing)
	svr.BindMsgHandler(packet.PK_C2CTextChat, handlers.HandleC2CTextChat)
}

func clean() {
	dao.IdMsgClean()
	dao.OfflineMsgClean()
	dao.UuidClean()
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", config.Addr)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// 服务器开始监听
	go svr.Start(listener)

	// 处理中断信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Signal: %v\r\n", <-ch)

	// 优雅的结束线程
	svr.Stop()
	clean()
}

func checkError(err error) {
	if err != nil {
		log.Printf("Error: %v\r\n", err)
		os.Exit(1)
	}
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
