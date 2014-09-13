package main

import (
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/dao/c2c/c2cmsg"
	"github.com/gansidui/chatserver/dao/group/groupinfo"
	"github.com/gansidui/chatserver/dao/group/groupmsg"
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

	// 初始化dao
	c2cmsg.Init()
	groupmsg.Init()
	groupinfo.Init()

	// 设置cpu数量和日志目录
	runtime.GOMAXPROCS(config.NumCpu)
	setLogOutput(config.LogFile)

	// 服务器初始化
	svr = server.NewServer()
	svr.SetAcceptTimeout(time.Duration(config.AcceptTimeout) * time.Second)
	svr.SetReadTimeout(time.Duration(config.ReadTimeout) * time.Second)
	svr.SetWriteTimeout(time.Duration(config.WriteTimeout) * time.Second)

	// 消息处理函数绑定
	bindMsgHandler()
}

func clean() {
	c2cmsg.Clean()
	groupmsg.Clean()
	groupinfo.Clean()
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

// 绑定 msg --> handler
func bindMsgHandler() {
	// 登陆和心跳
	svr.BindMsgHandler(packet.PK_ClientLogin, handlers.HandleClientLogin)
	svr.BindMsgHandler(packet.PK_ClientLogout, handlers.HandleClientLogout)
	svr.BindMsgHandler(packet.PK_ClientPing, handlers.HandleClientPing)

	// C2C、讨论组消息处理
	svr.BindMsgHandler(packet.PK_C2CTextChat, handlers.HandleC2CTextChat)
	svr.BindMsgHandler(packet.PK_ClientRequestC2COfflineMsg, handlers.HandleClientRequestC2COfflineMsg)
	svr.BindMsgHandler(packet.PK_GroupTextChat, handlers.HandleGroupTextChat)
	svr.BindMsgHandler(packet.PK_ClientRequestGroupOfflineMsg, handlers.HandleClientRequestGroupOfflineMsg)
	svr.BindMsgHandler(packet.PK_ClientRequestGroupInfo, handlers.HandleClientRequestGroupInfo)

	// 讨论组的建立，加入讨论组，解散讨论组等
	svr.BindMsgHandler(packet.PK_ClientBuildGroup, handlers.HandleClientBuildGroup)
	svr.BindMsgHandler(packet.PK_ClientDisbandGroup, handlers.HandleClientDisbandGroup)
	svr.BindMsgHandler(packet.PK_ClientJoinGroup, handlers.HandleClientJoinGroup)
	svr.BindMsgHandler(packet.PK_ClientLeaveGroup, handlers.HandleClientLeaveGroup)
}
