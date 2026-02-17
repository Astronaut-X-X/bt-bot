package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/torrent"
	"bt-bot/utils"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MagnetCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := update.Message
	chatID := msg.Chat.ID

	uuid, ok, err := common.GetUserUUID(msg.From.ID)
	if !ok || err != nil {
		return
	}

	user, _, err := common.GetUserAndPermissions(uuid)
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

	info, err := torrent.ParseMagnetLink(magnetLink)
	if err != nil {
		errorMessage := i18n.Text(i18n.MagnetErrorMessageCode, user.Language)
		errorMessage = i18n.Replace(errorMessage, map[string]string{
			i18n.MagnetMessagePlaceholderErrorMessage: err.Error(),
		})
		editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, errorMessage)
		bot.Send(editMsg)
		return
	}

	// 存储
	info_ := info.Info()

	fileList := make([]string, 0)
	for index, file := range info_.Files {
		path := file.Path
		if len(file.PathUtf8) > 0 {
			path = file.PathUtf8
		}
		fileLine := fmt.Sprintf("• %d.%s (%s)", index, strings.Join(path, "/"), utils.FormatBytesToSizeString(file.Length))
		fileList = append(fileList, fileLine)
	}

	successMessage := i18n.Text(i18n.MagnetSuccessMessageCode, user.Language)
	successMessage = i18n.Replace(successMessage, map[string]string{
		i18n.MagnetMessagePlaceholderMagnetLink: magnetLink,
		i18n.MagnetMessagePlaceholderFileName:   info_.Name,
		i18n.MagnetMessagePlaceholderFileSize:   utils.FormatBytesToSizeString(info_.TotalLength()),
		i18n.MagnetMessagePlaceholderFileCount:  strconv.Itoa(len(info_.Files)),
		i18n.MagnetMessagePlaceholderFileList:   strings.Join(fileList, "\n"),
	})

	editMsg := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, successMessage)

	// 如果有文件，添加文件按钮
	if len(info_.Files) > 0 {
		editMsg.ReplyMarkup = createFileButtons(info_.Files, info.InfoHash().String())
	}

	bot.Send(editMsg)
}

// createFileButtons 创建文件按钮
func createFileButtons(files []metainfo.FileInfo, infoHash string) *tgbotapi.InlineKeyboardMarkup {
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
	buttonText := "📄 All"
	callbackData := fmt.Sprintf("file_%s_%d", infoHash, -1)
	// callback_data 必须小于等于 64 字节
	if len(callbackData) > maxButtonTextLen {
		callbackData = callbackData[:maxButtonTextLen]
	}
	button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})

	// 为每个文件创建按钮
	for i := 0; i < fileCount; i++ {
		fileName := "File"
		if len(files[i].PathUtf8) > 0 {
			// 取文件名最后一部分
			parts := files[i].PathUtf8
			if len(parts) > 0 {
				fileName = parts[len(parts)-1]
			}
		} else if len(files[i].Path) > 0 {
			parts := files[i].Path
			if len(parts) > 0 {
				fileName = parts[len(parts)-1]
			}
		}
		// 按钮文本: 文件名最多40字
		shortName := fileName
		if len([]rune(shortName)) > 40 {
			shortName = string([]rune(shortName)[:37]) + "..."
		}
		buttonText := fmt.Sprintf("📄 %s", shortName)

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
