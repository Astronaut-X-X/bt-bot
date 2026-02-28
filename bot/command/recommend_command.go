package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RecommendCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

	chatID := common.ParseMessageChatId(update)
	userId := common.ParseUserId(update)
	user, err := common.User(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	message := i18n.Replace(i18n.Text(i18n.RecommendMessageCode, user.Language), map[string]string{
		i18n.RecommendMessagePlaceholderGroupChannel:  GroupChannel(),
		i18n.RecommendMessagePlaceholderSearchWebsite: SearchWebsite(),
	})

	reply := tgbotapi.NewMessage(chatID, message)

	if _, err := bot.Send(reply); err != nil {
		log.Println("Send start message error:", err)
	}
}
