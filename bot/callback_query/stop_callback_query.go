package callback_query

import (
	"bt-bot/torrent"
	"errors"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StopCallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	data := update.CallbackQuery.Data

	infoHash, fileIndex, err := parseStopCallbackQueryData(data)
	if err != nil {
		log.Println("parse stop callback query data error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ invalid stop download data"))
		return
	}

	torrent.CancelDownload(infoHash, fileIndex)
}

func parseStopCallbackQueryData(data string) (string, int, error) {
	split := strings.Split(data, "_")
	if split[0]+"_"+split[1] != "stop_download" {
		return "", 0, errors.New("invalid data")
	}
	infoHash := split[2]
	fileIndex, err := strconv.Atoi(split[3])
	if err != nil {
		return "", 0, err
	}
	return infoHash, fileIndex, nil
}
