//
// 处理每个用户的C2C离线消息
// uuid的离线消息编号保存在redis中， uuid --> msg1, msg2, msg3, ...
//
package c2cmsg

import (
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/dao/redislist"
	"log"
	"strconv"
)

var c2cmsg_redislist *redislist.RedisListHelper

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 初始化
func Init() {
	// test
	// addr := "127.0.0.1:6379"
	// keyExpire, err := strconv.Atoi("604800")
	// checkError(err)

	addr := config.C2CRedisAddr
	keyExpire, err := strconv.Atoi(config.C2CRedisKeyExpire)
	checkError(err)

	c2cmsg_redislist = redislist.NewRedisListHelper(addr, keyExpire)
	c2cmsg_redislist.Init()
}

// 程序关闭时打扫一下
func Clean() {
	c2cmsg_redislist.Clean()
}

// 增加uuid的离线消息
func AddMsg(uuid string, msg string) {
	c2cmsg_redislist.RPush(uuid, msg)
}

// 删除uuid的最早的n条离线消息，total为uuid的离线消息总数
func DeleteMsgs(uuid string, n, total int) {
	c2cmsg_redislist.DelRangeValues(uuid, 0, n-1, total)
}

// 获取uuid的最早的n条离线消息
func GetMsgs(uuid string, n int) (msgs []string) {
	return c2cmsg_redislist.GetRangeValues(uuid, 0, n-1)
}

// 获取uuid的离线消息数量
func GetMsgNum(uuid string) int {
	return c2cmsg_redislist.GetLength(uuid)
}
