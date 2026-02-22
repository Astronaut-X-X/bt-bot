package main

import (
	"bt-bot/database"
	"bt-bot/database/model"
	"fmt"
)

func main() {
	database.InitDatabase(database.Config{
		Path:  "database.db",
		Debug: true,
	})

	database.DB.Model(&model.DownloadFileMessage{}).Delete(&model.DownloadFileMessage{})
	database.DB.Model(&model.DownloadFileComment{}).Delete(&model.DownloadFileComment{})

	fmt.Println("clear download file message and download file comment success")
}
