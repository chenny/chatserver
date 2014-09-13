//
// 处理讨论组和用户之间的关系:
// uuid -> groupid1, groupid2, groupid3, ...
// redisdb: 127.0.0.1:6381
// expire: -1 (表示不设置expire)
//
// groupid --> group_name, group_owner, uuid1, uuid2, uuid3, ...
// redisdb: 127.0.0.1:6382
// expire: -1 (表示不设置expire)
//
// 建立讨论组时，讨论组ID == rand_string(8)
//
package groupinfo

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gansidui/chatserver/config"
	"github.com/gansidui/chatserver/dao/redislist"
	"io"
	"log"
	"strconv"
)

var (
	uuid_redislist    *redislist.RedisListHelper
	groupid_redislist *redislist.RedisListHelper
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 初始化
func Init() {
	// test
	// addr := "127.0.0.1:6381"
	// keyExpire, err := strconv.Atoi("5")
	// checkError(err)

	addr := config.UuidGroupRedisAddr
	keyExpire, err := strconv.Atoi(config.UuidGroupRedisKeyExpire)
	checkError(err)
	uuid_redislist = redislist.NewRedisListHelper(addr, keyExpire)
	uuid_redislist.Init()

	// test
	// addr = "127.0.0.1:6382"
	// keyExpire, err = strconv.Atoi("5")
	// checkError(err)

	addr = config.GroupUuidRedisAddr
	keyExpire, err = strconv.Atoi(config.GroupUuidRedisKeyExpire)
	checkError(err)
	groupid_redislist = redislist.NewRedisListHelper(addr, keyExpire)
	groupid_redislist.Init()
}

// 程序关闭时打扫一下
func Clean() {
	uuid_redislist.Clean()
	groupid_redislist.Clean()
}

// 建立讨论组
func BuildGroup(group_name, group_owner string) (bool, string) {
	buf := make([]byte, 8)
	io.ReadFull(rand.Reader, buf)
	groupid := base64.StdEncoding.EncodeToString(buf)
	ret1 := groupid_redislist.RPush(groupid, group_name)
	ret2 := groupid_redislist.RPush(groupid, group_owner)
	ret3 := uuid_redislist.RPush(group_owner, groupid)
	return ret1 && ret2 && ret3, groupid
}

// 加入讨论组
func JoinGroup(uuid, groupid string) bool {
	// 判断是否已经存在
	if ExistUuidFromGroup(groupid, uuid) {
		return false
	}
	return groupid_redislist.RPush(groupid, uuid) && uuid_redislist.RPush(uuid, groupid)
}

// 退出讨论组（群主只能解散，不能退出）
func ExitGroup(uuid, groupid string) bool {
	if groupid_redislist.GetIndexValue(groupid, 1) == uuid {
		return false
	}
	return groupid_redislist.DelValue(groupid, uuid) && uuid_redislist.DelValue(uuid, groupid)
}

// 解散讨论组
func DisbandGroup(group_owner, groupid string) bool {
	if groupid_redislist.GetIndexValue(groupid, 1) != group_owner {
		return false
	}
	groupid_redislist.DelKey(groupid)

	values := GetAllUuid(groupid)
	for i := 1; i < len(values); i++ {
		uuid_redislist.DelValue(values[i], groupid)
	}
	return true
}

// 得到uuid所有的groupid
func GetAllGroup(uuid string) []string {
	return uuid_redislist.GetRangeValues(uuid, 0, uuid_redislist.GetLength(uuid)-1)
}

// 得到groupid所有的value: (group_name, group_owner, uuid1, uuid2, uuid3, ...)
func GetAllUuid(groupid string) []string {
	return groupid_redislist.GetRangeValues(groupid, 0, groupid_redislist.GetLength(groupid)-1)
}

// 得到讨论组的group_name 和 group_owner
func GetGroupNameAndOwner(groupid string) (string, string) {
	ss := groupid_redislist.GetRangeValues(groupid, 0, 1)
	return ss[0], ss[1]
}

// 判断uuid是否存在讨论组group_id中
func ExistUuidFromGroup(groupid, uuid string) bool {
	return groupid_redislist.ExistValue(groupid, uuid)
}
