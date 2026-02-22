package callback_query

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	data := update.CallbackQuery.Data

	switch {
	case strings.HasPrefix(data, "lang_"):
		LangCallbackQueryHandler(bot, update)
	case strings.HasPrefix(data, "file_"):
		FileCallbackQueryHandler(bot, update)
	case strings.HasPrefix(data, "stop_download_"):
		StopCallbackQueryHandler(bot, update)
	case strings.HasPrefix(data, "info_more_"):
		MoreCallbackQuery(bot, update)
	default:
		return
	}
}
