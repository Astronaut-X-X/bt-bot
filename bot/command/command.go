package command

import (
	middleware "bt-bot/bot/middle_ware"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CommandStart  = "start"
	CommandHelp   = "help"
	CommandMagnet = "magnet"
	CommandSelf   = "self"
)

func CommandHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

	msg := update.Message

	if msg == nil {
		return
	}

	if msg.IsCommand() {
		switch msg.Command() {
		case CommandStart:
			StartCommand(bot, update)
		case CommandHelp:
			HelpCommand(bot, update)
		case CommandSelf:
			SelfCommand(bot, update)
		case CommandMagnet:
			middleware.MagnetMiddleWare(MagnetCommand)(bot, update)
		}
	}
}
