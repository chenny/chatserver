// 负责C2C离线消息的增、删、查
//  msgid --> msg
// 目前存在的问题是这4个conn都在使用中，要不要新建立一个conn？
// 或者通过channel缓存 ? 后期再优化这个
package c2c

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"sync/atomic"
	"time"
)

var (
	id2msg_redisAddr      string = "127.0.0.1:6380"
	id2msg_redisDBIndex   int    = 9
	id2msg_redisKeyExpire int    = 8 * 24 * 3600

	id2msg_writeCon1, id2msg_writeCon2, id2msg_writeCon3, id2msg_writeCon4     redis.Conn
	id2msg_writeMark1, id2msg_writeMark2, id2msg_writeMark3, id2msg_writeMark4 int32 = 0, 0, 0, 0

	id2msg_delCon1, id2msg_delCon2, id2msg_delCon3, id2msg_delCon4     redis.Conn
	id2msg_delMark1, id2msg_delMark2, id2msg_delMark3, id2msg_delMark4 int32 = 0, 0, 0, 0

	id2msg_readCon1, id2msg_readCon2, id2msg_readCon3, id2msg_readCon4     redis.Conn
	id2msg_readMark1, id2msg_readMark2, id2msg_readMark3, id2msg_readMark4 int32 = 0, 0, 0, 0
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	var err error
	id2msg_writeCon1, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_writeCon2, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_writeCon3, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_writeCon4, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	_, err = id2msg_writeCon1.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_writeCon2.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_writeCon3.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_writeCon4.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)

	id2msg_delCon1, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_delCon2, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_delCon3, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_delCon4, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	_, err = id2msg_delCon1.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_delCon2.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_delCon3.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_delCon4.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)

	id2msg_readCon1, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_readCon2, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_readCon3, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	id2msg_readCon4, err = redis.DialTimeout("tcp", id2msg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	_, err = id2msg_readCon1.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_readCon2.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_readCon3.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
	_, err = id2msg_readCon4.Do("SELECT", id2msg_redisDBIndex)
	checkError(err)
}

// 程序关闭时打扫一下
func id2msgClean() {
	id2msg_writeCon1.Do("SAVE")
	id2msg_delCon1.Do("SAVE")
	id2msg_readCon1.Do("SAVE")
}

// 增加(msgid, msg)
func addMsg(msgid, msg string) {
	if atomic.CompareAndSwapInt32(&id2msg_writeMark1, 0, 1) {
		id2msg_writeCon1.Do("SET", msgid, msg)
		id2msg_writeCon1.Do("EXPIRE", msgid, id2msg_redisKeyExpire)
		atomic.StoreInt32(&id2msg_writeMark1, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_writeMark2, 0, 1) {
		id2msg_writeCon2.Do("SET", msgid, msg)
		id2msg_writeCon2.Do("EXPIRE", msgid, id2msg_redisKeyExpire)
		atomic.StoreInt32(&id2msg_writeMark2, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_writeMark3, 0, 1) {
		id2msg_writeCon3.Do("SET", msgid, msg)
		id2msg_writeCon3.Do("EXPIRE", msgid, id2msg_redisKeyExpire)
		atomic.StoreInt32(&id2msg_writeMark3, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_writeMark4, 0, 1) {
		id2msg_writeCon4.Do("SET", msgid, msg)
		id2msg_writeCon4.Do("EXPIRE", msgid, id2msg_redisKeyExpire)
		atomic.StoreInt32(&id2msg_writeMark4, 0)
	}
}

// 删除(msgid)
func deleteMsg(msgid string) {
	if atomic.CompareAndSwapInt32(&id2msg_delMark1, 0, 1) {
		id2msg_delCon1.Do("DEL", msgid)
		atomic.StoreInt32(&id2msg_delMark1, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_delMark2, 0, 1) {
		id2msg_delCon2.Do("DEL", msgid)
		atomic.StoreInt32(&id2msg_delMark2, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_delMark3, 0, 1) {
		id2msg_delCon3.Do("DEL", msgid)
		atomic.StoreInt32(&id2msg_delMark3, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_delMark4, 0, 1) {
		id2msg_delCon4.Do("DEL", msgid)
		atomic.StoreInt32(&id2msg_delMark4, 0)
	}
}

// 查找(msgid --> msg)
func getMsg(msgid string) string {
	var (
		msg string
		err error
	)

	if atomic.CompareAndSwapInt32(&id2msg_readMark1, 0, 1) {
		msg, err = redis.String(id2msg_readCon1.Do("GET", msgid))
		atomic.StoreInt32(&id2msg_readMark1, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_readMark2, 0, 1) {
		msg, err = redis.String(id2msg_readCon2.Do("GET", msgid))
		atomic.StoreInt32(&id2msg_readMark2, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_readMark3, 0, 1) {
		msg, err = redis.String(id2msg_readCon3.Do("GET", msgid))
		atomic.StoreInt32(&id2msg_readMark3, 0)

	} else if atomic.CompareAndSwapInt32(&id2msg_readMark4, 0, 1) {
		msg, err = redis.String(id2msg_readCon4.Do("GET", msgid))
		atomic.StoreInt32(&id2msg_readMark4, 0)
	}

	if err != nil {
		return ""
	}
	return msg
}
