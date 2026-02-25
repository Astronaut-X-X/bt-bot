package torrent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

var (
	MagnetTimeout int = 5

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

// normalizeMagnetLink 将 magnet 中的 infohash 转为 anacrolix/torrent 可接受的格式。
// 库对 32 字符的 base32 使用 base32.StdEncoding（仅支持大写 A-Z 和 2-7），小写会报错。
func normalizeMagnetLink(magnet string) string {
	const prefix = "urn:btih:"
	i := strings.Index(magnet, prefix)
	if i == -1 {
		return magnet
	}
	start := i + len(prefix)
	end := start
	for end < len(magnet) && magnet[end] != '&' {
		end++
	}
	infohash := magnet[start:end]
	if len(infohash) == 32 {
		return magnet[:start] + strings.ToUpper(infohash) + magnet[end:]
	}
	return magnet
}

func ParseMagnetLink(ctx context.Context, magnet string) (*torrent.Torrent, error) {
	magnet = normalizeMagnetLink(magnet)
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
	timeout := time.Duration(MagnetTimeout) * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-t.GotInfo():
		// 元信息获取成功
	case <-ctx.Done():
		// 超时，清理 torrent
		dropOnce()
		return nil, ctx.Err()
	}

	info := t.Info()
	if info == nil {
		// Info 为 nil，清理 torrent
		dropOnce()
		return nil, errors.New("get torrent info failed")
	}

	// 成功获取信息，返回 torrent 供调用者使用
	// 调用者负责在使用完后调用 Drop() 清理资源
	return t, nil

}
