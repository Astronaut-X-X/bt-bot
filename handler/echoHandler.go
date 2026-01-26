package handler

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// EchoHandler 处理 /echo 命令
func EchoHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	args := strings.TrimSpace(msg.CommandArguments())

	if args == "" {
		reply := tgbotapi.NewMessage(chatID, "请提供要回显的消息，例如: /echo 你好")
		bot.Send(reply)
	} else {
		reply := tgbotapi.NewMessage(chatID, "你说了: "+args)
		bot.Send(reply)
	}
}

