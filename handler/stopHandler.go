package handler

import (
	"bt-bot/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StopHandler 处理 /stop 命令
func StopHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// 尝试停止下载
	if service.StopDownload() {
		reply := tgbotapi.NewMessage(chatID, "🛑 已停止当前下载任务")
		bot.Send(reply)
	} else {
		reply := tgbotapi.NewMessage(chatID, "ℹ️ 当前没有正在进行的下载任务")
		bot.Send(reply)
	}
}

