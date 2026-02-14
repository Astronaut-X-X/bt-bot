package bot

import (
	"bt-bot/bot/command"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot 结构体
type Bot struct {
	token   string
	debug   bool
	timeout int

	bot *tgbotapi.BotAPI
}

// NewBot 创建新的 Bot 实例
func NewBot(token string, debug bool) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("创建 bot 实例失败: %w", err)
	}
	bot.Debug = debug
	bot_ := &Bot{
		token:   token,
		debug:   debug,
		timeout: 60,
		bot:     bot,
	}

	return bot_, nil
}

// Run 启动服务器并开始处理消息
func (b *Bot) Run() error {
	// 创建更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.timeout

	// 获取更新通道
	updates := b.bot.GetUpdatesChan(u)

	log.Println("Bot 已启动，等待消息...")

	// 处理更新
	for update := range updates {
		// 处理回调查询（按钮点击）
		if update.CallbackQuery != nil {
			continue
		}

		if update.Message == nil {
			continue
		}

		msg := update.Message

		// 处理命令
		command.CommandHandler(b.bot, &update)

		// 处理普通文本消息
		// 检查是否包含磁力链接
		if containsMagnetLink(msg.Text) {
		} else {
		}
	}

	return nil
}

// containsMagnetLink 检查文本是否包含磁力链接
func containsMagnetLink(text string) bool {
	return text != "" && strings.Contains(text, "magnet:")
}
