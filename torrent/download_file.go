package torrent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

var _downloadMapLock sync.Mutex

type ProgressParams struct {
	bytesCompleted int64
	totalBytes     int64
	fileName       string
}

type DownloadParams struct {
	InfoHash  string
	FileIndex int

	ProgressCallback func(ProgressParams)
	CancelCallback   func(t *torrent.Torrent)
	TimeoutCallback  func(t *torrent.Torrent)
	SuccessCallback  func(t *torrent.Torrent)
}

func Download(params DownloadParams) {
	magnetLink := fmt.Sprintf("magnet:?xt=urn:btih:%s", params.InfoHash)

	// 解析磁力链接，获取 Torrent 句柄
	t, err := ParseMagnetLink(magnetLink)
	if err != nil {
		log.Println("parse magnet link error", err)
		return
	}

	// 获取总长度
	totalLength := t.Info().TotalLength()

	// 获取文件列表
	files := t.Files()
	filename := ""
	if params.FileIndex == -1 {
		for i := range files {
			files[i].SetPriority(torrent.PiecePriorityNormal)
		}
		filename = "All files"
	} else {
		for i := range files {
			files[i].SetPriority(torrent.PiecePriorityNone)
		}
		files[params.FileIndex].SetPriority(torrent.PiecePriorityNormal)
		filename = files[params.FileIndex].DisplayPath()
	}
	t.DownloadAll()

	// 估计下载时间
	estimatedTime := estimatedDownloadTime(totalLength)
	baseCtx, baseCancel := context.WithTimeout(context.Background(), estimatedTime)
	downloadCtx, downloadCancel := context.WithCancel(baseCtx)

	// 设置下载取消函数
	SetDownloadCancel(params.InfoHash, params.FileIndex, downloadCancel)
	defer RemoveDownloadCancel(params.InfoHash, params.FileIndex)

	// 清理资源
	defer func() {
		t.Drop()
		downloadCancel()
		baseCancel()
		delete(downloadCancelMap, magnetLink)
	}()

	// 下载主循环
	for {
		select {
		case <-downloadCtx.Done():
			// 判断被取消还是超时
			if downloadCtx.Err() == context.Canceled {
				params.CancelCallback(t)
				log.Println("download all file canceled")
			} else {
				params.TimeoutCallback(t)
				log.Println("download all file timeout")
			}
			return
		default:
			// 查询下载进度
			bytesCompleted := t.BytesCompleted()
			if bytesCompleted >= totalLength {
				// 下载完成
				params.SuccessCallback(t)
				return
			}
			// 调用进度回调
			params.ProgressCallback(ProgressParams{
				bytesCompleted: bytesCompleted,
				totalBytes:     totalLength,
				fileName:       filename,
			})
			time.Sleep(5 * time.Second) // 每5秒刷新一次进度
		}
	}
}

// 估计下载时间
func estimatedDownloadTime(totalLength int64) time.Duration {
	minSpeed := int64(100 * 1024)
	estimatedTime := time.Duration(totalLength/minSpeed)*time.Second + 30*time.Minute
	if estimatedTime < 2*time.Hour {
		estimatedTime = 2 * time.Hour
	}
	if estimatedTime > 6*time.Hour {
		estimatedTime = 6 * time.Hour
	}
	return estimatedTime
}
