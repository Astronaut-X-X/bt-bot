package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartHandler 处理 /start 命令
func StartHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userName := msg.From.FirstName
	if userName == "" {
		userName = "用户"
	}

	reply := tgbotapi.NewMessage(chatID, "你好，"+userName+"！👋\n\n我是你的 Telegram Bot。\n\n使用 /help 查看可用命令。")
	bot.Send(reply)
}

