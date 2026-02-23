package common

import (
	"bt-bot/bot/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendErrorMessage 发送错误消息
func SendErrorMessage(bot *tgbotapi.BotAPI, chatID int64, lang string, err error) {
	// 生成错误消息
	message := i18n.Text(i18n.ErrorCommonMessageCode, lang)
	message = i18n.Replace(message, map[string]string{
		i18n.ErrorMessagePlaceholderErrorMessage: err.Error(),
	})
	// 发送错误消息
	reply := tgbotapi.NewMessage(chatID, message)

	// 发送错误消息
	bot.Send(reply)
}
