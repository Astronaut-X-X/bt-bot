package callback_query

import (
	"bt-bot/torrent"
	"errors"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StopMagnetCallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	data := update.CallbackQuery.Data

	infoHash, userId, err := parseStopMagnetCallbackQueryData(data)
	if err != nil {
		log.Println("parse stop magnet callback query data error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ invalid stop magnet data"))
	}

	ok := torrent.TorrentCancel(infoHash, userId)
	if !ok {
		editMsg := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			"❌ 任务已无法取消，可能已完成或不存在。",
		)
		bot.Send(editMsg)
	}
}

func parseStopMagnetCallbackQueryData(data string) (string, int64, error) {
	split := strings.Split(data, "_")
	if split[0]+"_"+split[1] != "stop_magnet" {
		return "", 0, errors.New("invalid data")
	}
	infoHash := split[2]
	userId, err := strconv.ParseInt(split[3], 10, 64)
	if err != nil {
		return "", 0, err
	}
	return infoHash, userId, nil
}
