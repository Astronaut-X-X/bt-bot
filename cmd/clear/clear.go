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

	database.DB.Model(&model.DownloadFileMessage{}).Where("message_id = ?", "*").Delete(&model.DownloadFileMessage{})
	database.DB.Model(&model.DownloadFileComment{}).Where("file_index = ?", "*").Delete(&model.DownloadFileComment{})

	fmt.Println("clear download file message and download file comment success")
}
