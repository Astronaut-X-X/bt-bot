package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bt-bot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallbackQueryHandler 处理回调查询（按钮点击）
func CallbackQueryHandler(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	// 先确认回调，避免按钮一直显示加载状态
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	bot.Request(callbackConfig)

	// 解析 callback_data
	data := callback.Data
	chatID := callback.Message.Chat.ID

	// 处理文件按钮点击
	if strings.HasPrefix(data, "file_") {
		// 格式：file_<infoHash>_<index>
		parts := strings.Split(data, "_")
		if len(parts) >= 3 {
			infoHash := parts[1]
			fileIndexStr := parts[2]

			// 解析文件索引
			fileIndex, err := strconv.Atoi(fileIndexStr)
			if err != nil {
				reply := tgbotapi.NewMessage(chatID, "❌ 无效的文件索引")
				bot.Send(reply)
				return
			}

			// 处理文件下载
			handleFileDownload(bot, chatID, infoHash, fileIndex)
			return
		}
	}

	// 处理信息按钮点击
	if strings.HasPrefix(data, "info_") {
		parts := strings.Split(data, "_")
		if len(parts) >= 2 {
			infoHash := parts[1]
			reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("📋 Info Hash: `%s`\n\n文件列表较长，仅显示部分文件按钮。", infoHash))
			reply.ParseMode = tgbotapi.ModeMarkdown
			bot.Send(reply)
			return
		}
	}

	// 未知的回调数据
	reply := tgbotapi.NewMessage(chatID, "❌ 未知的回调操作")
	bot.Send(reply)
}

// handleFileDownload 处理文件下载
func handleFileDownload(bot *tgbotapi.BotAPI, chatID int64, infoHash string, fileIndex int) {
	// 从缓存获取 torrent 信息
	// 注意：torrentCache 在 magnetHandler.go 中定义，由于在同一个包中可以直接访问
	if torrentCache == nil {
		reply := tgbotapi.NewMessage(chatID, "❌ 缓存服务未启用，无法下载文件")
		bot.Send(reply)
		return
	}

	torrentInfo, err := torrentCache.Get(infoHash)
	if err != nil || torrentInfo == nil {
		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ 未找到缓存信息，InfoHash: %s\n\n请先解析磁力链接。", infoHash))
		bot.Send(reply)
		return
	}

	// 检查文件索引
	if fileIndex < 0 || fileIndex >= len(torrentInfo.Files) {
		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ 文件索引无效: %d (共 %d 个文件)", fileIndex, len(torrentInfo.Files)))
		bot.Send(reply)
		return
	}

	// 检查磁力链接是否存在
	if torrentInfo.MagnetLink == "" {
		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ 缓存数据不完整（缺少磁力链接信息）\n\n🔑 InfoHash: `%s`\n\n请重新解析磁力链接以更新缓存。", infoHash))
		reply.ParseMode = tgbotapi.ModeMarkdown
		bot.Send(reply)
		return
	}

	// 获取文件信息
	fileInfo := torrentInfo.Files[fileIndex]
	fileName := filepath.Base(fileInfo.Path)
	if fileName == "" {
		fileName = fmt.Sprintf("file_%d", fileIndex)
	}

	// 发送下载中消息
	downloadingMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⏳ 正在下载文件: %s\n📦 大小: %s\n\n请稍候...", fileName, formatSize(fileInfo.Length)))
	sentMsg, _ := bot.Send(downloadingMsg)

	// 创建临时下载目录
	downloadDir := filepath.Join("./downloads", infoHash)
	defer func() {
		// 清理下载目录
		os.RemoveAll(downloadDir)
	}()

	// 创建 torrent 服务
	torrentService, err := service.NewTorrentService(torrentCache)
	if err != nil {
		errorText := fmt.Sprintf("❌ 创建下载服务失败: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		bot.Send(editMsg)
		return
	}
	defer torrentService.Close()

	// 创建进度更新回调函数
	progressCallback := func(bytesCompleted, totalBytes int64) {
		percentage := float64(bytesCompleted) * 100 / float64(totalBytes)
		progressText := fmt.Sprintf("⏳ 正在下载文件: %s\n📦 大小: %s\n\n📊 进度: %.2f%% (%s / %s)\n\n请稍候...",
			fileName,
			formatSize(fileInfo.Length),
			percentage,
			formatSize(bytesCompleted),
			formatSize(totalBytes))
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, progressText)
		bot.Send(editMsg)
	}

	// 下载文件
	filePath, err := torrentService.DownloadFile(torrentInfo.MagnetLink, fileIndex, downloadDir, progressCallback)
	if err != nil {
		errorText := fmt.Sprintf("❌ 下载失败: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		bot.Send(editMsg)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorText := fmt.Sprintf("❌ 文件不存在: %s", filePath)
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		bot.Send(editMsg)
		return
	}

	// 根据文件类型发送：图片、视频、还是普通文件
	ext := strings.ToLower(filepath.Ext(fileName))
	var fileConfig tgbotapi.Chattable

	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".webp":
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(filePath))
		photo.Caption = fmt.Sprintf("📷 %s", fileName)
		fileConfig = photo
	case ".mp4", ".mov", ".mkv", ".webm", ".avi":
		video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))
		video.Caption = fmt.Sprintf("🎞️ %s", fileName)
		fileConfig = video
	default:
		doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
		doc.Caption = fmt.Sprintf("📄 %s", fileName)
		fileConfig = doc
	}

	// 删除下载中消息
	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID))

	// 发送文件
	_, err = bot.Send(fileConfig)
	if err != nil {
		errorText := fmt.Sprintf("❌ 发送文件失败: %v", err)
		reply := tgbotapi.NewMessage(chatID, errorText)
		bot.Send(reply)
		return
	}

	// 删除临时文件
	os.Remove(filePath)
}
