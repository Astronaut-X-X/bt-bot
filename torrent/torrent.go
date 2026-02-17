package torrent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

var (
	globalClientMutex sync.Mutex
	globalClient      *torrent.Client
)

func InitTorrentClient(debug bool) error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = ""
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

func ParseMagnetLink(magnet string) (*torrent.Torrent, error) {

	t, err := globalClient.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	// 等待元信息获取完成（设置超时）
	timeout := 3 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-t.GotInfo():
		// 元信息获取成功
	case <-ctx.Done():
		// 超时
		t.Drop()
		return nil, fmt.Errorf("获取磁力链接元信息超时. Magnet: %s. 等待时长: %v, 错误: %w, 详细错误信息: %+v", magnet, timeout, ctx.Err(), ctx.Err())
	}

	info := t.Info()
	if info == nil {
		t.Drop()
		return nil, fmt.Errorf("无法获取磁力链接元信息，Info为nil. Magnet: %s", magnet)
	}

	t.Drop()

	return t, nil

}
