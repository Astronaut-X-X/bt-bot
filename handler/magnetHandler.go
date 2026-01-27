package handler

import (
	"fmt"
	"strings"

	"bt-bot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MagnetHandler 处理磁力链接解析
func MagnetHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// 提取磁力链接
	magnetLink := extractMagnetLink(msg.Text)
	if magnetLink == "" {
		reply := tgbotapi.NewMessage(chatID, "❌ 未找到有效的磁力链接。\n\n请发送磁力链接或使用命令：\n/magnet <磁力链接>")
		bot.Send(reply)
		return
	}

	// 发送解析中消息
	processingMsg := tgbotapi.NewMessage(chatID, "⏳ 正在解析磁力链接，请稍候...")
	sentMsg, _ := bot.Send(processingMsg)

	// 创建 torrent 服务
	torrentService, err := service.NewTorrentService()
	if err != nil {
		errorText := fmt.Sprintf("❌ 创建解析服务失败: %v", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		bot.Send(editMsg)
		return
	}
	defer torrentService.Close()

	// 解析磁力链接
	info, err := torrentService.ParseMagnetLink(magnetLink)
	if err != nil {
		errorText := fmt.Sprintf("❌ 解析失败: %v\n\n可能的原因：\n• 网络连接问题\n• 磁力链接无效\n• 超时（3分钟）", err)
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorText)
		bot.Send(editMsg)
		return
	}

	// 格式化结果
	result := formatTorrentInfo(info)

	// 更新消息
	editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, result)
	editMsg.ParseMode = tgbotapi.ModeMarkdown

	// 如果有文件，添加文件按钮
	if len(info.Files) > 0 {
		editMsg.ReplyMarkup = createFileButtons(info.Files, info.InfoHash)
	}

	bot.Send(editMsg)
}

// extractMagnetLink 从文本中提取磁力链接
func extractMagnetLink(text string) string {
	if text == "" {
		return ""
	}

	// 如果是命令，提取参数
	if strings.HasPrefix(text, "/magnet") {
		parts := strings.Fields(text)
		if len(parts) > 1 {
			text = strings.Join(parts[1:], " ")
		} else {
			return ""
		}
	}

	// 查找磁力链接
	if strings.HasPrefix(text, "magnet:") {
		// 提取完整的磁力链接（到行尾或空格）
		spaceIndex := strings.Index(text, " ")
		if spaceIndex > 0 {
			return text[:spaceIndex]
		}
		return text
	}

	// 尝试从文本中查找磁力链接
	start := strings.Index(text, "magnet:")
	if start >= 0 {
		remaining := text[start:]
		spaceIndex := strings.Index(remaining, " ")
		if spaceIndex > 0 {
			return remaining[:spaceIndex]
		}
		return remaining
	}

	return ""
}

// formatTorrentInfo 格式化磁力链接信息
func formatTorrentInfo(info *service.TorrentInfo) string {
	var builder strings.Builder

	builder.WriteString("✅ *磁力链接解析成功*\n\n")

	// 基本信息
	builder.WriteString(fmt.Sprintf("📛 *名称:* %s\n", escapeMarkdown(info.Name)))
	builder.WriteString(fmt.Sprintf("🔑 *Info Hash:* `%s`\n", info.InfoHash))
	builder.WriteString(fmt.Sprintf("📦 *总大小:* %s\n", formatSize(info.TotalLength)))
	builder.WriteString(fmt.Sprintf("🧩 *分片数:* %d\n", info.NumPieces))
	builder.WriteString(fmt.Sprintf("📏 *分片大小:* %s\n\n", formatSize(info.PieceLength)))

	// 文件列表
	if len(info.Files) > 0 {
		builder.WriteString(fmt.Sprintf("📁 *文件列表* (%d 个文件):\n", len(info.Files)))
		maxFiles := 10 // 最多显示10个文件
		for i, file := range info.Files {
			if i >= maxFiles {
				builder.WriteString(fmt.Sprintf("\n... 还有 %d 个文件", len(info.Files)-maxFiles))
				break
			}
			builder.WriteString(fmt.Sprintf("  • %s (%s)\n", escapeMarkdown(file.Path), formatSize(file.Length)))
		}
		builder.WriteString("\n")
	}

	// Tracker 列表
	if len(info.Trackers) > 0 {
		builder.WriteString(fmt.Sprintf("🔗 *Trackers* (%d 个):\n", len(info.Trackers)))
		maxTrackers := 5 // 最多显示5个 tracker
		for i, tracker := range info.Trackers {
			if i >= maxTrackers {
				builder.WriteString(fmt.Sprintf("\n... 还有 %d 个 tracker", len(info.Trackers)-maxTrackers))
				break
			}
			builder.WriteString(fmt.Sprintf("  • `%s`\n", tracker))
		}
	}

	return builder.String()
}

// formatSize 格式化文件大小
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// escapeMarkdown 转义 Markdown 特殊字符
func escapeMarkdown(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// createFileButtons 创建文件按钮
func createFileButtons(files []service.TorrentFileInfo, infoHash string) *tgbotapi.InlineKeyboardMarkup {
	const maxButtons = 50   // Telegram 限制每个键盘最多 100 个按钮，这里设置 50 个文件按钮
	const buttonsPerRow = 1 // 每行显示一个按钮（文件名可能较长）

	var buttons [][]tgbotapi.InlineKeyboardButton

	// 计算要显示的文件数量
	fileCount := len(files)
	if fileCount > maxButtons {
		fileCount = maxButtons
	}

	// 为每个文件创建按钮
	for i := 0; i < fileCount; i++ {
		file := files[i]
		// 获取文件名和大小
		fileName := getFileName(file.Path)
		fileSize := formatSize(file.Length)

		// 组合按钮文本：文件名 + 大小（Telegram 按钮文本限制 64 字符）
		buttonText := fmt.Sprintf("📄 %s (%s)", truncateString(fileName, 40), fileSize)

		// 创建 callback_data，格式：file_<infoHash>_<index>
		callbackData := fmt.Sprintf("file_%s_%d", infoHash, i)

		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// 如果文件数量超过显示限制，添加"查看更多"提示
	if len(files) > maxButtons {
		infoButton := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📋 共 %d 个文件（仅显示前 %d 个）", len(files), maxButtons),
			fmt.Sprintf("info_%s", infoHash),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{infoButton})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return &keyboard
}

// getFileName 从路径中提取文件名
func getFileName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
