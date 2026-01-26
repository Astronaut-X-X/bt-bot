package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AboutHandler 处理 /about 命令
func AboutHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	reply := tgbotapi.NewMessage(chatID, "这是一个基础的 Telegram Bot 示例。\n\n使用 Go 和 go-telegram-bot-api 构建。")
	bot.Send(reply)
}

