package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := update.Message

	chatID := msg.Chat.ID
	userName := common.FullName(msg.From)

	message := i18n.Replace(i18n.Text("start_message"), map[string]string{
		i18n.StartMessagePlaceholderUserName:        userName,
		i18n.StartMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		i18n.StartMessagePlaceholderHelpChannel:     "@bt1bot1channel",
	})

	reply := tgbotapi.NewMessage(chatID, message)

	bot.Send(reply)
}
