package callback_query

import (
	"github.com/anacrolix/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CallbackQueryHandler(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	log.Printf("CallbackQueryHandler: %+v", callback)
}
