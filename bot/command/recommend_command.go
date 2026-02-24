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

	groupChannel := `
@cili8888 - 磁力链接精选福利集
@javday - AV日报-种子|磁链|下载链接|日本|有码|无码|骑兵|步兵
@dianying4K - 4K影视屋(分屋）-蓝光无损电影
@jp_ziyuan - 🇯🇵pikpak日本AV无码 [磁力|磁链|Bt种子]
@new2048cc - 2048核基地磁力|每日更新
@rrclck - 磁力仓库
@AV688 - AV收藏|优质精选|无码破解|中文字幕|番号磁力大全
@TheMissesX - The MissesX🧲磁力链接福利
@gifdaquan - 📖 GIF出處大全
	`

	message := i18n.Replace(i18n.Text(i18n.RecommendMessageCode, user.Language), map[string]string{
		i18n.RecommendMessagePlaceholderGroupChannel: groupChannel,
	})

	reply := tgbotapi.NewMessage(chatID, message)

	if _, err := bot.Send(reply); err != nil {
		log.Println("Send start message error:", err)
	}

}
