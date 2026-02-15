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

	user.Language = lang
	err = database.DB.Save(&user).Error
	if err != nil {
		return
	}

	text := i18n.Replace(i18n.Text("callback_message", user.Language), map[string]string{
		i18n.CallbackMessagePlaceholderLanguage: lang,
	})

	message := tgbotapi.NewMessage(udpate.CallbackQuery.Message.Chat.ID, text)
	bot.Send(message)
}
