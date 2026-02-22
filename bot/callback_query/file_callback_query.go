package callback_query

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/torrent"
	"errors"
	"log"
	"strconv"
	"strings"

	t "github.com/anacrolix/torrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
检查是否有正在下载的文件，如果有则取消下载，并删除下载的文件
defer 删除标识
检查是否当天还有剩余数量
defer 减少当天剩余数量

开始下载文件
发送下载文件消息，带上停止按钮

循环：实时返回下载进度
停止：发送停止消息

下载完成，寻找到文件路径

使用另一个Tg号，在频道里发送消息
需要一个表结构存放之前发送过的消息，提供查询方法
消息是文件解析出的内容，再在评论里回复该文件即可
*/
func FileCallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// 校验下载限制
	user, _, err := common.UserAndPermissions(update.CallbackQuery.From.ID)
	if err != nil {
		log.Println("get user and permissions error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ get user and permissions error"))
		return
	}
	// 并发下载限制
	ok, err := common.DecrementDownloadCount(user.Premium)
	if !ok {
		log.Println("download count not enough")
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ download count not enough"))
		return
	}
	if err != nil {
		log.Println("decrement download count error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ decrement download count error"))
		return
	}
	defer common.IncrementDownloadCount(user.Premium)

	// 每日下载限制
	ok, err = common.RemainDailyDownload(user.Premium)
	if !ok {
		log.Println("daily download count not enough")
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ daily download count not enough"))
		return
	}
	if err != nil {
		log.Println("remain daily download error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ remain daily download error"))
		return
	}
	defer common.DecrementDailyDownloadQuantity(user.Premium)

	data := update.CallbackQuery.Data
	infoHash, fileIndex, err := parseFileCallbackQueryData(data)
	if err != nil {
		log.Println("parse file callback query data error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ invalid download file data"))
		return
	}

	// 文件下载大小限制
	// TODO

	startMessage := i18n.Text(i18n.DownloadStartMessageCode, user.Language)
	startMessage = i18n.Replace(startMessage, map[string]string{
		i18n.DownloadMessagePlaceholderMagnet: infoHash,
	})
	newMessage := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, startMessage)
	newMessage.ReplyMarkup = stopDownloadReplyMarkup(infoHash, fileIndex, user.Language)
	message, err := bot.Send(newMessage)
	if err != nil {
		log.Println("send start message error", err)
		return
	}

	chatID := message.Chat.ID
	messageID := message.MessageID

	log.Println("chatID", chatID)
	log.Println("messageID", messageID)
	log.Println("download file", infoHash, fileIndex)

	params := torrent.DownloadParams{
		InfoHash:  infoHash,
		FileIndex: fileIndex,
		ProgressCallback: func(progressParams torrent.ProgressParams) {
			log.Println("progressParams", progressParams)
		},
		CancelCallback: func(t *t.Torrent) {
			log.Println("cancel callback")
		},
		TimeoutCallback: func(t *t.Torrent) {
			log.Println("timeout callback")
		},
		SuccessCallback: func(t *t.Torrent) {
			log.Println("success callback")
		},
	}

	torrent.Download(params)
	// infoHash,
	// fileIndex,
	// func(bytesCompleted, totalBytes int64, fileName string) {
	// 	downloadProcessingMessage := i18n.Text(i18n.DownloadProcessingMessageCode, user.Language)
	// 	downloadProcessingMessage = i18n.Replace(downloadProcessingMessage, map[string]string{
	// 		i18n.DownloadMessagePlaceholderMagnet:         infoHash,
	// 		i18n.DownloadMessagePlaceholderDownloadFiles:  fileName,
	// 		i18n.DownloadMessagePlaceholderPercent:        utils.FormatPercentage(bytesCompleted, totalBytes),
	// 		i18n.DownloadMessagePlaceholderBytesCompleted: utils.FormatBytesToSizeString(bytesCompleted),
	// 		i18n.DownloadMessagePlaceholderTotalBytes:     utils.FormatBytesToSizeString(totalBytes),
	// 	})
	// 	newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, downloadProcessingMessage)
	// 	newEditMessage.ReplyMarkup = stopDownloadReplyMarkup(infoHash, fileIndex, user.Language)
	// 	bot.Send(newEditMessage)
	// },
	// func(fileName string) {
	// 	downloadFailedMessage := i18n.Text(i18n.DownloadFailedMessageCode, user.Language)
	// 	downloadFailedMessage = i18n.Replace(downloadFailedMessage, map[string]string{
	// 		i18n.DownloadMessagePlaceholderMagnet:        infoHash,
	// 		i18n.DownloadMessagePlaceholderErrorMessage:  "Cancel",
	// 		i18n.DownloadMessagePlaceholderDownloadFiles: fileName,
	// 	})
	// 	newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, downloadFailedMessage)
	// 	bot.Send(newEditMessage)
	// },
	// func(fileName string) {
	// 	downloadFailedMessage := i18n.Text(i18n.DownloadFailedMessageCode, user.Language)
	// 	downloadFailedMessage = i18n.Replace(downloadFailedMessage, map[string]string{
	// 		i18n.DownloadMessagePlaceholderMagnet:        infoHash,
	// 		i18n.DownloadMessagePlaceholderErrorMessage:  "Timeout",
	// 		i18n.DownloadMessagePlaceholderDownloadFiles: fileName,
	// 	})
	// 	newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, downloadFailedMessage)
	// 	bot.Send(newEditMessage)
	// },
	// func(fileName string) {

	// 	downloadSuccessMessage := i18n.Text(i18n.DownloadSuccessMessageCode, user.Language)
	// 	downloadSuccessMessage = i18n.Replace(downloadSuccessMessage, map[string]string{
	// 		i18n.DownloadMessagePlaceholderMagnet:          infoHash,
	// 		i18n.DownloadMessagePlaceholderDownloadFiles:   fileName,
	// 		i18n.DownloadMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
	// 	})
	// 	bot.Send(tgbotapi.NewEditMessageText(chatID, messageID, downloadSuccessMessage))

	// },
}

func parseFileCallbackQueryData(data string) (string, int, error) {
	split := strings.Split(data, "_")
	if split[0] != "file" {
		return "", 0, errors.New("invalid data")
	}
	infoHash := split[1]
	fileIndex, err := strconv.Atoi(split[2])
	if err != nil {
		return "", 0, err
	}
	return infoHash, fileIndex, nil
}

func stopDownloadReplyMarkup(infoHash string, fileIndex int, language string) *tgbotapi.InlineKeyboardMarkup {
	data := "stop_download_" + infoHash + "_" + strconv.Itoa(fileIndex)

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData(i18n.Text(i18n.ButtonStopDownloadCode, language), data)},
		},
	}
}
