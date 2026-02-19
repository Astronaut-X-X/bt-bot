package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HelpCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	user, _, err := common.UserAndPermissions(update.Message.From.ID)
	if err != nil {
		log.Println("UserAndPermissions error:", err)
		return
	}

	message := i18n.Replace(i18n.Text("help_message", user.Language), map[string]string{
		i18n.HelpMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		i18n.HelpMessagePlaceholderHelpChannel:     "@bt1bot1channel",
	})

	reply := tgbotapi.NewMessage(chatID, message)
	bot.Send(reply)
}
