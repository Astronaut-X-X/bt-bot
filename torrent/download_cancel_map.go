package torrent

import (
	"context"
	"fmt"
	"sync"
)

// 用于同步操作 downloadCancelMap 的互斥锁
var (
	downloadCancelMapLock sync.Mutex
	// 保存每个下载任务对应的取消函数
	downloadCancelMap map[string]context.CancelFunc
)

// 设置一个下载任务的取消函数
func SetDownloadCancel(infoHash string, fileIndex int, cancel context.CancelFunc) {
	downloadCancelMapLock.Lock()
	defer downloadCancelMapLock.Unlock()

	key := fmt.Sprintf("%s-%d", infoHash, fileIndex)
	downloadCancelMap[key] = cancel
}

// 移除一个下载任务的取消函数
func RemoveDownloadCancel(infoHash string, fileIndex int) {
	downloadCancelMapLock.Lock()
	defer downloadCancelMapLock.Unlock()

	key := fmt.Sprintf("%s-%d", infoHash, fileIndex)
	delete(downloadCancelMap, key)
}

// 调用并移除某个下载任务的取消函数，实现任务取消
func DownloadCancel(infoHash string, fileIndex int) {
	downloadCancelMapLock.Lock()
	defer downloadCancelMapLock.Unlock()

	key := fmt.Sprintf("%s-%d", infoHash, fileIndex)
	cancel, ok := downloadCancelMap[key]
	if ok {
		cancel()
		delete(downloadCancelMap, key)
	}
}
