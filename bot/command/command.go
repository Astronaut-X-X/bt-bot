package command

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
		case CommandMagnet:
		case CommandSelf:
		}
	}
}
