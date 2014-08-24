// 用户上线下线处理
// uuid -- 用户唯一标识符(uuid)，记录用户是否在线，保存在内存中

package cid

import (
	"sync"
)

var (
	uuid_lock   *sync.RWMutex   = new(sync.RWMutex)
	uuid_online map[string]bool = make(map[string]bool)
)

// 判断uuid是否在线
func UuidCheckOnline(uuid string) bool {
	uuid_lock.RLock()
	defer uuid_lock.RUnlock()
	return uuid_online[uuid]
}

// uuid上线
func UuidOnLine(uuid string) {
	uuid_lock.Lock()
	defer uuid_lock.Unlock()
	uuid_online[uuid] = true
}

// uuid下线
func UuidOffLine(uuid string) {
	uuid_lock.Lock()
	defer uuid_lock.Unlock()
	uuid_online[uuid] = false
}

// 检查uuid的合法性
func UuidCheckExist(uuid string) bool {
	// ...
	return true
}
