package main

import (
	"log"

	"bt-bot/bot"
	"bt-bot/database"
	"bt-bot/utils"
)

func main() {
	// 加载配置文件
	config, err := utils.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	database.InitDatabase(database.Config{
		Path:  "database.db",
		Debug: false,
	})

	bot, err := bot.NewBot(config.Bot.Token, config.Bot.Debug)
	if err != nil {
		log.Fatal("创建 bot 失败:", err)
	}

	bot.Run()
}
