package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/database/model"
	"bt-bot/torrent"
	"bt-bot/utils"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MagnetCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := msg.From.ID

	user, _, err := common.UserAndPermissions(userID)
	if err != nil {
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
		bot.Send(reply)
		return
	}

	// 发送解析中消息
	processingMessage := i18n.Text(i18n.MagnetProcessingMessageCode, user.Language)
	processingMsg := tgbotapi.NewMessage(chatID, processingMessage)
	sentMsg, _ := bot.Send(processingMsg)

	info, err := parseMagnetLink(magnetLink)
	if err != nil {
		errorMessage := i18n.Text(i18n.MagnetErrorMessageCode, user.Language)
		errorMessage = i18n.Replace(errorMessage, map[string]string{
			i18n.MagnetMessagePlaceholderErrorMessage: err.Error(),
		})
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorMessage)
		bot.Send(editMsg)
		return
	}

	successMessage := i18n.Text(i18n.MagnetSuccessMessageCode, user.Language)
	successMessage = i18n.Replace(successMessage, map[string]string{
		i18n.MagnetMessagePlaceholderMagnetLink: magnetLink,
		i18n.MagnetMessagePlaceholderFileName:   info.Name,
		i18n.MagnetMessagePlaceholderFileSize:   utils.FormatBytesToSizeString(info.Length),
		i18n.MagnetMessagePlaceholderFileCount:  strconv.Itoa(len(info.Files)),
		i18n.MagnetMessagePlaceholderFileList:   strings.Join(fileList(info.Files), "\n"),
	})

	editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, successMessage)

	bot.Send(editMsg)
}

// parse magnet link to info
func parseMagnetLink(magnetLink string) (*model.Torrent, error) {
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
		info, err := torrent.ParseMagnetLink(magnetLink)
		if err != nil {
			return nil, err
		}
		// 使用 defer 确保 torrent 在使用完后被清理
		defer func() {
			if info != nil {
				// 安全地调用 Drop，捕获可能的 panic
				defer func() {
					if r := recover(); r != nil {
						// 如果 Drop 失败（torrent 不存在等），忽略 panic
					}
				}()
				info.Drop()
			}
		}()

		parseInfo := info.Info()

		// 存储
		info_.InfoHash = infoHash
		info_.Length = parseInfo.Length
		info_.Pieces = parseInfo.Pieces
		info_.PieceLength = parseInfo.PieceLength
		info_.Name = parseInfo.Name
		info_.NameUtf8 = parseInfo.NameUtf8
		info_.IsDir = parseInfo.IsDir()
		info_.Files = make([]model.TorrentFile, 0, 16)

		for _, file := range parseInfo.Files {
			info_.Files = append(info_.Files, model.TorrentFile{
				InfoHash: infoHash,
				Length:   file.Length,
				Path:     strings.Join(file.Path, "/"),
				PathUtf8: strings.Join(file.PathUtf8, "/"),
			})
		}

		if err := common.SaveTorrentInfo(infoHash, parseInfo); err != nil {
			log.Panicln("common.SaveTorrentInfo err: ", err)
		}
	}

	return &info_, nil
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
func createFileButtons(files []model.TorrentFile, infoHash string) *tgbotapi.InlineKeyboardMarkup {
	log.Println("infoHash", infoHash)

	const maxButtons = 50       // Telegram 限制每个键盘最多 100 个按钮，这里设置 50 个文件按钮
	const maxButtonTextLen = 64 // Telegram 按钮 callback_data 最大 64 字符
	var buttons [][]tgbotapi.InlineKeyboardButton

	// 计算要显示的文件数量
	fileCount := len(files)
	if fileCount > maxButtons {
		fileCount = maxButtons
	}

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
	for i := 0; i < fileCount; i++ {
		emoji := "📄"
		fileName := "File"
		path := files[i].PathUtf8
		if len(path) == 0 {
			path = files[i].Path
		}
		emoji = emojifyFilename(getFileExt(path))
		if len(path) > 0 {
			fileName = getFileExt(path)
		}
		// 按钮文本: 文件名最多40字
		shortName := fileName
		if len([]rune(shortName)) > 40 {
			shortName = string([]rune(shortName)[:37]) + "..."
		}
		buttonText := fmt.Sprintf("%s %d.%s", emoji, i+1, shortName)

		callbackData := fmt.Sprintf("file_%s_%d", infoHash, i)
		// 保证 callback_data 不超过 64
		if len(callbackData) > maxButtonTextLen {
			callbackData = callbackData[:maxButtonTextLen]
		}
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// 如果文件数量超过显示限制，添加"查看更多"提示
	if len(files) > maxButtons {
		infoButton := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📋 共 %d 个文件（仅显示前 %d 个）", len(files), maxButtons),
			fmt.Sprintf("info_%s", infoHash),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{infoButton})
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
		return emoji + " " + filename
	} else {
		return filename
	}
}

func getFileExt(filename string) string {
	ext := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i:]
			break
		}
	}
	return ext
}
