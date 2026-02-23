package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"bt-bot/utils"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SelfCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := common.ParseUserId(update)
	chatID := common.ParseMessageChatId(update)
	userName := common.ParseFullName(update)

	// 获取用户信息
	user, err := common.User(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	// 获取用户权限
	permissions, err := common.Permissions(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	// 生成个人消息
	message := i18n.Replace(i18n.Text(i18n.SelfMessageCode, user.Language), map[string]string{
		i18n.SelfMessagePlaceholderUserName:              userName,
		i18n.SelfMessagePlaceholderUUID:                  user.UUID,
		i18n.SelfMessagePlaceholderLanguage:              user.Language,
		i18n.SelfMessagePlaceholderDailyDownloadRemain:   strconv.Itoa(permissions.DailyDownloadQuantity),
		i18n.SelfMessagePlaceholderAsyncDownloadQuantity: strconv.Itoa(permissions.AsyncDownloadQuantity),
		i18n.SelfMessagePlaceholderDailyDownloadQuantity: strconv.Itoa(permissions.DailyDownloadQuantity),
		i18n.SelfMessagePlaceholderFileDownloadSize:      utils.FormatBytesToSizeString(permissions.FileDownloadSize),
	})

	// 创建个人消息
	reply := tgbotapi.NewMessage(chatID, message)

	// 发送个人消息
	if _, err := bot.Send(reply); err != nil {
		log.Println("Send self message error:", err)
	}
}
