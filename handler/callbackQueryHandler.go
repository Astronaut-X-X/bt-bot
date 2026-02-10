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

	// 处理停止下载按钮点击
	if data == "stop_download" {
		// 尝试停止下载
		if service.StopDownload() {
			// 更新消息，显示已停止，并移除按钮
			stopText := "🛑 下载已停止"
			editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, stopText)
			editMsg.ReplyMarkup = nil // 移除按钮
			bot.Send(editMsg)
		} else {
			// 没有正在进行的下载，更新消息并移除按钮
			noDownloadText := "ℹ️ 当前没有正在进行的下载任务"
			editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, noDownloadText)
			editMsg.ReplyMarkup = nil // 移除按钮
			bot.Send(editMsg)
		}
		return
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

	// 创建临时下载目录
	downloadDir := filepath.Join("./downloads", infoHash)
	defer func() {
		// 清理下载目录
		// os.RemoveAll(downloadDir)
	}()

	// 先检查本地文件是否存在
	// 可能的文件路径：downloadDir + "/" + fileInfo.Path（完整路径）
	// 或者：downloadDir + "/" + torrentInfo.Name + "/" + fileName
	var localFilePath string
	possiblePaths := []string{
		filepath.Join(downloadDir, fileInfo.Path),              // 完整路径
		filepath.Join(downloadDir, torrentInfo.Name, fileName), // torrent名称/文件名
		filepath.Join(downloadDir, fileName),                   // 直接文件名
	}

	// 检查每个可能的路径
	for _, path := range possiblePaths {
		if stat, err := os.Stat(path); err == nil {
			// 文件存在，检查大小是否匹配
			if stat.Size() == fileInfo.Length {
				localFilePath = path
				break
			}
		}
	}

	// 如果找到本地文件，直接使用
	var filePath string
	var sentMsg tgbotapi.Message
	// var isLocalFile bool
	if localFilePath != "" {
		// 发送消息，告知使用本地文件
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ 找到本地文件: %s\n📦 大小: %s\n\n正在发送...", fileName, formatSize(fileInfo.Length)))
		sentMsg, _ = bot.Send(msg)
		// 直接使用本地文件，跳过下载步骤
		filePath = localFilePath
		// isLocalFile = true
	} else {
		// 发送下载中消息（带停止按钮）
		stopButton := tgbotapi.NewInlineKeyboardButtonData("🛑 停止下载", "stop_download")
		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{stopButton})

		downloadingMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⏳ 正在下载文件: %s\n📦 大小: %s\n\n请稍候...", fileName, formatSize(fileInfo.Length)))
		downloadingMsg.ReplyMarkup = keyboard
		sentMsg, _ = bot.Send(downloadingMsg)

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

			// 创建停止按钮
			stopButton := tgbotapi.NewInlineKeyboardButtonData("🛑 停止下载", "stop_download")
			keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{stopButton})

			editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, progressText)
			editMsg.ReplyMarkup = &keyboard
			bot.Send(editMsg)
		}

		// 下载文件
		var downloadErr error
		filePath, downloadErr = torrentService.DownloadFile(torrentInfo.MagnetLink, fileIndex, downloadDir, progressCallback)
		if downloadErr != nil {
			// 检查是否是用户取消
			errorText := fmt.Sprintf("❌ 下载失败: %v", downloadErr)
			if strings.Contains(downloadErr.Error(), "下载已取消") {
				errorText = "🛑 下载已取消"
			}
			editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
			// 移除按钮（设置为 nil）
			editMsg.ReplyMarkup = nil
			bot.Send(editMsg)
			return
		}

		// 检查文件是否存在
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			errorText := fmt.Sprintf("❌ 文件不存在: %s", filePath)
			editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
			bot.Send(editMsg)
			return
		}
	}
	// 获取文件信息（大小和绝对路径）
	fileStat, statErr := os.Stat(filePath)
	if statErr != nil {
		errorText := fmt.Sprintf("❌ 无法获取文件信息: %v", statErr)
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		bot.Send(editMsg)
		return
	}
	fileSize := fileStat.Size()

	// 获取文件的绝对路径
	absPath, absErr := filepath.Abs(filePath)
	if absErr != nil {
		absPath = filePath // 如果获取绝对路径失败，使用相对路径
	}

	// Telegram Bot API 文件大小限制
	const (
		maxPhotoSize    = 10 * 1024 * 1024 // 10MB
		maxVideoSize    = 50 * 1024 * 1024 // 50MB
		maxDocumentSize = 50 * 1024 * 1024 // 50MB
	)

	// 根据文件类型发送：图片、视频、还是普通文件
	ext := strings.ToLower(filepath.Ext(fileName))
	var fileConfig tgbotapi.Chattable
	var maxSize int64
	var fileTypeName string

	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".webp":
		maxSize = maxPhotoSize
		fileTypeName = "图片"
		if fileSize > maxPhotoSize {
			// 文件太大，发送错误消息
			errorText := fmt.Sprintf("⚠️ 文件过大，无法发送\n\n📄 文件名: %s\n📦 文件大小: %s\n📏 限制大小: %s\n\n💡 提示：Telegram Bot API 限制图片文件最大为 10MB。\n\n📁 文件路径:\n`%s`",
				fileName, formatSize(fileSize), formatSize(maxPhotoSize), absPath)
			editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
			editMsg.ParseMode = tgbotapi.ModeMarkdown
			editMsg.ReplyMarkup = nil
			bot.Send(editMsg)
			return
		}
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(filePath))
		photo.Caption = fmt.Sprintf("📷 %s", fileName)
		fileConfig = photo
	case ".mp4", ".mov", ".mkv", ".webm", ".avi":
		maxSize = maxVideoSize
		fileTypeName = "视频"
		if fileSize > maxVideoSize {
			// 文件太大，发送错误消息
			errorText := fmt.Sprintf("⚠️ 文件过大，无法发送\n\n📄 文件名: %s\n📦 文件大小: %s\n📏 限制大小: %s\n\n💡 提示：Telegram Bot API 限制视频文件最大为 50MB。\n\n📁 文件路径:\n`%s`",
				fileName, formatSize(fileSize), formatSize(maxVideoSize), absPath)
			editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
			editMsg.ParseMode = tgbotapi.ModeMarkdown
			editMsg.ReplyMarkup = nil
			bot.Send(editMsg)
			return
		}
		video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))
		video.Caption = fmt.Sprintf("🎞️ %s", fileName)
		fileConfig = video
	default:
		maxSize = maxDocumentSize
		fileTypeName = "文档"
		if fileSize > maxDocumentSize {
			// 文件太大，发送错误消息
			errorText := fmt.Sprintf("⚠️ 文件过大，无法发送\n\n📄 文件名: %s\n📦 文件大小: %s\n📏 限制大小: %s\n\n💡 提示：Telegram Bot API 限制文档文件最大为 50MB。\n\n📁 文件路径:\n`%s`",
				fileName, formatSize(fileSize), formatSize(maxDocumentSize), absPath)
			editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
			editMsg.ParseMode = tgbotapi.ModeMarkdown
			editMsg.ReplyMarkup = nil
			bot.Send(editMsg)
			return
		}
		doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
		doc.Caption = fmt.Sprintf("📄 %s", fileName)
		fileConfig = doc
	}

	// // 删除下载中消息（如果存在）
	// if sentMsg.MessageID != 0 {
	// 	bot.Request(tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID))
	// }

	// 发送文件
	_, err = bot.Send(fileConfig)
	if err != nil {
		// 检查是否是文件过大错误
		errorText := fmt.Sprintf("❌ 发送文件失败: %v", err)
		if strings.Contains(err.Error(), "Request Entity Too Large") || strings.Contains(err.Error(), "file is too big") {
			errorText = fmt.Sprintf("⚠️ 文件过大，无法发送\n\n📄 文件名: %s\n📦 文件大小: %s\n📏 %s限制: %s\n\n💡 提示：文件超过了 Telegram Bot API 的大小限制。\n\n📁 文件路径:\n`%s`",
				fileName, formatSize(fileSize), fileTypeName, formatSize(maxSize), absPath)
		}
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		editMsg.ParseMode = tgbotapi.ModeMarkdown
		editMsg.ReplyMarkup = nil
		bot.Send(editMsg)
		return
	}

	// // 删除临时文件（仅删除下载的文件，不删除本地已存在的文件）
	// if !isLocalFile {
	// 	os.Remove(filePath)
	// }
}
