package handler

import (
	"fmt"
	"strings"

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
			fileIndex := parts[2]

			// 发送文件信息
			reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("📄 文件索引: %s\n🔑 Info Hash: `%s`\n\n点击了文件按钮，功能开发中...", fileIndex, infoHash))
			reply.ParseMode = tgbotapi.ModeMarkdown
			bot.Send(reply)
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

