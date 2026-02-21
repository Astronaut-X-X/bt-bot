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

	err := SendCommentMessage("movie.mp4", 20)
	if err != nil {
		log.Println("failed to send comment message:", err)
		return
	}

	log.Println("comment message sent")
}
