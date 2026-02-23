package middleware

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"errors"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	_downloadMaplock sync.Mutex
	_downloadMap     = map[string]bool{}
)

func DownloadMiddleWare(next func(bot *tgbotapi.BotAPI, update *tgbotapi.Update)) func(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
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

		cacheKey := fmt.Sprintf("%d", userId)
		if isDownloadInCache(cacheKey) {
			messageText := i18n.Text(i18n.DownloadAlreadyDownloadingMessageCode, user.Language)
			reply := tgbotapi.NewMessage(chatID, messageText)
			bot.Send(reply)
			return
		}

		addDownloadToCache(cacheKey)
		defer removeDownloadFromCache(cacheKey)

		next(bot, update)
	}
}

func addDownloadToCache(cacheKey string) {
	_downloadMaplock.Lock()
	defer _downloadMaplock.Unlock()
	_downloadMap[cacheKey] = true
}

func removeDownloadFromCache(cacheKey string) {
	_downloadMaplock.Lock()
	defer _downloadMaplock.Unlock()
	delete(_downloadMap, cacheKey)
}

func isDownloadInCache(cacheKey string) bool {
	_downloadMaplock.Lock()
	defer _downloadMaplock.Unlock()
	return _downloadMap[cacheKey]
}
