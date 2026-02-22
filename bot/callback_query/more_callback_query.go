package callback_query

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/database/model"
	"bt-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MoreCallbackQuery(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// // 校验下载限制
	user, _, err := common.UserAndPermissions(update.CallbackQuery.From.ID)
	if err != nil {
		log.Println("get user and permissions error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ get user and permissions error"))
		return
	}

	data := update.CallbackQuery.Data
	infoHash, page, err := parseMoreCallbackQueryData(data)
	if err != nil {
		log.Println("parse more callback query data error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ parse more callback query data error"))
		return
	}

	torrentInfo, err := common.GetTorrentInfo(infoHash)
	if err != nil {
		log.Println("get torrent error", err)
		bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ get torrent error"))
		return
	}

	successMessage := i18n.Text(i18n.MagnetSuccessMessageCode, user.Language)
	successMessage = i18n.Replace(successMessage, map[string]string{
		i18n.MagnetMessagePlaceholderMagnetLink: fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash),
		i18n.MagnetMessagePlaceholderFileName:   torrentInfo.Name,
		i18n.MagnetMessagePlaceholderFileSize:   utils.FormatBytesToSizeString(torrentInfo.TotalLength()),
		i18n.MagnetMessagePlaceholderFileCount:  strconv.Itoa(len(torrentInfo.Files)),
		i18n.MagnetMessagePlaceholderFileList:   strings.Join(fileList(torrentInfo.Files), "\n"),
	})

	editMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, successMessage)
	editMsg.ReplyMarkup = createFileButtons(torrentInfo.Files, infoHash, page)
	bot.Send(editMsg)
}

func parseMoreCallbackQueryData(data string) (string, int, error) {
	parts := strings.Split(data, "_")
	if len(parts) != 4 {
		return "", 0, errors.New("invalid data")
	}
	if parts[0]+"_"+parts[1] != "info_more" {
		return "", 0, errors.New("invalid data")
	}
	infoHash := parts[2]
	page, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", 0, err
	}
	return infoHash, page, nil
}

func fileList(files []model.TorrentFile) []string {
	fileList := make([]string, 0)
	for index, file := range files {
		path := file.Path
		if len(file.PathUtf8) > 0 {
			path = file.PathUtf8
		}
		fileLine := fmt.Sprintf("• %d.%s (%s)", index+1, path, utils.FormatBytesToSizeString(file.Length))
		fileList = append(fileList, fileLine)
	}
	return fileList
}

// createFileButtons 创建文件按钮
func createFileButtons(files []model.TorrentFile, infoHash string, page int) *tgbotapi.InlineKeyboardMarkup {
	log.Println("infoHash", infoHash)

	const maxButtons = 50       // Telegram 限制每个键盘最多 100 个按钮，这里设置 50 个文件按钮
	const maxButtonTextLen = 64 // Telegram 按钮 callback_data 最大 64 字符
	var buttons [][]tgbotapi.InlineKeyboardButton

	// 添加所有文件按钮（全体下载，index = -1）
	buttonText := "📄 All Files"
	callbackData := fmt.Sprintf("file_%s_%d", infoHash, -1)
	// callback_data 必须小于等于 64 字节
	if len(callbackData) > maxButtonTextLen {
		callbackData = callbackData[:maxButtonTextLen]
	}
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})

	// 为每个文件创建按钮
	for i := (page - 1) * maxButtons; i < min(page*maxButtons, len(files)); i++ {
		file := files[i]
		path := file.PathUtf8
		if len(path) == 0 {
			path = files[i].Path
		}

		fileName := path
		emoji := emojifyFilename(fileName)

		buttonText := fmt.Sprintf("%s %d.%s", emoji, file.FileIndex+1, fileName)
		callbackData := fmt.Sprintf("file_%s_%d", infoHash, file.FileIndex)
		if len(callbackData) > maxButtonTextLen {
			callbackData = callbackData[:maxButtonTextLen]
		}
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// 如果文件数量超过显示限制，添加"查看更多"提示
	if len(files) > maxButtons {
		if page > 1 {
			infoButton := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📋 前一页 <"),
				fmt.Sprintf("info_more_%s_%d", infoHash, page-1),
			)
			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{infoButton})
		}
		if page*maxButtons < len(files) {
			infoButton := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📋 后一页 >"),
				fmt.Sprintf("info_more_%s_%d", infoHash, page+1),
			)
			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{infoButton})
		}
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
