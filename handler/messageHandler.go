package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler 处理普通文本消息
func MessageHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	if text != "" {
		reply := tgbotapi.NewMessage(chatID, "收到你的消息: "+text)
		bot.Send(reply)
	}
}

