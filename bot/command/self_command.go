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
	msg := update.Message

	chatID := msg.Chat.ID
	userName := common.FullName(msg.From)

	UUID, ok, err := common.GetUserUUID(msg.From.ID)
	if !ok || err != nil {
		log.Println("GetUserUUID error:", err)
		return
	}

	user, permissions, err := common.GetUserAndPermissions(UUID)
	if err != nil {
		log.Println("GetUserAndPermissions error:", err)
		return
	}

	message := i18n.Replace(i18n.Text("self_message", user.Language), map[string]string{
		i18n.SelfMessagePlaceholderUserName:              userName,
		i18n.SelfMessagePlaceholderUUID:                  user.UUID,
		i18n.SelfMessagePlaceholderLanguage:              user.Language,
		i18n.SelfMessagePlaceholderDailyDownloadRemain:   strconv.Itoa(permissions.DailyDownloadQuantity),
		i18n.SelfMessagePlaceholderAsyncDownloadQuantity: strconv.Itoa(permissions.AsyncDownloadQuantity),
		i18n.SelfMessagePlaceholderDailyDownloadQuantity: strconv.Itoa(permissions.DailyDownloadQuantity),
		i18n.SelfMessagePlaceholderFileDownloadSize:      utils.FormatBytesToSizeString(permissions.FileDownloadSize),
	})

	reply := tgbotapi.NewMessage(chatID, message)
	bot.Send(reply)
}
