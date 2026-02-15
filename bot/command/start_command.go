package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := update.Message

	chatID := msg.Chat.ID
	userName := common.FullName(msg.From)

	UUID, ok, err := common.GetUserUUID(msg.From.ID)
	if !ok || err != nil {
		if _, _, err = common.CreateUser(msg.From.ID); err != nil {
			log.Println("CreateUser error:", err)
			return
		}
	}

	user, _, err := common.GetUserAndPermissions(UUID)
	if err != nil {
		log.Println("GetUserAndPermissions error:", err)
		return
	}

	message := i18n.Replace(i18n.Text("start_message", user.Language), map[string]string{
		i18n.StartMessagePlaceholderUserName:        userName,
		i18n.StartMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		i18n.StartMessagePlaceholderHelpChannel:     "@bt1bot1channel",
	})

	reply := tgbotapi.NewMessage(chatID, message)
	reply.ReplyMarkup = startReplyMarkup()

	bot.Send(reply)
}

func startReplyMarkup() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("🇨🇳中文", "lang_zh")},
			{tgbotapi.NewInlineKeyboardButtonData("🇺🇸English", "lang_en")},
		},
	}
}
