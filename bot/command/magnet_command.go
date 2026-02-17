package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/torrent"
	"bt-bot/utils"
	"fmt"
	"strconv"
	"strings"

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
	for _, file := range info_.Files {
		path := file.Path
		if len(file.PathUtf8) > 0 {
			path = file.PathUtf8
		}
		fileLine := fmt.Sprintf("%s (%s)", strings.Join(path, "/"), utils.FormatBytesToSizeString(file.Length))
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

	// // 如果有文件，添加文件按钮
	// if len(info.Files) > 0 {
	// 	editMsg.ReplyMarkup = createFileButtons(info.Files, info.InfoHash)
	// }

	bot.Send(editMsg)
}
