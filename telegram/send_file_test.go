package telegram

import (
	"log"
	"testing"
)

func TestSendFile(t *testing.T) {
	LoadGolbalClient()

	client := GetIdleGlobalClient()
	if client == nil {
		log.Println("no idle client")
		return
	}

	err := SendCommentMessage("45ddbbb4-27e7-4b0c-8bcb-f8245133ed69.jpeg", 20)
	if err != nil {
		log.Println("failed to send comment message:", err)
		return
	}

	log.Println("comment message sent")
}
