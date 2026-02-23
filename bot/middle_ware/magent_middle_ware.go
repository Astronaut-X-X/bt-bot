package middleware

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	_magnetMaplock sync.Mutex
	_magnetMap     = map[string]bool{}
)

func MagnetMiddleWare(next func(bot *tgbotapi.BotAPI, update *tgbotapi.Update)) func(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	return func(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

		chatID := common.ParseMessageChatId(update)
		userId := common.ParseUserId(update)
		cacheKey := fmt.Sprintf("%d", userId)
		user, err := common.User(userId)
		if err != nil {
			messageText := i18n.Text(i18n.ErrorCommonMessageCode, user.Language)
			messageText = i18n.Replace(messageText, map[string]string{
				i18n.ErrorMessagePlaceholderErrorMessage: err.Error(),
			})
			reply := tgbotapi.NewMessage(chatID, messageText)
			bot.Send(reply)
			return
		}

		if isMagnetLinkInCache(cacheKey) {
			messageText := i18n.Text(i18n.MagnetAlreadyParsingMessageCode, user.Language)
			reply := tgbotapi.NewMessage(chatID, messageText)
			bot.Send(reply)
			return
		}

		addMagnetLinkToCache(cacheKey)
		defer removeMagnetLinkFromCache(cacheKey)

		next(bot, update)
	}
}

func addMagnetLinkToCache(cacheKey string) {
	_magnetMaplock.Lock()
	defer _magnetMaplock.Unlock()
	_magnetMap[cacheKey] = true
}

func removeMagnetLinkFromCache(cacheKey string) {
	_magnetMaplock.Lock()
	defer _magnetMaplock.Unlock()
	delete(_magnetMap, cacheKey)
}

func isMagnetLinkInCache(cacheKey string) bool {
	_magnetMaplock.Lock()
	defer _magnetMaplock.Unlock()
	return _magnetMap[cacheKey]
}
