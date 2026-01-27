package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// TorrentService 磁力链接服务
type TorrentService struct {
	client *torrent.Client
}

// NewTorrentService 创建新的 TorrentService 实例
func NewTorrentService() (*TorrentService, error) {
	// 创建 torrent 客户端配置
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = "" // 不保存文件到磁盘，仅解析元信息
	cfg.Debug = true

	// 创建客户端
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 torrent 客户端失败: %w; 详细错误信息: %+v", err, err)
	}

	return &TorrentService{
		client: client,
	}, nil
}

// TorrentInfo 磁力链接信息
type TorrentInfo struct {
	InfoHash    string            `json:"info_hash"`    // Info Hash
	Name        string            `json:"name"`         // 名称
	TotalLength int64             `json:"total_length"` // 总大小（字节）
	Files       []TorrentFileInfo `json:"files"`        // 文件列表
	Trackers    []string          `json:"trackers"`     // Tracker 列表
	PieceLength int64             `json:"piece_length"` // 分片大小
	NumPieces   int               `json:"num_pieces"`   // 分片数量
}

// TorrentFileInfo 文件信息
type TorrentFileInfo struct {
	Path   string `json:"path"`   // 文件路径
	Length int64  `json:"length"` // 文件大小（字节）
}

// ParseMagnetLink 解析磁力链接内容
func (ts *TorrentService) ParseMagnetLink(magnetLink string) (*TorrentInfo, error) {
	// 添加磁力链接到客户端
	t, err := ts.client.AddMagnet(magnetLink)
	if err != nil {
		return nil, fmt.Errorf("添加磁力链接失败: %w; 详细错误信息: %+v; 磁力链接内容: %s", err, err, magnetLink)
	}

	// 等待元信息获取完成（设置超时）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 等待元信息
	select {
	case <-t.GotInfo():
		// 元信息获取成功
	case <-ctx.Done():
		// 超时
		t.Drop()
		return nil, fmt.Errorf("获取磁力链接元信息超时. Magnet: %s. 等待时长: %v, 错误: %w, 详细错误信息: %+v", magnetLink, 30*time.Second, ctx.Err(), ctx.Err())
	}

	// 获取元信息
	info := t.Info()
	if info == nil {
		t.Drop()
		return nil, fmt.Errorf("无法获取磁力链接元信息，Info为nil. Magnet: %s", magnetLink)
	}

	// 构建文件列表
	files := make([]TorrentFileInfo, 0, len(info.Files))
	for _, file := range info.Files {
		files = append(files, TorrentFileInfo{
			Path:   file.DisplayPath(info),
			Length: file.Length,
		})
	}

	// 获取 tracker 列表
	trackers := make([]string, 0)
	metaInfo := t.Metainfo()
	for _, tier := range metaInfo.AnnounceList {
		for _, tracker := range tier {
			trackers = append(trackers, tracker)
		}
	}
	// 如果没有从 AnnounceList 获取到，尝试从 Announce 获取
	if len(trackers) == 0 && metaInfo.Announce != "" {
		trackers = append(trackers, metaInfo.Announce)
	}

	// 构建返回信息
	torrentInfo := &TorrentInfo{
		InfoHash:    t.InfoHash().String(),
		Name:        info.Name,
		TotalLength: info.TotalLength(),
		Files:       files,
		Trackers:    trackers,
		PieceLength: info.PieceLength,
		NumPieces:   info.NumPieces(),
	}

	// 清理资源
	t.Drop()

	return torrentInfo, nil
}

// ParseTorrentFile 解析 torrent 文件
func (ts *TorrentService) ParseTorrentFile(torrentPath string) (*TorrentInfo, error) {
	// 读取 torrent 文件
	mi, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		// 读取文件是否存在
		if _, statErr := os.Stat(torrentPath); statErr != nil {
			return nil, fmt.Errorf("读取 torrent 文件失败: %w; 详细错误信息: %+v, 目标路径: %s, 文件状态错误: %v", err, err, torrentPath, statErr)
		}
		return nil, fmt.Errorf("读取 torrent 文件失败: %w; 详细错误信息: %+v, 目标路径: %s", err, err, torrentPath)
	}

	// 解析元信息
	info, err := mi.UnmarshalInfo()
	if err != nil {
		return nil, fmt.Errorf("解析 torrent 文件元信息失败: %w; 详细错误信息: %+v, 文件路径: %s", err, err, torrentPath)
	}

	// 构建文件列表
	files := make([]TorrentFileInfo, 0, len(info.Files))
	for _, file := range info.Files {
		files = append(files, TorrentFileInfo{
			Path:   file.DisplayPath(&info),
			Length: file.Length,
		})
	}

	// 获取 tracker 列表
	trackers := make([]string, 0)
	for _, tier := range mi.AnnounceList {
		for _, tracker := range tier {
			trackers = append(trackers, tracker)
		}
	}
	// 如果没有从 AnnounceList 获取到，尝试从 Announce 获取
	if len(trackers) == 0 && mi.Announce != "" {
		trackers = append(trackers, mi.Announce)
	}

	// 构建返回信息
	torrentInfo := &TorrentInfo{
		InfoHash:    mi.HashInfoBytes().String(),
		Name:        info.Name,
		TotalLength: info.TotalLength(),
		Files:       files,
		Trackers:    trackers,
		PieceLength: info.PieceLength,
		NumPieces:   info.NumPieces(),
	}

	return torrentInfo, nil
}

// Close 关闭服务
func (ts *TorrentService) Close() error {
	if ts.client != nil {
		ts.client.Close()
	}
	return nil
}
