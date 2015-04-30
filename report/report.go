package report

import (
	"fmt"
	"github.com/chenny/chatserver/config"
	"github.com/chenny/chatserver/utils/convert"
	"github.com/chenny/email"
	"log"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// 统计每种状态的数量

const (
	TryConnect     = iota // 尝试连接
	SuccessConnect        // 连接成功
	OnlineMsg             // 在线消息
	OfflineMsg            // 离线消息
	OnlineUser            // 在线用户
)

var (
	lock          *sync.RWMutex
	mpCount       map[int]int // 统计各个状态
	maxOnlineUser int         // 最高在线人数
)

func init() {
	lock = new(sync.RWMutex)
	mpCount = make(map[int]int)
	maxOnlineUser = 0
}

func reset() {
	for k, _ := range mpCount {
		mpCount[k] = 0
	}
	maxOnlineUser = 0
}

func AddCount(status, count int) {
	lock.Lock()
	defer lock.Unlock()
	mpCount[status] += count
	if status == OnlineUser && mpCount[status] > maxOnlineUser {
		maxOnlineUser = mpCount[status]
	}
}

func Work() {
	ticker := time.NewTicker(time.Duration(config.EmailDuration) * time.Minute)
	startTime := time.Now().Unix()
	for _ = range ticker.C {
		t1 := convert.TimestampToTimeString(startTime)
		t2 := convert.TimestampToTimeString(time.Now().Unix())
		startTime = time.Now().Unix()
		sendEmail(t1, t2)
		reset()
	}
}

// 发送startTime -- endTime这个时间区间的报告
func sendEmail(startTime, endTime string) {
	// 生成邮件正文
	lock.RLock()
	duration := fmt.Sprintf("\n<h2>统计时间区间: [%v ---- %v]</h2>\n", startTime, endTime)
	conCount := fmt.Sprintf("\n<h2>接受连接请求: %v 次成功, %v 次失败</h2>\n", mpCount[SuccessConnect], mpCount[TryConnect])
	msgCount := fmt.Sprintf("\n<h2>在线消息数: %v, 离线消息数: %v</h2>\n", mpCount[OnlineMsg], mpCount[OfflineMsg])
	userCount := fmt.Sprintf("\n<h2>最高在线人数: %v</h2>\n", maxOnlineUser)
	lock.RUnlock()

	content := duration + conCount + msgCount + userCount

	// 发送邮件列表
	ss := strings.Split(config.EmailToList, " ")
	tolist := make([]string, 0)
	for _, v := range ss {
		if v != "" {
			tolist = append(tolist, v)
		}
	}

	e := email.NewEmail()
	e.From = config.EmailAccount
	e.To = tolist
	e.Subject = "聊天服务器每日报告"
	e.HTML = []byte(content)

	emailServer := config.EmailServerAddr + ":" + config.EmailServerPort
	for i := 0; i < 3; i++ {
		err := e.Send(emailServer, smtp.PlainAuth("", e.From, config.EmailPassword, config.EmailServerAddr))
		if err != nil {
			log.Printf("send email failed: %v\r\n", err)
		} else {
			log.Printf("邮件发送成功\r\n")
			break
		}
	}
}
