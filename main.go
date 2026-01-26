package main

import (
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// 从环境变量获取 bot token
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("错误: 请在环境变量中设置 BOT_TOKEN")
	}

	// 创建 bot 实例
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("创建 bot 失败:", err)
	}

	// 设置 debug 模式（可选）
	bot.Debug = false

	log.Printf("已授权为 %s", bot.Self.UserName)

	// 创建更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// 获取更新通道
	updates := bot.GetUpdatesChan(u)

	// 处理更新
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message
		chatID := msg.Chat.ID
		text := msg.Text

		// 处理命令
		if msg.IsCommand() {
			switch msg.Command() {
			case "start":
				userName := msg.From.FirstName
				if userName == "" {
					userName = "用户"
				}
				reply := tgbotapi.NewMessage(chatID, "你好，"+userName+"！👋\n\n我是你的 Telegram Bot。\n\n使用 /help 查看可用命令。")
				bot.Send(reply)

			case "help":
				helpText := "可用命令：\n\n" +
					"/start - 开始使用 bot\n" +
					"/help - 显示帮助信息\n" +
					"/echo <消息> - 回显你的消息\n" +
					"/about - 关于这个 bot"
				reply := tgbotapi.NewMessage(chatID, helpText)
				bot.Send(reply)

			case "echo":
				args := strings.TrimSpace(msg.CommandArguments())
				if args == "" {
					reply := tgbotapi.NewMessage(chatID, "请提供要回显的消息，例如: /echo 你好")
					bot.Send(reply)
				} else {
					reply := tgbotapi.NewMessage(chatID, "你说了: "+args)
					bot.Send(reply)
				}

			case "about":
				reply := tgbotapi.NewMessage(chatID, "这是一个基础的 Telegram Bot 示例。\n\n使用 Go 和 go-telegram-bot-api 构建。")
				bot.Send(reply)

			default:
				reply := tgbotapi.NewMessage(chatID, "未知命令。使用 /help 查看可用命令。")
				bot.Send(reply)
			}
			continue
		}

		// 处理普通文本消息
		if text != "" {
			reply := tgbotapi.NewMessage(chatID, "收到你的消息: "+text)
			bot.Send(reply)
		}
	}
}
