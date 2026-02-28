package command

import (
	"bt-bot/bot/common"
	"bt-bot/bot/i18n"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := common.ParseUserId(update)
	chatID := common.ParseMessageChatId(update)
	userName := common.ParseFullName(update)

	user, err := common.User(userId)
	if err != nil {
		common.SendErrorMessage(bot, chatID, user.Language, err)
		return
	}

	message := i18n.Replace(i18n.Text(i18n.StartMessageCode, user.Language), map[string]string{
		i18n.StartMessagePlaceholderUserName:           userName,
		i18n.StartMessagePlaceholderDownloadChannel:    "@tgqpXOZ2tzXN",
		i18n.StartMessagePlaceholderHelpChannel:        "@bt1bot1channel",
		i18n.StartMessagePlaceholderCooperationContact: "@IIAlbertEinsteinII",
		i18n.StartMessagePlaceholderGroupChannel:       GroupChannel(),
		i18n.StartMessagePlaceholderSearchWebsite:      SearchWebsite(),
	})

	reply := tgbotapi.NewMessage(chatID, message)
	reply.ReplyMarkup = startReplyMarkup()

	if _, err := bot.Send(reply); err != nil {
		log.Println("Send start message error:", err)
	}
}

func startReplyMarkup() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("🇨🇳中文", "lang_zh")},
			{tgbotapi.NewInlineKeyboardButtonData("🇺🇸English", "lang_en")},
		},
	}
}

func GroupChannel() string {
	return `
@cili8888 - 磁力链接精选福利集
@javday - AV日报-种子|磁链|下载链接|日本|有码|无码|骑兵|步兵
@jp_ziyuan - 🇯🇵pikpak日本AV无码 [磁力|磁链|Bt种子]
@new2048cc - 2048核基地磁力|每日更新
@rrclck - 磁力仓库
@AV688 - AV收藏|优质精选|无码破解|中文字幕|番号磁力大全
@TheMissesX - The MissesX🧲磁力链接福利
@gifdaquan - 📖 GIF出處大全
	`
}

func SearchWebsite() string {
	return `
https://mmnnmmnn.mnmnmnmnmn.com/
https://u3c3u3c3.u3c3u3c3u3c3.com/
https://skrbtso.top/
https://btdig.com/
	`
}
