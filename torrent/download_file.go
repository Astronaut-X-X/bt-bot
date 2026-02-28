package torrent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

var _downloadMapLock sync.Mutex

type ProgressParams struct {
	BytesCompleted int64
	TotalBytes     int64
	FileName       string
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

	// 创建下载上下文
	downloadCtx, downloadCancel := context.WithCancel(context.Background())
	SetDownloadCancel(params.InfoHash, params.FileIndex, downloadCancel)
	defer RemoveDownloadCancel(params.InfoHash, params.FileIndex)

	// 解析磁力链接，获取 Torrent 句柄
	t, err := ParseMagnetLink(downloadCtx, magnetLink)
	if err != nil {
		log.Println("parse magnet link error", err)
		return
	}

	// 获取总长度
	totalLength := int64(0)

	// 获取文件列表
	files := t.Files()
	filename := ""
	var targetFile *torrent.File
	if params.FileIndex == -1 {
		for i := range files {
			files[i].SetPriority(torrent.PiecePriorityNormal)
		}
		filename = "All files"
		totalLength = t.Info().TotalLength()
	} else if params.FileIndex == -2 {
		for i := range files {
			if HasImageExtension(files[i].DisplayPath()) {
				files[i].SetPriority(torrent.PiecePriorityNormal)
			}
		}
		filename = "All images"
		totalLength = t.Info().TotalLength()
	} else if params.FileIndex == -3 {
		for i := range files {
			if HasVideoExtension(files[i].DisplayPath()) {
				files[i].SetPriority(torrent.PiecePriorityNormal)
			}
		}
		filename = "All videos"
		totalLength = t.Info().TotalLength()
	} else {
		for i := range files {
			files[i].SetPriority(torrent.PiecePriorityNone)
		}
		targetFile = files[params.FileIndex]
		targetFile.SetPriority(torrent.PiecePriorityNormal)
		filename = targetFile.DisplayPath()
		totalLength = targetFile.Length()
	}
	t.DownloadAll()

	// 估计下载时间
	estimatedTime := estimatedDownloadTime(totalLength)
	baseCtx, baseCancel := context.WithTimeout(downloadCtx, estimatedTime)

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
		case <-baseCtx.Done():
			// 判断被取消还是超时
			if baseCtx.Err() == context.Canceled {
				params.CancelCallback(t)
				log.Println("download all file canceled")
			} else {
				params.TimeoutCallback(t)
				log.Println("download all file timeout")
			}
			return
		default:
			bytesCompleted := int64(0)
			if params.FileIndex == -1 {
				bytesCompleted = t.BytesCompleted()
			} else {
				bytesCompleted = targetFile.BytesCompleted()
			}

			// 查询下载进度
			if bytesCompleted >= totalLength {
				// 下载完成
				params.SuccessCallback(t)
				return
			}
			// 调用进度回调
			params.ProgressCallback(ProgressParams{
				BytesCompleted: bytesCompleted,
				TotalBytes:     totalLength,
				FileName:       filename,
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

func HasImageExtension(path string) bool {
	return strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".gif") ||
		strings.HasSuffix(path, ".webp") ||
		strings.HasSuffix(path, ".bmp") ||
		strings.HasSuffix(path, ".tiff") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".svg")
}

func HasVideoExtension(path string) bool {
	return strings.HasSuffix(path, ".mp4") ||
		strings.HasSuffix(path, ".avi") ||
		strings.HasSuffix(path, ".mkv") ||
		strings.HasSuffix(path, ".mov") ||
		strings.HasSuffix(path, ".wmv") ||
		strings.HasSuffix(path, ".flv") ||
		strings.HasSuffix(path, ".webm")
}
