//
// 处理讨论组的离线消息
// 除了expire不同外，其余跟C2C的完全一样
// 讨论组的离线消息太多，所以expire设置的比较小
//
package groupmsg

import (
	"github.com/chenny/chatserver/config"
	"github.com/chenny/chatserver/dao/redislist"
	"log"
	"strconv"
)

var groupmsg_redislist *redislist.RedisListHelper

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 初始化
func Init() {
	// test
	// addr := "127.0.0.1:6380"
	// keyExpire, err := strconv.Atoi("25200")
	// checkError(err)

	addr := config.GroupRedisAddr
	keyExpire, err := strconv.Atoi(config.GroupRedisKeyExpire)
	checkError(err)

	groupmsg_redislist = redislist.NewRedisListHelper(addr, keyExpire)
	groupmsg_redislist.Init()
}

// 程序关闭时打扫一下
func Clean() {
	groupmsg_redislist.Clean()
}

// 增加uuid的离线消息
func AddMsg(uuid string, msg string) {
	groupmsg_redislist.RPush(uuid, msg)
}

// 删除uuid的最早的n条离线消息，total为uuid的离线消息总数
func DeleteMsgs(uuid string, n, total int) {
	groupmsg_redislist.DelRangeValues(uuid, 0, n-1, total)
}

// 获取uuid的最早的n条离线消息
func GetMsgs(uuid string, n int) (msgs []string) {
	return groupmsg_redislist.GetRangeValues(uuid, 0, n-1)
}

// 获取uuid的离线消息数量
func GetMsgNum(uuid string) int {
	return groupmsg_redislist.GetLength(uuid)
}
