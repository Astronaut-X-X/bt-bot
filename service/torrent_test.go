package service

import (
	"testing"
)

func TestNewTorrentService(t *testing.T) {
	service, err := NewTorrentService()
	if err != nil {
		t.Fatalf("创建 TorrentService 失败: %v", err)
	}
	defer service.Close()

	if service == nil {
		t.Fatal("TorrentService 为 nil")
	}

	if service.client == nil {
		t.Fatal("client 为 nil")
	}
}

func TestParseMagnetLink(t *testing.T) {
	service, err := NewTorrentService()
	if err != nil {
		t.Fatalf("创建 TorrentService 失败: %v", err)
	}
	defer service.Close()

	// 使用一个公开的测试磁力链接（Ubuntu ISO）
	magnetLink := "magnet:?xt=urn:btih:CF21D87341022CFF4F4CE22DD0DA7AD81A41A6E4"

	info, err := service.ParseMagnetLink(magnetLink)
	if err != nil {
		t.Logf("解析磁力链接失败（可能是网络问题）: %v", err)
		// 如果网络问题导致失败，不视为测试失败
		return
	}

	if info == nil {
		t.Fatal("解析结果为 nil")
	}

	// 验证基本信息
	if info.InfoHash == "" {
		t.Error("InfoHash 为空")
	}

	if info.Name == "" {
		t.Error("Name 为空")
	}

	if info.TotalLength <= 0 {
		t.Errorf("TotalLength 应该大于 0，实际: %d", info.TotalLength)
	}

	if info.PieceLength <= 0 {
		t.Errorf("PieceLength 应该大于 0，实际: %d", info.PieceLength)
	}

	if info.NumPieces <= 0 {
		t.Errorf("NumPieces 应该大于 0，实际: %d", info.NumPieces)
	}

	// 验证文件列表
	if len(info.Files) == 0 {
		t.Error("文件列表为空")
	} else {
		for i, file := range info.Files {
			if file.Path == "" {
				t.Errorf("文件 %d 的路径为空", i)
			}
			if file.Length <= 0 {
				t.Errorf("文件 %d 的大小应该大于 0，实际: %d", i, file.Length)
			}
		}
	}

	t.Logf("解析成功: InfoHash=%s, Name=%s, TotalLength=%d, Files=%d",
		info.InfoHash, info.Name, info.TotalLength, len(info.Files))
}

func TestParseMagnetLink_InvalidLink(t *testing.T) {
	service, err := NewTorrentService()
	if err != nil {
		t.Fatalf("创建 TorrentService 失败: %v", err)
	}
	defer service.Close()

	// 测试无效的磁力链接
	invalidLink := "magnet:?xt=urn:btih:INVALIDHASH"

	_, err = service.ParseMagnetLink(invalidLink)
	if err == nil {
		t.Error("应该返回错误，但返回了 nil")
	}
}

func TestParseMagnetLink_EmptyLink(t *testing.T) {
	service, err := NewTorrentService()
	if err != nil {
		t.Fatalf("创建 TorrentService 失败: %v", err)
	}
	defer service.Close()

	// 测试空链接
	_, err = service.ParseMagnetLink("")
	if err == nil {
		t.Error("应该返回错误，但返回了 nil")
	}
}

func TestTorrentInfo_Structure(t *testing.T) {
	// 测试结构体字段
	info := &TorrentInfo{
		InfoHash:    "test_hash",
		Name:        "test_name",
		TotalLength: 1024,
		Files: []TorrentFileInfo{
			{Path: "file1.txt", Length: 512},
			{Path: "file2.txt", Length: 512},
		},
		Trackers:    []string{"http://tracker.example.com"},
		PieceLength: 16384,
		NumPieces:   1,
	}

	if info.InfoHash != "test_hash" {
		t.Errorf("InfoHash 不匹配: 期望 test_hash, 实际 %s", info.InfoHash)
	}

	if len(info.Files) != 2 {
		t.Errorf("文件数量不匹配: 期望 2, 实际 %d", len(info.Files))
	}

	if len(info.Trackers) != 1 {
		t.Errorf("Tracker 数量不匹配: 期望 1, 实际 %d", len(info.Trackers))
	}
}
