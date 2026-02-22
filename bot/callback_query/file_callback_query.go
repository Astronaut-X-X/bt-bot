package callback_query

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/telegram"
	"bt-bot/torrent"
	"bt-bot/utils"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	t "github.com/anacrolix/torrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 文件下载回调处理
func FileCallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// // 校验下载限制
	user, _, err := common.UserAndPermissions(update.CallbackQuery.From.ID)
	if err != nil {
		log.Println("get user and permissions error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ get user and permissions error"))
		return
	}
	// // 并发下载限制
	// ok, err := common.DecrementDownloadCount(user.Premium)
	// if !ok {
	// 	log.Println("download count not enough")
	// 	bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ download count not enough"))
	// 	return
	// }
	// if err != nil {
	// 	log.Println("decrement download count error", err)
	// 	bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ decrement download count error"))
	// 	return
	// }
	// defer common.IncrementDownloadCount(user.Premium)

	// // 每日下载限制
	// ok, err = common.RemainDailyDownload(user.Premium)
	// if !ok {
	// 	log.Println("daily download count not enough")
	// 	bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ daily download count not enough"))
	// 	return
	// }
	// if err != nil {
	// 	log.Println("remain daily download error", err)
	// 	bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ remain daily download error"))
	// 	return
	// }
	// defer common.DecrementDailyDownloadQuantity(user.Premium)

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

	// 下载进度
	progressCallback := func(params torrent.ProgressParams) {
		message := i18n.Text(i18n.DownloadFailedMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:         infoHash,
			i18n.DownloadMessagePlaceholderDownloadFiles:  params.FileName,
			i18n.DownloadMessagePlaceholderPercent:        utils.FormatPercentage(params.BytesCompleted, params.TotalBytes),
			i18n.DownloadMessagePlaceholderBytesCompleted: utils.FormatBytesToSizeString(params.BytesCompleted),
			i18n.DownloadMessagePlaceholderTotalBytes:     utils.FormatBytesToSizeString(params.TotalBytes),
		})
		newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, message)
		newEditMessage.ReplyMarkup = stopDownloadReplyMarkup(infoHash, fileIndex, user.Language)
		bot.Send(newEditMessage)
	}

	// 下载取消
	cancelCallback := func(t *t.Torrent) {
		message := i18n.Text(i18n.DownloadFailedMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:        infoHash,
			i18n.DownloadMessagePlaceholderErrorMessage:  "Cancel",
			i18n.DownloadMessagePlaceholderDownloadFiles: parseFileName(t, fileIndex),
		})
		newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, message)
		bot.Send(newEditMessage)
	}

	// 下载超时
	timeoutCallback := func(t *t.Torrent) {
		message := i18n.Text(i18n.DownloadFailedMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:        infoHash,
			i18n.DownloadMessagePlaceholderErrorMessage:  "Timeout",
			i18n.DownloadMessagePlaceholderDownloadFiles: parseFileName(t, fileIndex),
		})
		newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, message)
		bot.Send(newEditMessage)
	}

	// 下载成功
	successCallback := func(t *t.Torrent) {

		// 发送下载消息
		sendDownloadMessage(infoHash, fileIndex, t)

		message := i18n.Text(i18n.DownloadSuccessMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:          infoHash,
			i18n.DownloadMessagePlaceholderDownloadFiles:   parseFileName(t, fileIndex),
			i18n.DownloadMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		})
		bot.Send(tgbotapi.NewEditMessageText(chatID, messageID, message))
	}

	params := torrent.DownloadParams{
		InfoHash:         infoHash,
		FileIndex:        fileIndex,
		ProgressCallback: progressCallback,
		CancelCallback:   cancelCallback,
		TimeoutCallback:  timeoutCallback,
		SuccessCallback:  successCallback,
	}

	torrent.Download(params)
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

func parseFileName(t *t.Torrent, fileIndex int) string {
	if fileIndex == -1 {
		return "All files"
	}
	files := t.Files()
	if fileIndex < 0 || fileIndex >= len(files) {
		return "Invalid file index"
	}
	return files[fileIndex].DisplayPath()
}

func sendDownloadMessage(infoHash string, fileIndex int, t *t.Torrent) {
	messageId, ok, _ := common.CheckDownloadMessage(infoHash)
	if !ok {
		messageText := `
#{info_hash}

Magnet: {magnet}
Files:
{files}
		`
		messageText = strings.ReplaceAll(messageText, "{info_hash}", infoHash)
		messageText = strings.ReplaceAll(messageText, "{magnet}", "magnet:?xt=urn:btih:"+infoHash)
		files := t.Info().Files
		filesText := ""
		for _, file := range files {
			filesText += fmt.Sprintf("%s (%s)\n", file.DisplayPath(t.Info()), utils.FormatBytesToSizeString(file.Length))
		}
		messageText = strings.ReplaceAll(messageText, "{files}", filesText)

		messageId_, err := telegram.SendChannelMessage(messageText)
		if err != nil {
			log.Println("send download message error", err)
			return
		}
		messageId = int64(messageId_)
	}

	err := common.RecordDownloadMessage(infoHash, messageId)
	if err != nil {
		log.Println("record download message error", err)
	}

	log.Println("send download message success", messageId)

	sendDownloadComment(infoHash, fileIndex, t, messageId)
}

func sendDownloadComment(infoHash string, fileIndex int, t *t.Torrent, messageId int64) {
	ok, err := common.CheckDownloadComment(infoHash, fileIndex)
	if ok {
		return
	}
	if err != nil {
		log.Println("check download comment error", err)
		return
	}

	filePath := []string{}
	if fileIndex == -1 {
		fileName := t.Info().Name
		if t.Info().NameUtf8 != "" {
			fileName = t.Info().NameUtf8
		}

		if t.Info().IsDir() {
			filePath = append(filePath, filepath.Join(torrent.DownloadDir, fileName))
		} else {
			for _, file := range t.Info().Files {
				filePath = append(filePath, filepath.Join(torrent.DownloadDir, file.DisplayPath(t.Info())))
			}
		}
	} else {
		filePath = append(filePath, t.Info().Files[fileIndex].DisplayPath(t.Info()))
	}

	for _, filePath := range filePath {
		err := telegram.SendCommentMessage(filePath, int(messageId))
		if err != nil {
			log.Println("send download comment error", err)
			return
		}

	}
	if err := common.RecordDownloadComment(infoHash, fileIndex); err != nil {
		log.Println("record download comment error", err)
		return
	}
}
