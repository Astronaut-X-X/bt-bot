package callback_query

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CallbackQueryHandler(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	// CallbackQueryHandler: &{ID:5913080342474078341 From:yob00000000 Message:0xc0001cc008 InlineMessageID: ChatInstance:6889523692344572250 Data:lang_zh GameShortName:}

	switch {
	case strings.HasPrefix(callback.Data, "lang_"):
		LangCallbackQueryHandler(bot, callback)
	default:
		return
	}
}
