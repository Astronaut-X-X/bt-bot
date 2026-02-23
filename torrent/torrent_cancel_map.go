package torrent

import (
	"context"
	"fmt"
	"sync"
)

// 用于同步操作 torrentCancelMap 的互斥锁
var (
	torrentCancelMapLock sync.Mutex
	// 保存每个磁力链接对应的取消函数
	torrentCancelMap map[string]context.CancelFunc
)

func init() {
	torrentCancelMap = make(map[string]context.CancelFunc)
}

// 设置一个磁力链接的取消函数
func SetTorrentCancel(magnet string, userId int64, cancel context.CancelFunc) {
	torrentCancelMapLock.Lock()
	defer torrentCancelMapLock.Unlock()

	key := fmt.Sprintf("%s-%d", magnet, userId)
	torrentCancelMap[key] = cancel
}

// 移除一个磁力链接的取消函数
func RemoveTorrentCancel(magnet string, userId int64) {
	torrentCancelMapLock.Lock()
	defer torrentCancelMapLock.Unlock()

	key := fmt.Sprintf("%s-%d", magnet, userId)
	delete(torrentCancelMap, key)
}

// 调用并移除某个磁力链接的取消函数，实现任务取消
func TorrentCancel(magnet string, userId int64) bool {
	torrentCancelMapLock.Lock()
	defer torrentCancelMapLock.Unlock()

	key := fmt.Sprintf("%s-%d", magnet, userId)
	cancel, ok := torrentCancelMap[key]
	if ok {
		cancel()
		delete(torrentCancelMap, key)
	}

	return ok
}
