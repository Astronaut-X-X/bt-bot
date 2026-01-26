package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UnknownHandler 处理未知命令
func UnknownHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	reply := tgbotapi.NewMessage(chatID, "未知命令。使用 /help 查看可用命令。")
	bot.Send(reply)
}

