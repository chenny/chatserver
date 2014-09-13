//
// 功能：将 string --> list 保存到redis中
// 例如： uuid --> msg1, msg2, msg3, ...
//
// 注意:
// (1)根据测试， 采用4个以上的线程访问redis效率最佳
// (2)不能多线程同时使用一个conn（tcp的conn，redis的conn）
//
// 使用方法，构造时需要2个参数：
// Addr: redisdb的ip和port
// KeyExpire: 每个key的存活期，这里每当key所在的list有新增value时，expire会被重置，
// 意味着只要在expire时间内该list没有新增的value，这个list就会被删除
// KeyExire为-1时表示该key没有设置expire，则说明该key永久保存。
//

package redislist

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"sync/atomic"
	"time"
)

type RedisListHelper struct {
	Addr      string // redis的ip:port
	KeyExpire int    // 每个key的expire，单位为秒，例如604800秒(7天)

	writeCon1, writeCon2, writeCon3, writeCon4     redis.Conn
	writeMark1, writeMark2, writeMark3, writeMark4 int32

	delCon1, delCon2, delCon3, delCon4     redis.Conn
	delMark1, delMark2, delMark3, delMark4 int32

	readCon1, readCon2, readCon3, readCon4     redis.Conn
	readMark1, readMark2, readMark3, readMark4 int32

	isValidExpire bool // KeyExire为-1时表示该key没有设置expire，此时isValidExpire为false
}

func NewRedisListHelper(addr string, keyExpire int) *RedisListHelper {
	isValidExpire := true
	if keyExpire == -1 {
		isValidExpire = false
	}

	return &RedisListHelper{
		Addr:       addr,
		KeyExpire:  keyExpire,
		writeMark1: 0, writeMark2: 0, writeMark3: 0, writeMark4: 0,
		delMark1: 0, delMark2: 0, delMark3: 0, delMark4: 0,
		readMark1: 0, readMark2: 0, readMark3: 0, readMark4: 0,
		isValidExpire: isValidExpire,
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 初始化
func (this *RedisListHelper) Init() {
	var err error
	this.writeCon1, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.writeCon2, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.writeCon3, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.writeCon4, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)

	this.delCon1, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.delCon2, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.delCon3, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.delCon4, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)

	this.readCon1, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.readCon2, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.readCon3, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
	this.readCon4, err = redis.DialTimeout("tcp", this.Addr, time.Second, time.Second, time.Second)
	checkError(err)
}

// 程序关闭时保存，不执行也是可以的，因为redis-server会定时保存
func (this *RedisListHelper) Clean() {
	this.writeCon1.Do("SAVE")
	this.delCon1.Do("SAVE")
	this.readCon1.Do("SAVE")
}

// 根据operator来决定从头部还是尾部插入
func (this *RedisListHelper) push(operator, key, value string) bool {
	if atomic.CompareAndSwapInt32(&this.writeMark1, 0, 1) {
		this.writeCon1.Do(operator, key, value)
		if this.isValidExpire {
			this.writeCon1.Do("EXPIRE", key, this.KeyExpire)
		}
		atomic.StoreInt32(&this.writeMark1, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.writeMark2, 0, 1) {
		this.writeCon2.Do(operator, key, value)
		if this.isValidExpire {
			this.writeCon2.Do("EXPIRE", key, this.KeyExpire)
		}
		atomic.StoreInt32(&this.writeMark2, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.writeMark3, 0, 1) {
		this.writeCon3.Do(operator, key, value)
		if this.isValidExpire {
			this.writeCon3.Do("EXPIRE", key, this.KeyExpire)
		}
		atomic.StoreInt32(&this.writeMark3, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.writeMark4, 0, 1) {
		this.writeCon4.Do(operator, key, value)
		if this.isValidExpire {
			this.writeCon4.Do("EXPIRE", key, this.KeyExpire)
		}
		atomic.StoreInt32(&this.writeMark4, 0)
		return true
	}

	return false
}

// 在名称为key的list头部添加一个值为value的元素
func (this *RedisListHelper) LPush(key, value string) bool {
	return this.push("LPUSH", key, value)
}

// 在名称为key的list尾部添加一个值为value的元素
func (this *RedisListHelper) RPush(key, value string) bool {
	return this.push("RPUSH", key, value)
}

// 删除名称为key的list的index在[start,end]区间的所有value, total为list的长度
// 绝大部分情况下start不会超过3， 那么先截取[end+1, total-1], 再往头部插入[0, start-1]
func (this *RedisListHelper) DelRangeValues(key string, start, end, total int) bool {
	if key == "" || total < 1 || start < 0 || end < start || total-1 < end {
		return false
	}

	values := make([]string, 0)

	if atomic.CompareAndSwapInt32(&this.delMark1, 0, 1) {
		if start >= 1 {
			values, _ = redis.Strings(this.delCon1.Do("LRANGE", key, 0, start-1))
		}
		this.delCon1.Do("LTRIM", key, end+1, total-1)
		for i := len(values) - 1; i >= 0; i-- {
			this.delCon1.Do("LPUSH", key, values[i])
		}
		atomic.StoreInt32(&this.delMark1, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark2, 0, 1) {
		if start >= 1 {
			values, _ = redis.Strings(this.delCon2.Do("LRANGE", key, 0, start-1))
		}
		this.delCon2.Do("LTRIM", key, end+1, total-1)
		for i := len(values) - 1; i >= 0; i-- {
			this.delCon2.Do("LPUSH", key, values[i])
		}
		atomic.StoreInt32(&this.delMark2, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark3, 0, 1) {
		if start >= 1 {
			values, _ = redis.Strings(this.delCon3.Do("LRANGE", key, 0, start-1))
		}
		this.delCon3.Do("LTRIM", key, end+1, total-1)
		for i := len(values) - 1; i >= 0; i-- {
			this.delCon3.Do("LPUSH", key, values[i])
		}
		atomic.StoreInt32(&this.delMark3, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark4, 0, 1) {
		if start >= 1 {
			values, _ = redis.Strings(this.delCon4.Do("LRANGE", key, 0, start-1))
		}
		this.delCon4.Do("LTRIM", key, end+1, total-1)
		for i := len(values) - 1; i >= 0; i-- {
			this.delCon4.Do("LPUSH", key, values[i])
		}
		atomic.StoreInt32(&this.delMark4, 0)
		return true
	}

	return false
}

// 在名称为key的list中，从尾到头删除1个值为value的元素
func (this *RedisListHelper) DelValue(key, value string) bool {
	if atomic.CompareAndSwapInt32(&this.delMark1, 0, 1) {
		this.delCon1.Do("LREM", key, -1, value)
		atomic.StoreInt32(&this.delMark1, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark2, 0, 1) {
		this.delCon2.Do("LREM", key, -1, value)
		atomic.StoreInt32(&this.delMark2, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark3, 0, 1) {
		this.delCon3.Do("LREM", key, -1, value)
		atomic.StoreInt32(&this.delMark3, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark4, 0, 1) {
		this.delCon4.Do("LREM", key, -1, value)
		atomic.StoreInt32(&this.delMark4, 0)
		return true
	}

	return false
}

// 删除key
func (this *RedisListHelper) DelKey(key string) bool {
	if atomic.CompareAndSwapInt32(&this.delMark1, 0, 1) {
		this.delCon1.Do("DEL", key)
		atomic.StoreInt32(&this.delMark1, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark2, 0, 1) {
		this.delCon2.Do("DEL", key)
		atomic.StoreInt32(&this.delMark2, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark3, 0, 1) {
		this.delCon3.Do("DEL", key)
		atomic.StoreInt32(&this.delMark3, 0)
		return true

	} else if atomic.CompareAndSwapInt32(&this.delMark4, 0, 1) {
		this.delCon4.Do("DEL", key)
		atomic.StoreInt32(&this.delMark4, 0)
		return true
	}

	return false
}

// 在名称为key的list中，判断值为value的元素是否存在
func (this *RedisListHelper) ExistValue(key, value string) bool {
	values := this.GetRangeValues(key, 0, this.GetLength(key)-1)
	for i, _ := range values {
		if values[i] == value {
			return true
		}
	}
	return false
}

// 返回名称为key的list的index在[start,end]区间的所有value
func (this *RedisListHelper) GetRangeValues(key string, start, end int) (values []string) {
	if end < start {
		return
	}

	if atomic.CompareAndSwapInt32(&this.readMark1, 0, 1) {
		values, _ = redis.Strings(this.readCon1.Do("LRANGE", key, start, end))
		atomic.StoreInt32(&this.readMark1, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark2, 0, 1) {
		values, _ = redis.Strings(this.readCon2.Do("LRANGE", key, start, end))
		atomic.StoreInt32(&this.readMark2, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark3, 0, 1) {
		values, _ = redis.Strings(this.readCon3.Do("LRANGE", key, start, end))
		atomic.StoreInt32(&this.readMark3, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark4, 0, 1) {
		values, _ = redis.Strings(this.readCon4.Do("LRANGE", key, start, end))
		atomic.StoreInt32(&this.readMark4, 0)
	}
	return values
}

// 返回名称为key的list中index位置的元素
func (this *RedisListHelper) GetIndexValue(key string, index int) (value string) {
	if atomic.CompareAndSwapInt32(&this.readMark1, 0, 1) {
		value, _ = redis.String(this.readCon1.Do("LINDEX", key, index))
		atomic.StoreInt32(&this.readMark1, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark2, 0, 1) {
		value, _ = redis.String(this.readCon2.Do("LINDEX", key, index))
		atomic.StoreInt32(&this.readMark2, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark3, 0, 1) {
		value, _ = redis.String(this.readCon3.Do("LINDEX", key, index))
		atomic.StoreInt32(&this.readMark3, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark4, 0, 1) {
		value, _ = redis.String(this.readCon4.Do("LINDEX", key, index))
		atomic.StoreInt32(&this.readMark4, 0)
	}
	return value
}

// 返回名称为key的list的长度
func (this *RedisListHelper) GetLength(key string) int {
	var (
		n   int
		err error
	)

	if atomic.CompareAndSwapInt32(&this.readMark1, 0, 1) {
		n, err = redis.Int(this.readCon1.Do("LLEN", key))
		atomic.StoreInt32(&this.readMark1, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark2, 0, 1) {
		n, err = redis.Int(this.readCon2.Do("LLEN", key))
		atomic.StoreInt32(&this.readMark2, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark3, 0, 1) {
		n, err = redis.Int(this.readCon3.Do("LLEN", key))
		atomic.StoreInt32(&this.readMark3, 0)

	} else if atomic.CompareAndSwapInt32(&this.readMark4, 0, 1) {
		n, err = redis.Int(this.readCon4.Do("LLEN", key))
		atomic.StoreInt32(&this.readMark4, 0)
	}

	if err != nil {
		return 0
	}
	return n
}
