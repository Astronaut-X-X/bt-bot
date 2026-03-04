package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/database/model"
	"bt-bot/torrent"
	"bt-bot/utils"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MagnetCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := msg.From.ID

	user, err := common.User(userID)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	// 提取磁力链接
	magnetLink := torrent.ExtractMagnetLink(msg.Text)
	if magnetLink == "" {
		messageText := i18n.Text(i18n.MagnetInvalidLinkMessageCode, user.Language)
		message := i18n.Replace(messageText, map[string]string{
			i18n.MagnetMessagePlaceholderMagnetLink: msg.Text,
		})
		reply := tgbotapi.NewMessage(chatID, message)
		common.SendWithRetry(bot, reply)
		return
	}
	infoHash := torrent.ExtractTorrentInfoHash(magnetLink)

	startTime := time.Now()

	// 发送解析中消息
	processingMessage := i18n.Text(i18n.MagnetProcessingMessageCode, user.Language)
	processingMessage = i18n.Replace(processingMessage, map[string]string{
		i18n.MagnetMessagePlaceholderMagnetLink:  magnetLink,
		i18n.MagnetMessagePlaceholderElapsedTime: "--:--:--",
	})
	processingMsg := tgbotapi.NewMessage(chatID, processingMessage)
	sentMsg, _ := common.SendWithRetry(bot, processingMsg)

	ctx, cancel := context.WithCancel(context.Background())
	torrent.SetTorrentCancel(infoHash, userID, cancel)
	defer torrent.RemoveTorrentCancel(infoHash, userID)

	var info *model.Torrent
	var errParse error
	go func() {
		log.Println("======================= parseMagnetLink ========================", magnetLink)
		info, errParse = parseMagnetLink(ctx, magnetLink)
	}()

	// 解析中
	for {
		elapsedTime := time.Since(startTime)
		hours := int(elapsedTime.Hours())
		minutes := int(elapsedTime.Minutes()) % 60
		seconds := int(elapsedTime.Seconds()) % 60
		elapsedTimeString := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

		processingMessage = i18n.Text(i18n.MagnetProcessingMessageCode, user.Language)
		processingMessage = i18n.Replace(processingMessage, map[string]string{
			i18n.MagnetMessagePlaceholderMagnetLink:  magnetLink,
			i18n.MagnetMessagePlaceholderElapsedTime: elapsedTimeString,
		})
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, processingMessage)
		editMsg.ReplyMarkup = stopMagnetReplyMarkup(infoHash, userID, user.Language)
		if _, err := common.SendWithRetry(bot, editMsg); err != nil {
			log.Println("Send magnet processing message error:", err)
		}

		time.Sleep(1 * time.Second)

		if info != nil || errParse != nil {
			break
		}
	}

	// 解析失败
	if errParse != nil {
		errorMessage := i18n.Text(i18n.MagnetErrorMessageCode, user.Language)
		errorMessage = i18n.Replace(errorMessage, map[string]string{
			i18n.MagnetMessagePlaceholderErrorMessage: errParse.Error(),
			i18n.MagnetMessagePlaceholderMagnetLink:   magnetLink,
			i18n.MagnetMessagePlaceholderTimeout:      strconv.Itoa(int(torrent.MagnetTimeout)),
		})
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorMessage)
		common.SendWithRetry(bot, editMsg)
		return
	}

	// 获取文件列表
	const maxButtons = 48
	files := info.Files
	filesFirstPage := files[:min(maxButtons, len(files))]

	// 发送第一页成功消息
	successMessage := i18n.Text(i18n.MagnetSuccessMessageCode, user.Language)
	successMessage = i18n.Replace(successMessage, map[string]string{
		i18n.MagnetMessagePlaceholderMagnetLink: magnetLink,
		i18n.MagnetMessagePlaceholderFileName:   info.Name,
		i18n.MagnetMessagePlaceholderFileSize:   utils.FormatBytesToSizeString(info.TotalLength()),
		i18n.MagnetMessagePlaceholderFileCount:  strconv.Itoa(len(filesFirstPage)),
		i18n.MagnetMessagePlaceholderFileList:   strings.Join(fileList(filesFirstPage), "\n"),
	})
	editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, successMessage)
	replyMarkup := createFileButtons(filesFirstPage, info.InfoHash)
	replyMarkup.InlineKeyboard = append([][]tgbotapi.InlineKeyboardButton{allFileButton(info.InfoHash)}, replyMarkup.InlineKeyboard...)
	editMsg.ReplyMarkup = replyMarkup
	common.SendWithRetry(bot, editMsg)

	// 发送后续页成功消息
	for i := maxButtons; i < len(files); i += maxButtons {
		filesPage := files[i:min(i+maxButtons, len(files))]
		successMessage = i18n.Text(i18n.MagnetSuccessMessageCode, user.Language)
		successMessage = i18n.Replace(successMessage, map[string]string{
			i18n.MagnetMessagePlaceholderMagnetLink: magnetLink,
			i18n.MagnetMessagePlaceholderFileName:   info.Name,
			i18n.MagnetMessagePlaceholderFileSize:   utils.FormatBytesToSizeString(info.TotalLength()),
			i18n.MagnetMessagePlaceholderFileCount:  strconv.Itoa(len(filesPage)),
			i18n.MagnetMessagePlaceholderFileList:   strings.Join(fileList(filesPage), "\n"),
		})

		message := tgbotapi.NewMessage(chatID, successMessage)
		message.ReplyMarkup = createFileButtons(filesPage, info.InfoHash)
		common.SendWithRetry(bot, message)
	}
}

// parse magnet link to info
func parseMagnetLink(ctx context.Context, magnetLink string) (*model.Torrent, error) {
	var info_ model.Torrent

	infoHash := torrent.ExtractTorrentInfoHash(magnetLink)

	dbInfo, err := common.GetTorrentInfo(infoHash)
	if err != nil {
		log.Println("common.GetTorrentInfo err: ", err)
	}
	if dbInfo != nil {
		info_ = *dbInfo
	} else {
		// 提取 InfoHash
		info, err := torrent.ParseMagnetLink(ctx, magnetLink)
		if err != nil {
			return nil, err
		}
		defer func() {
			if r := recover(); r != nil {
				log.Println("Drop torrent failed: ", r)
			}
			if info != nil {
				info.Drop()
			}
		}()

		parseInfo := info.Info()
		torrentInfo, err := common.SaveTorrentInfo(infoHash, parseInfo)
		if err != nil {
			log.Panicln("common.SaveTorrentInfo err: ", err)
		}
		info_ = *torrentInfo
	}

	return &info_, nil
}

func fileList(files []model.TorrentFile) []string {
	fileList := make([]string, 0)
	for _, file := range files {
		path := file.Path
		if len(file.PathUtf8) > 0 {
			path = file.PathUtf8
		}
		fileLine := fmt.Sprintf("%s %d.%s (%s)", emojifyFilename(path), file.FileIndex+1, path, utils.FormatBytesToSizeString(file.Length))
		fileList = append(fileList, fileLine)
	}
	return fileList
}

func allFileButton(infoHash string) []tgbotapi.InlineKeyboardButton {
	button := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("All files", "file_"+infoHash+"_-1"),
		tgbotapi.NewInlineKeyboardButtonData("All images", "file_"+infoHash+"_-2"),
		tgbotapi.NewInlineKeyboardButtonData("All videos", "file_"+infoHash+"_-3"),
	}
	return button
}

// createFileButtons 创建文件按钮（多按钮同行）
func createFileButtons(files []model.TorrentFile, infoHash string) *tgbotapi.InlineKeyboardMarkup {
	const buttonsPerRow = 8 // 每行显示的按钮数

	var buttons [][]tgbotapi.InlineKeyboardButton

	// 文件按钮，每行多个按钮
	row := []tgbotapi.InlineKeyboardButton{}
	for i := 0; i < len(files); i++ {
		file := files[i]

		buttonText := fmt.Sprintf("%d", file.FileIndex+1)
		callbackData := fmt.Sprintf("file_%s_%d", infoHash, file.FileIndex)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		row = append(row, button)

		// 每buttonsPerRow个按钮一行
		if len(row) == buttonsPerRow {
			buttons = append(buttons, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	// 有剩余按钮未满一行
	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return &keyboard
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

func stopMagnetReplyMarkup(infoHash string, userId int64, language string) *tgbotapi.InlineKeyboardMarkup {
	data := "stop_magnet_" + infoHash + "_" + strconv.FormatInt(userId, 10)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(i18n.Text(i18n.ButtonStopMagnetCode, language), data),
		},
	)

	return &keyboard
}
