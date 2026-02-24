package callback_query

import (
	"strings"

	middleware "bt-bot/bot/middle_ware"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	data := update.CallbackQuery.Data

	switch {
	case strings.HasPrefix(data, "lang_"):
		LangCallbackQueryHandler(bot, update)
	case strings.HasPrefix(data, "file_"):
		next := middleware.DownloadMiddleWare(FileCallbackQueryHandler)
		next = middleware.DailyDownloadMiddleWare(next)
		next(bot, update)
	case strings.HasPrefix(data, "stop_download_"):
		StopCallbackQueryHandler(bot, update)
	case strings.HasPrefix(data, "stop_magnet_"):
		StopMagnetCallbackQueryHandler(bot, update)
	default:
		return
	}
}
