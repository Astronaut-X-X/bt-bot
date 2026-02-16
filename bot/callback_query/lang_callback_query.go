package callback_query

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/database"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func LangCallbackQueryHandler(bot *tgbotapi.BotAPI, udpate *tgbotapi.Update) {
	data := udpate.CallbackQuery.Data

	lang := strings.TrimPrefix(data, "lang_")

	uuid, ok, err := common.GetUserUUID(udpate.CallbackQuery.From.ID)
	if !ok || err != nil {
		return
	}

	user, _, err := common.GetUserAndPermissions(uuid)
	if err != nil {
		return
	}

	username := common.FullName(udpate.CallbackQuery.From)
	user.Language = lang
	err = database.DB.Save(&user).Error
	if err != nil {
		return
	}

	text := i18n.Replace(i18n.Text("start_message", user.Language), map[string]string{
		i18n.StartMessagePlaceholderUserName:        username,
		i18n.StartMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		i18n.StartMessagePlaceholderHelpChannel:     "@bt1bot1channel",
	})

	message := tgbotapi.NewEditMessageText(udpate.CallbackQuery.Message.Chat.ID, udpate.CallbackQuery.Message.MessageID, text)
	message.ReplyMarkup = startReplyMarkup()

	bot.Send(message)
}

func startReplyMarkup() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("🇨🇳中文", "lang_zh")},
			{tgbotapi.NewInlineKeyboardButtonData("🇺🇸English", "lang_en")},
		},
	}
}
