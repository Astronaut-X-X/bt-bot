package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HelpHandler 处理 /help 命令
func HelpHandler(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	helpText := "可用命令：\n\n" +
		"/start - 开始使用 bot\n" +
		"/help - 显示帮助信息\n" +
		"/echo <消息> - 回显你的消息\n" +
		"/magnet <磁力链接> - 解析磁力链接信息\n" +
		"/stop - 停止当前下载任务\n" +
		"/about - 关于这个 bot\n\n" +
		"💡 提示：直接发送磁力链接也可以自动解析"

	reply := tgbotapi.NewMessage(chatID, helpText)
	bot.Send(reply)
}

