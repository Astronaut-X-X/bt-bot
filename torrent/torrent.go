package torrent

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

var (
	DownloadDir string = "downloads"

	globalClientMutex sync.Mutex
	globalClient      *torrent.Client
)

func init() {
	downloadCancelMap = make(map[string]context.CancelFunc)

	if err := os.MkdirAll(DownloadDir, 0755); err != nil {
		log.Println("创建下载目录失败: ", err)
	}
}

func InitTorrentClient(debug bool) error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = DownloadDir
	cfg.Debug = debug

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}
	globalClient = client

	return nil
}

func CloseTorrentClient() error {
	if globalClient != nil {
		errs := globalClient.Close()
		if len(errs) > 0 {
			return fmt.Errorf("关闭客户端失败: %v", errs)
		}
	}
	return nil
}

func ParseMagnetLink(ctx context.Context, magnet string) (*torrent.Torrent, error) {

	t, err := globalClient.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	// 使用标志位跟踪是否已经 Drop，避免重复调用
	dropped := false
	dropOnce := func() {
		if !dropped {
			dropped = true
			// 安全地调用 Drop，捕获可能的 panic
			defer func() {
				if r := recover(); r != nil {
					// 如果 Drop 失败（torrent 不存在等），忽略 panic
					log.Println("Drop torrent failed: ", r)
				}
			}()
			t.Drop()
		}
	}

	// 等待元信息获取完成（设置超时）
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-t.GotInfo():
		// 元信息获取成功
	case <-ctx.Done():
		// 超时，清理 torrent
		dropOnce()
		return nil, fmt.Errorf("获取磁力链接元信息超时. Magnet: %s. 等待时长: %v, 错误: %w, 详细错误信息: %+v", magnet, timeout, ctx.Err(), ctx.Err())
	}

	info := t.Info()
	if info == nil {
		// Info 为 nil，清理 torrent
		dropOnce()
		return nil, fmt.Errorf("无法获取磁力链接元信息，Info为nil. Magnet: %s", magnet)
	}

	// 成功获取信息，返回 torrent 供调用者使用
	// 调用者负责在使用完后调用 Drop() 清理资源
	return t, nil

}
