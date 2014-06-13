package server

import (
	"github.com/gansidui/chatserver/handlers"
	"github.com/gansidui/chatserver/packet"
	"github.com/gansidui/chatserver/utils/convert"
	"github.com/gansidui/chatserver/utils/funcmap"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	exitCh        chan bool        // 结束信号
	waitGroup     *sync.WaitGroup  // 等待goroutine
	funcMap       *funcmap.FuncMap // 映射消息处理函数(uint32 --> func)
	acceptTimeout time.Duration    // 连接超时时间
	readTimeout   time.Duration    // 读超时时间,其实也就是心跳维持时间
	writeTimeout  time.Duration    // 写超时时间
}

func NewServer() *Server {
	return &Server{
		exitCh:        make(chan bool),
		waitGroup:     &sync.WaitGroup{},
		funcMap:       funcmap.NewFuncMap(),
		acceptTimeout: 30,
		readTimeout:   60,
		writeTimeout:  60,
	}
}

func (this *Server) SetAcceptTimeout(acceptTimeout time.Duration) {
	this.acceptTimeout = acceptTimeout
}

func (this *Server) SetReadTimeout(readTimeout time.Duration) {
	this.readTimeout = readTimeout
}

func (this *Server) SetWriteTimeout(writeTimeout time.Duration) {
	this.writeTimeout = writeTimeout
}

func (this *Server) Start(listener *net.TCPListener) {
	log.Printf("Start listen on %v\r\n", listener.Addr())
	this.waitGroup.Add(1)
	defer func() {
		listener.Close()
		this.waitGroup.Done()
	}()

	// 防止恶意连接
	go this.dealSpamConn()

	for {
		select {
		case <-this.exitCh:
			log.Printf("Stop listen on %v\r\n", listener.Addr())
			return
		default:
		}

		listener.SetDeadline(time.Now().Add(this.acceptTimeout))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				// log.Printf("Accept timeout: %v\r\n", opErr)
				continue
			}
			log.Printf("Accept error: %v\r\n", err)
			continue
		}

		// 连接后等待登陆验证
		handlers.ConnMapLoginStatus.Set(conn, time.Now())
		log.Printf("Accept: %v\r\n", conn.RemoteAddr())

		go this.handleClientConn(conn)
	}
}

// 处理恶意连接，定时监测。
// 若conn的loginstatus为nil,则说明conn已经登陆成功。
// 若conn的loginstatus不为nil,则表示loginstatus为该conn连接服务器时的时间戳(time.Time)
// 判断这个时间戳是否已经超过登陆限制时间，若超过，则断开。
func (this *Server) dealSpamConn() {
	limitTime := 60 * time.Second
	ticker := time.NewTicker(limitTime)
	for _ = range ticker.C {
		items := handlers.ConnMapLoginStatus.Items()
		for conn, loginstatus := range items {
			if loginstatus != nil {
				deadline := loginstatus.(time.Time).Add(limitTime)
				if time.Now().After(deadline) {
					conn.(*net.TCPConn).Close()
					handlers.ConnMapLoginStatus.Delete(conn.(*net.TCPConn))
				}
			}
		}
	}
}

func (this *Server) Stop() {
	// close后，所有的exitCh都返回false
	close(this.exitCh)
	this.waitGroup.Wait()
}

func (this *Server) BindMsgHandler(pacType uint32, fn interface{}) error {
	return this.funcMap.Bind(pacType, fn)
}

func (this *Server) handleClientConn(conn *net.TCPConn) {
	this.waitGroup.Add(1)
	defer this.waitGroup.Done()

	receivePackets := make(chan *packet.Packet, 20) // 接收到的包
	chStop := make(chan bool)                       // 通知停止消息处理
	addr := conn.RemoteAddr().String()

	defer func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("Panic: %v\r\n", e)
			}
		}()

		handlers.CloseConn(conn)

		log.Printf("Disconnect: %v\r\n", addr)
		chStop <- true
	}()

	// 处理接收到的包
	go this.handlePackets(conn, receivePackets, chStop)

	// 接收数据
	log.Printf("HandleClient: %v\r\n", addr)
	request := make([]byte, 1024)
	buf := make([]byte, 0)
	var bufLen uint32 = 0

	for {
		select {
		case <-this.exitCh:
			log.Printf("Stop handleClientConn\r\n")
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(this.readTimeout))
		readSize, err := conn.Read(request)
		if err != nil {
			log.Printf("Read failed: %v\r\n", err)
			return
		}

		if readSize > 0 {
			buf = append(buf, request[:readSize]...)
			bufLen += uint32(readSize)

			// 包长(4) + 包体填充长度(2) + 类型(4) + 包体(len([]byte))
			if bufLen >= 10 {
				pacLen := convert.BytesToUint32(buf[0:4])
				pacPadLen := convert.BytesToUint16(buf[4:6])
				if bufLen >= pacLen {
					receivePackets <- &packet.Packet{
						Len:    pacLen,
						PadLen: pacPadLen,
						Type:   convert.BytesToUint32(buf[6:10]),
						Data:   buf[10:pacLen],
					}
					bufLen -= pacLen
					buf = buf[:bufLen]
				}
			}

		}

	}
}

func (this *Server) handlePackets(conn *net.TCPConn, receivePackets <-chan *packet.Packet, chStop <-chan bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("Panic: %v\r\n", e)
		}
	}()

	for {
		select {
		case <-chStop:
			log.Printf("Stop handle receivePackets.\r\n")
			return

		// 消息包处理
		case p := <-receivePackets:
			// 防止模拟的客户端未经登陆就发送消息
			// 如果接收的不是登陆包，则检查是否已经在线，若没在线，则无视消息包，等待登陆检测机制处理。
			if p.Type != packet.PK_ClientLogin {
				if !handlers.ConnMapUuid.Check(conn) {
					continue
				}
			}

			if this.funcMap.Exist(p.Type) {
				this.funcMap.Call(p.Type, conn, p)
			} else {
				log.Printf("Unknown packet type\r\n")
			}
		}
	}
}
