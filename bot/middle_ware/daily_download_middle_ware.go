package middleware

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func DailyDownloadMiddleWare(next func(bot *tgbotapi.BotAPI, update *tgbotapi.Update)) func(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	return func(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
		chatID := common.ParseCallbackQueryChatId(update)
		userId := common.ParseCallbackQueryUserId(update)
		if userId == 0 {
			common.SendErrorMessage(bot, chatID, "zh", errors.New("invalid user id"))
			return
		}

		user, err := common.User(userId)
		if err != nil {
			common.SendErrorMessage(bot, chatID, user.Language, err)
			return
		}

		remain, err := common.RemainDailyDownloadQuantity(user.Premium)
		if err != nil {
			common.SendErrorMessage(bot, chatID, user.Language, err)
			return
		}

		if remain <= 0 {
			messageText := i18n.Text(i18n.DownloadDailyDownloadCountNotEnoughMessageCode, user.Language)
			reply := tgbotapi.NewMessage(chatID, messageText)
			bot.Send(reply)
			return
		}

		defer common.DecrementDailyDownloadQuantity(user.Premium)

		next(bot, update)
	}
}
