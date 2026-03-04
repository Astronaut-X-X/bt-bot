package common

import (
	"errors"
	"log"
	"time"

	"bt-bot/bot/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const maxRetryOnRateLimit = 3 // 429 时最多重试次数

// SendWithRetry 发送请求，遇到 429 Too Many Requests 时按 retry_after 等待后重试
func SendWithRetry(bot *tgbotapi.BotAPI, c tgbotapi.Chattable) (tgbotapi.Message, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetryOnRateLimit; attempt++ {
		msg, err := bot.Send(c)
		if err == nil {
			return msg, nil
		}
		lastErr = err

		var apiErr tgbotapi.Error
		if !errors.As(err, &apiErr) || apiErr.Code != 429 || apiErr.RetryAfter <= 0 {
			return msg, err
		}

		wait := time.Duration(apiErr.RetryAfter) * time.Second
		log.Printf("Telegram API 429 Too Many Requests, retry after %ds (attempt %d/%d)", apiErr.RetryAfter, attempt+1, maxRetryOnRateLimit+1)
		time.Sleep(wait)
	}
	return tgbotapi.Message{}, lastErr
}

// SendErrorMessage 发送错误消息
func SendErrorMessage(bot *tgbotapi.BotAPI, chatID int64, lang string, err error) {
	// 生成错误消息
	message := i18n.Text(i18n.ErrorCommonMessageCode, lang)
	message = i18n.Replace(message, map[string]string{
		i18n.ErrorMessagePlaceholderErrorMessage: err.Error(),
	})
	// 发送错误消息
	reply := tgbotapi.NewMessage(chatID, message)
	SendWithRetry(bot, reply)
}
