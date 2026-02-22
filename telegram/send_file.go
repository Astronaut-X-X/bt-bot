package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

const channelUsername = "tgqpXOZ2tzXN"

func SendChannelMessage(text string) (int, error) {
	client := GetIdleGlobalClient()
	if client == nil {
		log.Println("no idle client")
		return 0, errors.New("no idle client")
	}

	channelId, accessHash, err := getInputPeerChannel(client)
	if err != nil {
		log.Println("failed to get channel:", err)
		return 0, err
	}

	// 发送信息
	sendMsg := &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channelId,
			AccessHash: accessHash,
		},
		Message:  text,
		RandomID: rand.Int64(),
	}

	update, err := client.API().MessagesSendMessage(context.TODO(), sendMsg)
	if err != nil {
		log.Println("failed to send message:", err)
		return 0, err
	}
	msgId := 0
	if updates, ok := update.(*tg.Updates); ok {
		for _, update := range updates.Updates {
			if update, ok := update.(*tg.UpdateMessageID); ok {
				msgId = update.ID
				break
			}
		}
	}

	log.Println("send message success", msgId)

	sleepTime := 2 * time.Second
	for {
		time.Sleep(sleepTime)
		msgId, err = getDiscussionMessageId(client, msgId, channelId, accessHash)
		if err == nil {
			log.Println("get discussion message id success", msgId)
			return msgId, nil
		}

		log.Println("failed to get discussion message id:", err)

		if !strings.Contains(err.Error(), "FLOOD_WAIT") {
			log.Println("failed to get discussion message id:", err)
			return 0, err
		}
		// FLOOD_WAIT 错误时指数退避，最大60秒
		if sleepTime < 60*time.Second {
			sleepTime *= 2
			if sleepTime > 60*time.Second {
				sleepTime = 60 * time.Second
			}
		}
		log.Printf("failed to get discussion message id, sleep for %d seconds ", sleepTime/time.Second)
	}
}

func getInputPeerChannel(client *Client) (channelId int64, accessHash int64, err error) {
	// 搜索频道详情
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	response, err := client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetDate: 0,
		OffsetID:   0,
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		log.Println("failed to get dialogs:", err)
		return 0, 0, err
	}

	chats := make([]tg.ChatClass, 0)
	switch response.(type) {
	case *tg.MessagesDialogs:
		dialogs := response.(*tg.MessagesDialogs)
		chats = dialogs.Chats
	case *tg.MessagesDialogsSlice:
		dialogs := response.(*tg.MessagesDialogsSlice)
		chats = dialogs.Chats
	}

	for _, chat := range chats {
		if chat, ok := chat.(*tg.Channel); ok {
			if chat.Username == channelUsername {
				return chat.ID, chat.AccessHash, nil
			}
		}
	}

	return 0, 0, errors.New("channel not found")
}

func getCommonetInputPeerChannel(client *Client, channelId int64, accessHash int64) (tg.InputPeerClass, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := client.API().ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID:  channelId,
		AccessHash: accessHash,
	})
	if err != nil {
		log.Println("failed to get full channel:", err)
		return nil, err
	}

	for _, chat := range response.Chats {
		if channel, ok := chat.(*tg.Channel); ok {
			if channel.Username != channelUsername {
				return &tg.InputPeerChannel{
					ChannelID:  channel.ID,
					AccessHash: channel.AccessHash,
				}, nil
			}
		}
	}

	return nil, errors.New("commonet channel not found")
}

func getDiscussionMessageId(client *Client, msgId int, channelId int64, accessHash int64) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	response, err := client.API().MessagesGetDiscussionMessage(ctx, &tg.MessagesGetDiscussionMessageRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channelId,
			AccessHash: accessHash,
		},
		MsgID: msgId,
	})
	if err != nil {
		log.Println("failed to get MessagesGetDiscussionMessage")
		return 0, err
	}

	for _, message := range response.Messages {
		if message, ok := message.(*tg.Message); ok {
			return message.ID, nil
		}
	}

	return 0, errors.New("discussion message not found")
}

func SendCommentMessage(path string, msgId int) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	client := GetIdleGlobalClient()
	if client == nil {
		log.Println("no idle client")
		return errors.New("no idle client")
	}

	inputFile, err := uploadFile(client, path)
	if err != nil {
		log.Println("failed to upload file:", err)
		return err
	}

	channelId, accessHash, err := getInputPeerChannel(client)
	if err != nil {
		log.Println("failed to get channel:", err)
		return err
	}

	commonetInputPeerChannel, err := getCommonetInputPeerChannel(client, channelId, accessHash)
	if err != nil {
		log.Println("failed to get commonet input peer channel:", err)
		return err
	}

	filename := filepath.Base(path)
	ext := filepath.Ext(path)
	sendMsg := &tg.MessagesSendMediaRequest{
		Peer:     commonetInputPeerChannel,
		RandomID: rand.Int64(),
		ReplyTo: &tg.InputReplyToMessage{
			TopMsgID:     msgId,
			ReplyToMsgID: msgId,
		},
	}

	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".webp":
		sendMsg.Media = &tg.InputMediaUploadedDocument{
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: filename},
				&tg.DocumentAttributeImageSize{},
			},
			File: inputFile,
		}
	case ".mp4", ".mov", ".mkv", ".webm", ".avi":
		sendMsg.Media = &tg.InputMediaUploadedDocument{
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: filename},
				&tg.DocumentAttributeVideo{},
			},
			File: inputFile,
		}
	default:
		sendMsg.Media = &tg.InputMediaUploadedDocument{
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: filename},
			},
			File: inputFile,
		}
	}

	if _, err = client.API().MessagesSendMedia(context.TODO(), sendMsg); err != nil {
		log.Println("failed to send message:", err)
		return err
	}

	return nil
}

func uploadFile(client *Client, path string) (tg.InputFileClass, error) {
	// 读取文件
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 上传文件
	up := uploader.NewUploader(client.API())
	return up.FromFile(context.TODO(), file)
}
