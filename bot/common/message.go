package common

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func ParseMessageText(update *tgbotapi.Update) string {
	message := update.Message
	if message == nil {
		return ""
	}

	return message.Text
}

func ParseMessageChatId(update *tgbotapi.Update) int64 {
	message := update.Message
	if message == nil {
		return 0
	}
	return message.Chat.ID
}

func ParseCallbackQueryChatId(update *tgbotapi.Update) int64 {
	callbackQuery := update.CallbackQuery
	if callbackQuery == nil {
		return 0
	}
	return callbackQuery.Message.Chat.ID
}
