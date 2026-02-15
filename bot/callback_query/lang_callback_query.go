package callback_query

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/database"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func LangCallbackQueryHandler(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	lang := strings.TrimPrefix(data, "lang_")

	uuid, ok, err := common.GetUserUUID(callback.Message.From.ID)
	if !ok || err != nil {
		return
	}

	user, _, err := common.GetUserAndPermissions(uuid)
	if err != nil {
		return
	}

	user.Language = lang
	err = database.DB.Save(&user).Error
	if err != nil {
		return
	}

	message := tgbotapi.NewMessage(callback.Message.Chat.ID, i18n.Text("callback_message", lang))
	bot.Send(message)
}
