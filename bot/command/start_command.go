package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := common.ParseUserId(update)
	chatID := common.ParseMessageChatId(update)
	userName := common.ParseFullName(update)

	user, err := common.User(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	message := i18n.Replace(i18n.Text(i18n.StartMessageCode, user.Language), map[string]string{
		i18n.StartMessagePlaceholderUserName:           userName,
		i18n.StartMessagePlaceholderDownloadChannel:    "@tgqpXOZ2tzXN",
		i18n.StartMessagePlaceholderHelpChannel:        "@bt1bot1channel",
		i18n.StartMessagePlaceholderCooperationContact: "@IIAlbertEinsteinII",
	})

	reply := tgbotapi.NewMessage(chatID, message)
	reply.ReplyMarkup = startReplyMarkup()

	if _, err := bot.Send(reply); err != nil {
		log.Println("Send start message error:", err)
	}
}

func startReplyMarkup() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("🇨🇳中文", "lang_zh")},
			{tgbotapi.NewInlineKeyboardButtonData("🇺🇸English", "lang_en")},
		},
	}
}
