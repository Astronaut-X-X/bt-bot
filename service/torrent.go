package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

var (
	globalClientMutex sync.Mutex      // 全局客户端互斥锁
	globalClient      *torrent.Client // 全局客户端（用于避免端口冲突）

	// 下载控制相关
	downloadCancelFunc  context.CancelFunc // 当前下载的取消函数
	downloadCancelMutex sync.Mutex         // 下载取消函数的互斥锁
)

// TorrentService 磁力链接服务
type TorrentService struct {
	client *torrent.Client
	cache  TorrentCache // 缓存服务
}

// NewTorrentService 创建新的 TorrentService 实例
func NewTorrentService(cache TorrentCache) (*TorrentService, error) {
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
		cache:  cache,
	}, nil
}

// extractInfoHashFromMagnet 从磁力链接中提取 InfoHash
func extractInfoHashFromMagnet(magnetLink string) (string, error) {
	// 解析 URL
	u, err := url.Parse(magnetLink)
	if err != nil {
		return "", fmt.Errorf("解析磁力链接失败: %w", err)
	}

	// 查找 xt 参数（通常是 urn:btih:XXXXX）
	xt := u.Query().Get("xt")
	if xt == "" {
		return "", fmt.Errorf("磁力链接中未找到 xt 参数")
	}

	// 提取 InfoHash（格式：urn:btih:XXXXX）
	parts := strings.Split(xt, ":")
	if len(parts) < 3 || parts[0] != "urn" || parts[1] != "btih" {
		return "", fmt.Errorf("无效的 xt 参数格式: %s", xt)
	}

	infoHash := strings.ToLower(parts[2])
	return infoHash, nil
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
	MagnetLink  string            `json:"magnet_link"`  // 磁力链接（用于下载）
}

// TorrentFileInfo 文件信息
type TorrentFileInfo struct {
	Path   string `json:"path"`   // 文件路径
	Length int64  `json:"length"` // 文件大小（字节）
}

// ParseMagnetLink 解析磁力链接内容
func (ts *TorrentService) ParseMagnetLink(magnetLink string) (*TorrentInfo, error) {
	// 尝试从磁力链接中提取 InfoHash
	var infoHash string
	var err error
	if ts.cache != nil {
		infoHash, err = extractInfoHashFromMagnet(magnetLink)
		if err == nil {
			// 先尝试从缓存获取
			cachedInfo, cacheErr := ts.cache.Get(infoHash)
			if cacheErr == nil && cachedInfo != nil {
				// 检查缓存数据是否完整（是否有 MagnetLink）
				if cachedInfo.MagnetLink == "" {
					log.Printf("⚠️ 缓存数据不完整（缺少 MagnetLink），重新解析: InfoHash=%s", infoHash)
					// 缓存数据不完整，继续执行解析流程
				} else {
					log.Printf("✅ 缓存命中: InfoHash=%s, Name=%s", infoHash, cachedInfo.Name)
					return cachedInfo, nil
				}
			}
		}
	}

	// 添加磁力链接到客户端
	t, err := ts.client.AddMagnet(magnetLink)
	if err != nil {
		return nil, fmt.Errorf("添加磁力链接失败: %w; 详细错误信息: %+v; 磁力链接内容: %s", err, err, magnetLink)
	}

	// 等待元信息获取完成（设置超时）
	timeout := 3 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 等待元信息
	select {
	case <-t.GotInfo():
		// 元信息获取成功
	case <-ctx.Done():
		// 超时
		t.Drop()
		return nil, fmt.Errorf("获取磁力链接元信息超时. Magnet: %s. 等待时长: %v, 错误: %w, 详细错误信息: %+v", magnetLink, timeout, ctx.Err(), ctx.Err())
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
		MagnetLink:  magnetLink, // 保存磁力链接用于后续下载
	}

	// 清理资源
	t.Drop()

	// 解析成功后立即存储到缓存
	if ts.cache != nil {
		if err := ts.cache.Set(torrentInfo.InfoHash, torrentInfo); err != nil {
			log.Printf("❌ 缓存存储失败: InfoHash=%s, Error=%v", torrentInfo.InfoHash, err)
		} else {
			log.Printf("💾 缓存已存储: InfoHash=%s, Name=%s, Files=%d", torrentInfo.InfoHash, torrentInfo.Name, len(torrentInfo.Files))
		}
	}

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

	// 获取 InfoHash，先检查缓存
	infoHash := mi.HashInfoBytes().String()
	if ts.cache != nil {
		cachedInfo, cacheErr := ts.cache.Get(infoHash)
		if cacheErr == nil && cachedInfo != nil {
			log.Printf("✅ 缓存命中: InfoHash=%s, Name=%s", infoHash, cachedInfo.Name)
			return cachedInfo, nil
		}
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
		InfoHash:    infoHash,
		Name:        info.Name,
		TotalLength: info.TotalLength(),
		Files:       files,
		Trackers:    trackers,
		PieceLength: info.PieceLength,
		NumPieces:   info.NumPieces(),
	}

	// 解析成功后立即存储到缓存
	if ts.cache != nil {
		if err := ts.cache.Set(torrentInfo.InfoHash, torrentInfo); err != nil {
			log.Printf("❌ 缓存存储失败: InfoHash=%s, Error=%v", torrentInfo.InfoHash, err)
		} else {
			log.Printf("💾 缓存已存储: InfoHash=%s, Name=%s, Files=%d", torrentInfo.InfoHash, torrentInfo.Name, len(torrentInfo.Files))
		}
	}

	return torrentInfo, nil
}

// ProgressCallback 下载进度回调函数
// bytesCompleted: 已下载字节数
// totalBytes: 总字节数
type ProgressCallback func(bytesCompleted, totalBytes int64)

// DownloadFile 下载指定索引的文件
// progressCallback: 可选的进度回调函数，每 5 秒调用一次
func (ts *TorrentService) DownloadFile(magnetLink string, fileIndex int, downloadDir string, progressCallback ProgressCallback) (string, error) {
	// 创建临时下载目录
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("创建下载目录失败: %w", err)
	}

	// 使用全局互斥锁确保同一时间只有一个客户端在运行
	globalClientMutex.Lock()
	defer globalClientMutex.Unlock()

	// 先关闭全局客户端（如果存在），释放端口
	if globalClient != nil {
		log.Printf("🔒 关闭全局客户端以释放端口...")
		globalClient.Close()
		globalClient = nil
		// 等待端口完全释放
		time.Sleep(2 * time.Second)
	}

	// 先关闭当前服务的客户端（如果存在）
	if ts.client != nil {
		log.Printf("🔒 关闭解析客户端以释放端口...")
		ts.client.Close()
		ts.client = nil
		// 等待端口完全释放
		time.Sleep(1 * time.Second)
	}

	// 创建新的客户端用于下载（需要设置 DataDir）
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = downloadDir // 设置下载目录
	cfg.Debug = false

	// 尝试创建下载客户端，如果端口冲突则重试
	var downloadClient *torrent.Client
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		downloadClient, err = torrent.NewClient(cfg)
		if err == nil {
			globalClient = downloadClient // 保存到全局变量
			break
		}

		if strings.Contains(err.Error(), "address already in use") {
			if i < maxRetries-1 {
				waitTime := time.Duration(i+1) * 2 * time.Second
				log.Printf("⚠️ 端口被占用，等待 %v 后重试 (%d/%d)...", waitTime, i+1, maxRetries)
				time.Sleep(waitTime)
			} else {
				return "", fmt.Errorf("创建下载客户端失败（端口冲突，已重试 %d 次）: %w\n提示：请稍后重试，或重启应用", maxRetries, err)
			}
		} else {
			return "", fmt.Errorf("创建下载客户端失败: %w", err)
		}
	}
	defer func() {
		downloadClient.Close()
		globalClient = nil // 清除全局客户端
	}()

	// 添加磁力链接到客户端
	t, err := downloadClient.AddMagnet(magnetLink)
	if err != nil {
		return "", fmt.Errorf("添加磁力链接失败: %w", err)
	}
	defer t.Drop()

	// 等待元信息获取完成
	timeout := 3 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-t.GotInfo():
		// 元信息获取成功
	case <-ctx.Done():
		return "", fmt.Errorf("获取磁力链接元信息超时: %w", ctx.Err())
	}

	// 获取元信息
	info := t.Info()
	if info == nil {
		return "", fmt.Errorf("无法获取磁力链接元信息")
	}

	// 检查文件索引是否有效
	if fileIndex < 0 || fileIndex >= len(info.Files) {
		return "", fmt.Errorf("文件索引无效: %d (共 %d 个文件)", fileIndex, len(info.Files))
	}

	// 获取要下载的文件
	targetFile := info.Files[fileIndex]
	filePath := targetFile.DisplayPath(info)

	// 创建文件路径（使用文件名，避免路径问题）
	fileName := filepath.Base(filePath)
	if fileName == "" || fileName == "." {
		fileName = fmt.Sprintf("file_%d", fileIndex)
	}

	// 下载文件
	log.Printf("📥 开始下载文件: %s (大小: %d 字节)", filePath, targetFile.Length)

	// 获取所有文件对象
	allFiles := t.Files()

	// 先将所有文件的优先级设置为 None（不下载）
	for i := range allFiles {
		allFiles[i].SetPriority(torrent.PiecePriorityNone)
	}

	// 获取目标文件对象
	file := allFiles[fileIndex]

	// 只设置目标文件的优先级为 Normal（下载）
	file.SetPriority(torrent.PiecePriorityNormal)

	// 开始下载（只会下载优先级不为 None 的文件）
	t.DownloadAll()

	// 根据文件大小动态计算超时时间
	// 假设最低下载速度为 100KB/s，至少保留 2 小时的基础时间
	// 对于大文件，按 100KB/s 计算所需时间，再加上 30 分钟缓冲
	minSpeed := int64(100 * 1024) // 100KB/s
	estimatedTime := time.Duration(targetFile.Length/minSpeed) * time.Second
	if estimatedTime < 2*time.Hour {
		estimatedTime = 2 * time.Hour
	}
	estimatedTime += 30 * time.Minute // 增加 30 分钟缓冲
	// 最大超时时间限制为 6 小时
	maxTimeout := 6 * time.Hour
	if estimatedTime > maxTimeout {
		estimatedTime = maxTimeout
	}

	log.Printf("⏱️ 设置下载超时时间: %v (文件大小: %d 字节)", estimatedTime, targetFile.Length)

	// 创建可取消的 context（支持超时和手动取消）
	baseCtx, baseCancel := context.WithTimeout(context.Background(), estimatedTime)
	downloadCtx, downloadCancel := context.WithCancel(baseCtx)

	// 保存取消函数到全局变量（用于 stop 命令）
	downloadCancelMutex.Lock()
	oldCancel := downloadCancelFunc
	downloadCancelFunc = func() {
		downloadCancel()
		baseCancel()
	}
	downloadCancelMutex.Unlock()

	// 清理函数
	defer func() {
		downloadCancel()
		baseCancel()
		downloadCancelMutex.Lock()
		downloadCancelFunc = nil
		downloadCancelMutex.Unlock()
		// 如果之前有旧的取消函数，调用它（清理）
		if oldCancel != nil {
			oldCancel()
		}
	}()

	// 进度更新间隔（每 5 秒更新一次）
	progressUpdateInterval := 5 * time.Second
	lastProgressUpdate := time.Now()

	// 等待下载完成
	for {
		select {
		case <-downloadCtx.Done():
			// 检查是否真的超时或被取消，还是已经下载完成
			bytesCompleted := file.BytesCompleted()
			if bytesCompleted >= targetFile.Length {
				log.Printf("✅ 文件下载完成: %s (已下载: %d 字节)", filePath, bytesCompleted)
				goto downloadComplete
			}
			// 检查是否是被手动取消
			if downloadCtx.Err() == context.Canceled {
				log.Printf("🛑 下载已被用户取消: %s (已下载: %d/%d 字节, %.2f%%)", filePath, bytesCompleted, targetFile.Length, float64(bytesCompleted)*100/float64(targetFile.Length))
				return "", fmt.Errorf("下载已取消 (已下载: %d/%d 字节, %.2f%%)", bytesCompleted, targetFile.Length, float64(bytesCompleted)*100/float64(targetFile.Length))
			}
			return "", fmt.Errorf("下载超时 (已下载: %d/%d 字节, %.2f%%)", bytesCompleted, targetFile.Length, float64(bytesCompleted)*100/float64(targetFile.Length))
		default:
			// 检查下载进度
			bytesCompleted := file.BytesCompleted()
			if bytesCompleted >= targetFile.Length {
				log.Printf("✅ 文件下载完成: %s (已下载: %d 字节)", filePath, bytesCompleted)
				goto downloadComplete
			}

			// 定期更新进度（每 5 秒）
			if progressCallback != nil && time.Since(lastProgressUpdate) >= progressUpdateInterval {
				progressCallback(bytesCompleted, targetFile.Length)
				lastProgressUpdate = time.Now()
			}

			time.Sleep(1 * time.Second)
		}
	}

downloadComplete:
	// 文件下载完成，获取实际文件路径
	// torrent 库会将文件保存到 DataDir + 文件路径
	// 尝试多个可能的路径

	// 路径1: downloadDir + filePath (完整相对路径)
	possiblePaths := []string{
		filepath.Join(downloadDir, filePath),
		filepath.Join(downloadDir, fileName),
	}

	// 如果 filePath 包含目录，也尝试直接使用文件名
	if filePath != fileName {
		// 路径2: downloadDir + torrent名称 + fileName
		// 注意：info.Name 是 torrent 的名称（根目录名）
		if info.Name != "" {
			possiblePaths = append(possiblePaths, filepath.Join(downloadDir, info.Name, fileName))
			possiblePaths = append(possiblePaths, filepath.Join(downloadDir, info.Name, filePath))
		}
	}

	// 尝试每个可能的路径
	var actualPath string
	found := false
	for _, path := range possiblePaths {
		if stat, err := os.Stat(path); err == nil {
			// 检查是否是文件且大小匹配
			if !stat.IsDir() && stat.Size() == targetFile.Length {
				actualPath = path
				found = true
				log.Printf("✅ 找到下载文件: %s (大小: %d 字节)", actualPath, stat.Size())
				break
			}
		}
	}

	// 如果还没找到，尝试递归查找
	if !found {
		log.Printf("⚠️ 在常见路径中未找到文件，尝试递归查找: %s", fileName)
		actualPath = findFileRecursive(downloadDir, fileName, targetFile.Length)
		if actualPath != "" {
			found = true
			log.Printf("✅ 递归查找到文件: %s", actualPath)
		}
	}

	if !found {
		return "", fmt.Errorf("下载的文件不存在。尝试的路径: %v", possiblePaths)
	}

	return actualPath, nil
}

// findFileRecursive 递归查找文件
func findFileRecursive(dir, fileName string, expectedSize int64) string {
	var foundPath string

	filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		// 如果已经找到，跳过后续处理
		if foundPath != "" {
			return filepath.SkipAll
		}

		if walkErr != nil {
			return nil // 忽略错误，继续查找
		}

		// 检查是否是目标文件
		if !info.IsDir() && info.Name() == fileName && info.Size() == expectedSize {
			foundPath = path
			return filepath.SkipAll // 找到后跳过剩余项
		}

		return nil
	})

	return foundPath
}

// Close 关闭服务
func (ts *TorrentService) Close() error {
	if ts.client != nil {
		ts.client.Close()
	}
	return nil
}

// StopDownload 停止当前正在进行的下载
func StopDownload() bool {
	downloadCancelMutex.Lock()
	cancelFunc := downloadCancelFunc
	downloadCancelFunc = nil
	downloadCancelMutex.Unlock()

	if cancelFunc != nil {
		log.Printf("🛑 用户请求停止下载")
		cancelFunc()

		// 尝试关闭全局客户端
		globalClientMutex.Lock()
		if globalClient != nil {
			log.Printf("🔒 关闭下载客户端...")
			globalClient.Close()
			globalClient = nil
		}
		globalClientMutex.Unlock()

		return true
	}
	return false
}
