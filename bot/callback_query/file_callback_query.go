package callback_query

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/database/model"
	"bt-bot/telegram"
	"bt-bot/torrent"
	"bt-bot/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	t "github.com/anacrolix/torrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 文件下载回调处理
func FileCallbackQueryHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

	// 解析用户ID和聊天ID
	userId := common.ParseCallbackQueryUserId(update)
	chatID := common.ParseCallbackQueryChatId(update)

	// // 校验下载限制
	user, err := common.User(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	permissions, err := common.Permissions(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	// 解析下载数据
	data := update.CallbackQuery.Data
	infoHash, fileIndex, err := parseFileCallbackQueryData(data)
	if err != nil {
		log.Println("parse file callback query data error", err)
		common.SendWithRetry(bot, tgbotapi.NewMessage(chatID, "❌ invalid download file data"))
		return
	}

	// 文件下载大小限制
	torrentInfo, err := common.GetTorrentInfo(infoHash)
	if err != nil {
		log.Println("get torrent info error", err)
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}
	if checkOverDownloadSize(infoHash, fileIndex, torrentInfo, permissions) {
		messageText := i18n.Text(i18n.DownloadFileDownloadSizeNotEnoughMessageCode, user.Language)
		reply := tgbotapi.NewMessage(chatID, messageText)
		common.SendWithRetry(bot, reply)
		return
	}

	// 发送开始下载消息
	startMessage := i18n.Text(i18n.DownloadStartMessageCode, user.Language)
	startMessage = i18n.Replace(startMessage, map[string]string{
		i18n.DownloadMessagePlaceholderMagnet: infoHash,
	})
	newMessage := tgbotapi.NewMessage(chatID, startMessage)
	newMessage.ReplyMarkup = stopDownloadReplyMarkup(infoHash, fileIndex, user.Language)
	message, err := common.SendWithRetry(bot, newMessage)
	if err != nil {
		log.Println("send start message error", err)
		return
	}
	messageID := message.MessageID

	// 下载进度
	startTime := time.Now()
	progressCallback := func(params torrent.ProgressParams) {
		elapsedTime := time.Since(startTime)
		hours := int(elapsedTime.Hours())
		minutes := int(elapsedTime.Minutes()) % 60
		seconds := int(elapsedTime.Seconds()) % 60
		elapsedTimeString := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		message := i18n.Text(i18n.DownloadProcessingMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:         infoHash,
			i18n.DownloadMessagePlaceholderDownloadFiles:  params.FileName,
			i18n.DownloadMessagePlaceholderPercent:        utils.FormatPercentage(params.BytesCompleted, params.TotalBytes),
			i18n.DownloadMessagePlaceholderBytesCompleted: utils.FormatBytesToSizeString(params.BytesCompleted),
			i18n.DownloadMessagePlaceholderTotalBytes:     utils.FormatBytesToSizeString(params.TotalBytes),
			i18n.DownloadMessagePlaceholderElapsedTime:    elapsedTimeString,
		})
		newEditMessage := tgbotapi.NewEditMessageText(chatID, messageID, message)
		newEditMessage.ReplyMarkup = stopDownloadReplyMarkup(infoHash, fileIndex, user.Language)
		common.SendWithRetry(bot, newEditMessage)
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
		common.SendWithRetry(bot, newEditMessage)
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
		common.SendWithRetry(bot, newEditMessage)
	}

	// 下载成功
	successCallback := func(t *t.Torrent) {
		// 发送文件发送消息
		message := i18n.Text(i18n.DownloadSendFileMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:        infoHash,
			i18n.DownloadMessagePlaceholderDownloadFiles: parseFileName(t, fileIndex),
		})
		common.SendWithRetry(bot, tgbotapi.NewEditMessageText(chatID, messageID, message))

		// 发送下载消息
		sendDownloadMessage(infoHash, fileIndex, t, user.Premium)

		// 发送下载成功消息
		message = i18n.Text(i18n.DownloadSuccessMessageCode, user.Language)
		message = i18n.Replace(message, map[string]string{
			i18n.DownloadMessagePlaceholderMagnet:          infoHash,
			i18n.DownloadMessagePlaceholderDownloadFiles:   parseFileName(t, fileIndex),
			i18n.DownloadMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		})
		common.SendWithRetry(bot, tgbotapi.NewEditMessageText(chatID, messageID, message))
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
	} else if fileIndex == -2 {
		return "All images"
	} else if fileIndex == -3 {
		return "All videos"
	}
	files := t.Files()
	if fileIndex < 0 || fileIndex >= len(files) {
		return "Invalid file index"
	}
	return files[fileIndex].DisplayPath()
}

func sendDownloadMessage(infoHash string, fileIndex int, t *t.Torrent, premium string) {
	messageId, ok, _ := common.CheckDownloadMessage(infoHash)
	if !ok {
		messageText := `
#{info_hash}
{torrent_name}
		`
		messageText = strings.ReplaceAll(messageText, "{info_hash}", infoHash)
		messageText = strings.ReplaceAll(messageText, "{torrent_name}", t.Info().Name)

		messageId_, err := telegram.SendChannelMessage(messageText)
		if err != nil {
			log.Println("send download message error", err)
			return
		}
		messageId = int64(messageId_)

		// 发送下载文件列表
		files := t.Info().Files
		filesText := ""
		for index, file := range files {
			filesText += fmt.Sprintf("%s %d. %s (%s)\n", emojifyFilename(file.DisplayPath(t.Info())), index+1, file.DisplayPath(t.Info()), utils.FormatBytesToSizeString(file.Length))
			if (index+1)%48 == 0 {
				telegram.SendCommentMessageText(filesText, int(messageId))
				filesText = ""
			}
		}
		if filesText != "" {
			telegram.SendCommentMessageText(filesText, int(messageId))
		}

		err = common.RecordDownloadMessage(infoHash, messageId)
		if err != nil {
			log.Println("record download message error", err)
		}
	}

	// 发送下载文件评论
	sendDownloadComment(infoHash, fileIndex, t, messageId, premium)
}

func sendDownloadComment(infoHash string, fileIndex int, t *t.Torrent, messageId int64, premium string) {
	ok, err := common.CheckDownloadComment(infoHash, fileIndex)
	if ok {
		return
	}
	if err != nil {
		log.Println("check download comment error", err)
	}

	filePaths := []string{}
	fileName := t.Info().Name
	if t.Info().NameUtf8 != "" {
		fileName = t.Info().NameUtf8
	}
	downloadDir := filepath.Join(torrent.DownloadDir, fileName)
	if fileIndex == -1 {
		if t.Info().IsDir() {
			filePaths = append(filePaths, downloadDir)
		} else {
			for _, file := range t.Info().Files {
				filePaths = append(filePaths, filepath.Join(downloadDir, file.DisplayPath(t.Info())))
			}
		}
	} else if fileIndex == -2 {
		for _, file := range t.Info().Files {
			if torrent.HasImageExtension(file.DisplayPath(t.Info())) {
				filePaths = append(filePaths, filepath.Join(downloadDir, file.DisplayPath(t.Info())))
			}
		}
	} else if fileIndex == -3 {
		for _, file := range t.Info().Files {
			if torrent.HasVideoExtension(file.DisplayPath(t.Info())) {
				filePaths = append(filePaths, filepath.Join(downloadDir, file.DisplayPath(t.Info())))
			}
		}
	} else {
		file := t.Info().Files[fileIndex]
		filePaths = append(filePaths, filepath.Join(downloadDir, file.DisplayPath(t.Info())))
	}

	for _, filePath := range filePaths {
		err := telegram.SendCommentMessage(filePath, int(messageId))
		if err != nil {
			log.Println("send download comment error", err)
			return
		}
		time.Sleep(2 * time.Second)
	}

	if err := common.RecordDownloadComment(infoHash, fileIndex); err != nil {
		log.Println("record download comment error", err)
		return
	}

	deleteDownloadFile(filePaths)

	err = common.DecrementDailyDownloadQuantity(premium)
	if err != nil {
		log.Println("decrement daily download quantity error", err)
	}
}

func deleteDownloadFile(filePath []string) {
	for _, filePath := range filePath {
		err := os.Remove(filePath)
		if err != nil {
			log.Println("delete download file error", err)
			return
		}
	}
}

func emojifyFilename(filename string) string {
	// 根据文件后缀返回带有 emoji 的文件名
	extToEmoji := map[string]string{
		".mp4":     "🎬",
		".mkv":     "🎥",
		".avi":     "📽️",
		".mov":     "🎞️",
		".ts":      "📼",
		".mp3":     "🎵",
		".flac":    "🎶",
		".wav":     "🔊",
		".ape":     "🎼",
		".aac":     "🎧",
		".ogg":     "🎶",
		".jpg":     "🖼️",
		".jpeg":    "🖼️",
		".png":     "📸",
		".gif":     "🎞️",
		".webp":    "🌆",
		".bmp":     "🖼️",
		".zip":     "🗜️",
		".rar":     "🗂️",
		".7z":      "📦",
		".tar":     "📦",
		".gz":      "🗄️",
		".pdf":     "📑",
		".epub":    "📚",
		".txt":     "📄",
		".doc":     "📝",
		".docx":    "📝",
		".ppt":     "📊",
		".pptx":    "📊",
		".xls":     "📈",
		".xlsx":    "📈",
		".apk":     "🤖",
		".exe":     "🖥️",
		".iso":     "💿",
		".torrent": "🧲",
	}

	ext := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i:]
			break
		}
	}
	emoji := ""
	if val, ok := extToEmoji[ext]; ok {
		emoji = val
	}
	if emoji != "" {
		return emoji
	} else {
		return "📄"
	}
}

func checkOverDownloadSize(infoHash string, fileIndex int, torrentInfo *model.Torrent, permissions *model.Permissions) bool {
	switch fileIndex {
	case -1:
		return torrentInfo.TotalLength() > permissions.FileDownloadSize
	case -2:
		totalLength := int64(0)
		for _, file := range torrentInfo.Files {
			if torrent.HasImageExtension(file.DisplayPath()) {
				totalLength += file.Length
			}
		}
		return totalLength > permissions.FileDownloadSize
	case -3:
		totalLength := int64(0)
		for _, file := range torrentInfo.Files {
			if torrent.HasVideoExtension(file.DisplayPath()) {
				totalLength += file.Length
			}
		}
		return totalLength > permissions.FileDownloadSize
	default:
		return torrentInfo.Files[fileIndex].Length > permissions.FileDownloadSize
	}
}
