package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TorrentCache 磁力链接缓存接口
type TorrentCache interface {
	// Get 获取缓存
	Get(infoHash string) (*TorrentInfo, error)
	// Set 存储缓存
	Set(infoHash string, info *TorrentInfo) error
	// Delete 删除缓存
	Delete(infoHash string) error
	// Exists 检查是否存在
	Exists(infoHash string) bool
}

// FileCache 文件缓存实现
type FileCache struct {
	cacheDir string
}

// NewFileCache 创建新的文件缓存实例
func NewFileCache(cacheDir string) (*FileCache, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	return &FileCache{
		cacheDir: cacheDir,
	}, nil
}

// getCachePath 获取缓存文件路径
func (fc *FileCache) getCachePath(infoHash string) string {
	// 使用 InfoHash 作为文件名，确保文件名安全
	safeHash := strings.ToLower(infoHash)
	return filepath.Join(fc.cacheDir, safeHash+".json")
}

// Get 获取缓存
func (fc *FileCache) Get(infoHash string) (*TorrentInfo, error) {
	cachePath := fc.getCachePath(infoHash)

	// 读取文件
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在，返回 nil 而不是错误
		}
		return nil, fmt.Errorf("读取缓存文件失败: %w", err)
	}

	// 解析 JSON
	var info TorrentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析缓存文件失败: %w", err)
	}

	return &info, nil
}

// Set 存储缓存
func (fc *FileCache) Set(infoHash string, info *TorrentInfo) error {
	cachePath := fc.getCachePath(infoHash)

	// 序列化为 JSON
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (fc *FileCache) Delete(infoHash string) error {
	cachePath := fc.getCachePath(infoHash)
	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，不算错误
		}
		return fmt.Errorf("删除缓存文件失败: %w", err)
	}
	return nil
}

// Exists 检查是否存在
func (fc *FileCache) Exists(infoHash string) bool {
	cachePath := fc.getCachePath(infoHash)
	_, err := os.Stat(cachePath)
	return err == nil
}
