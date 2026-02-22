package main

import (
	"bt-bot/database"
	"fmt"
)

func main() {
	database.InitDatabase(database.Config{
		Path:  "database.db",
		Debug: true,
	})

	database.DB.Exec("TRUNCATE TABLE download_file_messages")
	database.DB.Exec("TRUNCATE TABLE download_file_comments")

	fmt.Println("clear download file message and download file comment success")
}
