package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HelpCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := common.ParseUserId(update)
	chatID := common.ParseMessageChatId(update)

	// 获取用户信息
	user, err := common.User(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	// 生成帮助消息
	message := i18n.Replace(i18n.Text(i18n.HelpMessageCode, user.Language), map[string]string{
		i18n.HelpMessagePlaceholderDownloadChannel: "@tgqpXOZ2tzXN",
		i18n.HelpMessagePlaceholderHelpChannel:     "@bt1bot1channel",
	})

	// 创建帮助消息
	reply := tgbotapi.NewMessage(chatID, message)

	// 发送帮助消息
	if _, err := bot.Send(reply); err != nil {
		log.Println("Send help message error:", err)
	}
}
