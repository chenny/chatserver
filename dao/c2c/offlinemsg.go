//
// 处理每个用户的C2C离线消息, 每条消息的编号都是全局唯一的
// 编号是 消息加上一些额外的信息(时间，uuid等)计算出来的md5值
// uuid的离线消息编号保存在redis中， uuid --> msgid1, msgid2, msgid3, ...
//
package c2c

import (
	"github.com/gansidui/chatserver/utils/convert"
	"github.com/garyburd/redigo/redis"
	"sync"
	"sync/atomic"
	"time"
)

var (
	offlinemsg_redisAddr      string        = "127.0.0.1:6381"
	offlinemsg_redisDBIndex   int32         = 1
	offlinemsg_redisKeyExpire int32         = 7 * 24 * 3600
	offlinemsg_lock           *sync.RWMutex = new(sync.RWMutex)

	offlinemsg_writeCon1, offlinemsg_writeCon2, offlinemsg_writeCon3, offlinemsg_writeCon4     redis.Conn
	offlinemsg_writeMark1, offlinemsg_writeMark2, offlinemsg_writeMark3, offlinemsg_writeMark4 int32 = 0, 0, 0, 0

	offlinemsg_delCon1, offlinemsg_delCon2, offlinemsg_delCon3, offlinemsg_delCon4     redis.Conn
	offlinemsg_delMark1, offlinemsg_delMark2, offlinemsg_delMark3, offlinemsg_delMark4 int32 = 0, 0, 0, 0

	offlinemsg_readCon1, offlinemsg_readCon2, offlinemsg_readCon3, offlinemsg_readCon4     redis.Conn
	offlinemsg_readMark1, offlinemsg_readMark2, offlinemsg_readMark3, offlinemsg_readMark4 int32 = 0, 0, 0, 0
)

func init() {
	var err error
	offlinemsg_writeCon1, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_writeCon2, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_writeCon3, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_writeCon4, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	_, err = offlinemsg_writeCon1.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_writeCon2.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_writeCon3.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_writeCon4.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)

	offlinemsg_delCon1, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_delCon2, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_delCon3, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_delCon4, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	_, err = offlinemsg_delCon1.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_delCon2.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_delCon3.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_delCon4.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)

	offlinemsg_readCon1, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_readCon2, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_readCon3, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	offlinemsg_readCon4, err = redis.DialTimeout("tcp", offlinemsg_redisAddr, time.Second, time.Second, time.Second)
	checkError(err)
	_, err = offlinemsg_readCon1.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_readCon2.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_readCon3.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)
	_, err = offlinemsg_readCon4.Do("SELECT", offlinemsg_redisDBIndex)
	checkError(err)

	go updateDBIndex()
}

// 每天更换一次redis的dbindex
func updateDBIndex() {
	ticker := time.NewTicker(24 * time.Hour)
	for _ = range ticker.C {
		offlinemsg_lock.Lock()

		if atomic.AddInt32(&offlinemsg_redisDBIndex, 1) <= 7 {
			atomic.AddInt32(&offlinemsg_redisKeyExpire, -24*3600)
		} else {
			atomic.StoreInt32(&offlinemsg_redisDBIndex, 1)
			atomic.StoreInt32(&offlinemsg_redisKeyExpire, 7*24*3600)
		}

		offlinemsg_writeCon1.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_writeCon2.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_writeCon3.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_writeCon4.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))

		offlinemsg_delCon1.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_delCon2.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_delCon3.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_delCon4.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))

		offlinemsg_readCon1.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_readCon2.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_readCon3.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))
		offlinemsg_readCon4.Do("SELECT", atomic.LoadInt32(&offlinemsg_redisDBIndex))

		offlinemsg_lock.Unlock()
	}
}

// 程序关闭时打扫一下
func Clean() {
	offlinemsg_writeCon1.Do("SAVE")
	offlinemsg_delCon1.Do("SAVE")
	offlinemsg_readCon1.Do("SAVE")
	id2msgClean()
}

// 增加uuid的离线消息
func AddMsg(uuid string, msg string) {
	msgid := convert.StringToMd5(uuid + msg + time.Now().String())
	addMsg(msgid, msg)

	if atomic.CompareAndSwapInt32(&offlinemsg_writeMark1, 0, 1) {
		offlinemsg_writeCon1.Do("RPUSH", uuid, msgid)
		offlinemsg_writeCon1.Do("EXPIRE", uuid, offlinemsg_redisKeyExpire)
		atomic.StoreInt32(&offlinemsg_writeMark1, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_writeMark2, 0, 1) {
		offlinemsg_writeCon2.Do("RPUSH", uuid, msgid)
		offlinemsg_writeCon2.Do("EXPIRE", uuid, offlinemsg_redisKeyExpire)
		atomic.StoreInt32(&offlinemsg_writeMark2, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_writeMark3, 0, 1) {
		offlinemsg_writeCon3.Do("RPUSH", uuid, msgid)
		offlinemsg_writeCon3.Do("EXPIRE", uuid, offlinemsg_redisKeyExpire)
		atomic.StoreInt32(&offlinemsg_writeMark3, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_writeMark4, 0, 1) {
		offlinemsg_writeCon4.Do("RPUSH", uuid, msgid)
		offlinemsg_writeCon4.Do("EXPIRE", uuid, offlinemsg_redisKeyExpire)
		atomic.StoreInt32(&offlinemsg_writeMark4, 0)
	}
}

// 删除uuid的最早的n条离线消息，total为uuid的离线消息总数
func DeleteMsgs(uuid string, n, total int) {
	if n <= 0 || total <= 0 || total < n {
		return
	}

	var (
		msgids []string
		err    error
	)

	if atomic.CompareAndSwapInt32(&offlinemsg_delMark1, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_delCon1.Do("LRANGE", uuid, 0, n-1))
		offlinemsg_delCon1.Do("LTRIM", uuid, n, total-1)
		atomic.StoreInt32(&offlinemsg_delMark1, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_delMark2, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_delCon2.Do("LRANGE", uuid, 0, n-1))
		offlinemsg_delCon2.Do("LTRIM", uuid, n, total-1)
		atomic.StoreInt32(&offlinemsg_delMark2, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_delMark3, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_delCon3.Do("LRANGE", uuid, 0, n-1))
		offlinemsg_delCon3.Do("LTRIM", uuid, n, total-1)
		atomic.StoreInt32(&offlinemsg_delMark3, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_delMark4, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_delCon4.Do("LRANGE", uuid, 0, n-1))
		offlinemsg_delCon4.Do("LTRIM", uuid, n, total-1)
		atomic.StoreInt32(&offlinemsg_delMark4, 0)
	}

	if err != nil || len(msgids) == 0 {
		return
	}

	for i, _ := range msgids {
		deleteMsg(msgids[i])
	}
}

// 获取uuid的最早的n条离线消息
func GetMsgs(uuid string, n int) (msgs []string) {
	if n <= 0 {
		return
	}

	var (
		msgids []string
		err    error
	)

	if atomic.CompareAndSwapInt32(&offlinemsg_readMark1, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_readCon1.Do("LRANGE", uuid, 0, n-1))
		atomic.StoreInt32(&offlinemsg_readMark1, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_readMark2, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_readCon2.Do("LRANGE", uuid, 0, n-1))
		atomic.StoreInt32(&offlinemsg_readMark2, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_readMark3, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_readCon3.Do("LRANGE", uuid, 0, n-1))
		atomic.StoreInt32(&offlinemsg_readMark3, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_readMark4, 0, 1) {
		msgids, err = redis.Strings(offlinemsg_readCon4.Do("LRANGE", uuid, 0, n-1))
		atomic.StoreInt32(&offlinemsg_readMark4, 0)
	}

	if err != nil || len(msgids) == 0 {
		return
	}

	for i, _ := range msgids {
		msgs = append(msgs, getMsg(msgids[i]))
	}
	return
}

// 获取uuid的离线消息数量
func GetMsgNum(uuid string) int {
	offlinemsg_lock.RLock()
	defer offlinemsg_lock.RUnlock()
	var (
		n   int
		err error
	)

	if atomic.CompareAndSwapInt32(&offlinemsg_readMark1, 0, 1) {
		n, err = redis.Int(offlinemsg_readCon1.Do("LLEN", uuid))
		atomic.StoreInt32(&offlinemsg_readMark1, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_readMark2, 0, 1) {
		n, err = redis.Int(offlinemsg_readCon2.Do("LLEN", uuid))
		atomic.StoreInt32(&offlinemsg_readMark2, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_readMark3, 0, 1) {
		n, err = redis.Int(offlinemsg_readCon3.Do("LLEN", uuid))
		atomic.StoreInt32(&offlinemsg_readMark3, 0)

	} else if atomic.CompareAndSwapInt32(&offlinemsg_readMark4, 0, 1) {
		n, err = redis.Int(offlinemsg_readCon4.Do("LLEN", uuid))
		atomic.StoreInt32(&offlinemsg_readMark4, 0)
	}

	if err != nil {
		return 0
	}
	return n
}
