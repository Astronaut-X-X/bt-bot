package torrent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anacrolix/torrent"
)

// Download 根据 fileIndex 判断是下载单文件还是全部文件。
// - fileIndex == -1: 下载所有文件
// - fileIndex >= 0:  下载指定索引文件
// callback: 进度回调（已完成/总字节）
// cancelCallback: 取消下载回调
// timeoutCallback: 超时回调
func Download(
	magnetLink string,
	fileIndex int,
	callback func(bytesCompleted, totalBytes int64, fileName string),
	cancelCallback func(fileName string),
	timeoutCallback func(fileName string),
	successCallback func(fileName string),
) ([]string, error) {
	log.Println("download", magnetLink, fileIndex)
	if fileIndex == -1 {
		return DownloadAllFile(magnetLink, callback, cancelCallback, timeoutCallback, successCallback)
	} else {
		return DownloadFile(magnetLink, fileIndex, callback, cancelCallback, timeoutCallback, successCallback)
	}
}

// DownloadAllFile 下载种子中的所有文件
func DownloadAllFile(
	magnetLink string,
	callback func(bytesCompleted, totalBytes int64, fileName string),
	cancelCallback func(fileName string),
	timeoutCallback func(fileName string),
	successCallback func(fileName string),
) ([]string, error) {
	log.Println("download all file", magnetLink)

	// 解析磁力链接，获取 Torrent 句柄
	t, err := ParseMagnetLink("magnet:?xt=urn:btih:" + magnetLink)
	if err != nil {
		log.Println("parse magnet link error", err)
		return nil, err
	}

	log.Println("parse magnet link success", t.Info().Name)

	// 计算总大小
	totalLength := t.Info().TotalLength()

	// 所有文件设置为普通优先级，准备下载全部内容
	for i := range t.Files() {
		t.Files()[i].SetPriority(torrent.PiecePriorityNormal)
	}
	t.DownloadAll()

	log.Println("download all file", totalLength)

	// 估算下载所需时间（100KB/s），最低2小时，加30分钟缓冲，最长不超6小时
	minSpeed := int64(100 * 1024) // 100KB/s
	estimatedTime := time.Duration(totalLength/minSpeed) * time.Second
	if estimatedTime < 2*time.Hour {
		estimatedTime = 2 * time.Hour
	}
	estimatedTime += 30 * time.Minute
	maxTimeout := 6 * time.Hour
	if estimatedTime > maxTimeout {
		estimatedTime = maxTimeout
	}

	log.Printf("⏱️ 设置下载超时时间: %v (文件大小: %d 字节)", estimatedTime, totalLength)

	// 创建超时/可取消的 context
	baseCtx, baseCancel := context.WithTimeout(context.Background(), estimatedTime)
	downloadCtx, downloadCancel := context.WithCancel(baseCtx)
	// 保存 cancel 函数到全局 map，用于外部取消
	downloadCancelMap[fmt.Sprintf("%s-%d", magnetLink, -1)] = downloadCancel

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
				cancelCallback("All files")
				log.Println("download all file canceled")
			} else {
				timeoutCallback("All files")
				log.Println("download all file timeout")
			}
			return nil, nil
		default:
			// 查询下载进度
			bytesCompleted := t.BytesCompleted()
			if bytesCompleted >= totalLength {
				// 下载完成
				log.Printf("✅ 文件下载完成: %s (已下载: %d 字节)", t.Info().Name, bytesCompleted)
				successCallback("All files")
				return nil, nil
			}
			// 调用进度回调
			callback(bytesCompleted, totalLength, "All files")
			time.Sleep(5 * time.Second) // 每5秒刷新一次进度
		}
	}
}

// DownloadFile 下载指定索引的单个文件
func DownloadFile(
	magnetLink string,
	fileIndex int,
	callback func(bytesCompleted, totalBytes int64, fileName string),
	cancelCallback func(fileName string),
	timeoutCallback func(fileName string),
	successCallback func(fileName string),
) ([]string, error) {
	log.Printf("⏱️ 开始下载文件: %s (文件索引: %d)", magnetLink, fileIndex)

	// 解析磁力链接，获取 Torrent 句柄
	t, err := ParseMagnetLink("magnet:?xt=urn:btih:" + magnetLink)
	if err != nil {
		return nil, err
	}

	files := t.Files()
	targetFile := files[fileIndex]
	totalLength := targetFile.Length()

	// 只设置目标文件为普通优先级，其他为无优先级
	for i := range files {
		files[i].SetPriority(torrent.PiecePriorityNone)
	}
	targetFile.SetPriority(torrent.PiecePriorityNormal)

	// 开始下载（内部仍会下载所有已设置为有优先级的片段）
	t.DownloadAll()

	log.Printf("⏱️ 开始下载文件: %s (文件索引: %d)", magnetLink, fileIndex)

	// 估算下载所需时间（100KB/s），最低2小时，加30分钟缓冲，最长不超6小时
	minSpeed := int64(100 * 1024) // 100KB/s
	estimatedTime := time.Duration(totalLength/minSpeed) * time.Second
	if estimatedTime < 2*time.Hour {
		estimatedTime = 2 * time.Hour
	}
	estimatedTime += 30 * time.Minute
	maxTimeout := 6 * time.Hour
	if estimatedTime > maxTimeout {
		estimatedTime = maxTimeout
	}

	log.Printf("⏱️ 设置下载超时时间: %v (文件大小: %d 字节)", estimatedTime, totalLength)

	// 创建可取消的 context（支持超时和手动取消）
	baseCtx, baseCancel := context.WithTimeout(context.Background(), estimatedTime)
	downloadCtx, downloadCancel := context.WithCancel(baseCtx)
	downloadCancelMap[fmt.Sprintf("%s-%d", magnetLink, fileIndex)] = downloadCancel

	// 清理函数
	defer func() {
		t.Drop()
		downloadCancel()
		baseCancel()
		delete(downloadCancelMap, magnetLink)
	}()

	// 等待下载完成的主循环
	for {
		select {
		case <-downloadCtx.Done():
			// 判断是否是被外部取消
			if downloadCtx.Err() == context.Canceled {
				log.Println("download file canceled")
				cancelCallback(targetFile.DisplayPath())
			} else {
				log.Println("download file timeout")
				timeoutCallback(targetFile.DisplayPath())
			}
			return nil, nil
		default:
			// 实时获取目标文件的下载进度
			bytesCompleted := targetFile.BytesCompleted()
			if bytesCompleted >= totalLength {
				log.Printf("✅ 文件下载完成: %s (已下载: %d 字节)", t.Info().Name, bytesCompleted)
				successCallback(targetFile.DisplayPath())
				return nil, nil
			}
			// 定时触发进度回调
			callback(bytesCompleted, totalLength, targetFile.DisplayPath())
			time.Sleep(5 * time.Second) // 5秒刷一次进度
		}
	}
}

func CancelDownload(magnetLink string, fileIndex int) {
	downloadCancel, ok := downloadCancelMap[fmt.Sprintf("%s-%d", magnetLink, fileIndex)]
	if ok {
		downloadCancel()
		delete(downloadCancelMap, fmt.Sprintf("%s-%d", magnetLink, fileIndex))
		log.Println("cancel download", magnetLink, fileIndex)
	}
}
