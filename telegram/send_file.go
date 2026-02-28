package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Eyevinn/mp4ff/mp4"
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

	sleepTime := 4 * time.Second
	for {
		time.Sleep(sleepTime)
		msgId, err = getDiscussionMessageId(client, msgId, channelId, accessHash)
		if err == nil {
			log.Println("get discussion message id success", msgId)
			return msgId, nil
		}

		if !(strings.Contains(err.Error(), "FLOOD_WAIT") ||
			strings.Contains(err.Error(), "MSG_ID_INVALID")) {
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
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	switch ext {
	case ".mp4":
		width, height, duration, err := parseMp4VideoMetadata(path)
		if err != nil {
			log.Println("failed to parse mp4 video metadata:", err)
			return err
		}
		sendMsg.Media = &tg.InputMediaUploadedDocument{
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: filename},
				&tg.DocumentAttributeVideo{
					SupportsStreaming: true,
					Duration:          float64(duration),
					W:                 width,
					H:                 height,
				},
			},
			File:     inputFile,
			MimeType: mimeType,
		}
	default:
		sendMsg.Media = &tg.InputMediaUploadedDocument{
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: filename},
			},
			File:     inputFile,
			MimeType: mimeType,
		}
	}

	if _, err = client.API().MessagesSendMedia(context.TODO(), sendMsg); err != nil {
		log.Println("failed to send message:", err)
		return err
	}

	return nil
}

func uploadFile(client *Client, path string) (tg.InputFileClass, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 上传文件
	// partSize 设置为 512KB (524288 字节)，这是 Telegram 官方推荐的值
	// 参考: https://core.telegram.org/api/files#uploading-files
	// 使用 512KB 可以避免过多的协议开销，并且能够达到最大文件大小限制
	up := uploader.NewUploader(client.API())
	up.WithPartSize(524288)
	up.WithThreads(5)
	up.WithProgress(&UploadProgress{})
	return up.FromFile(context.TODO(), file)
}

type UploadProgress struct{}

func (p *UploadProgress) Chunk(ctx context.Context, state uploader.ProgressState) error {
	log.Println("upload progress:", state.Uploaded, state.Total)
	return nil
}

func SendCommentMessageText(text string, msgId int) error {
	client := GetIdleGlobalClient()
	if client == nil {
		log.Println("no idle client")
		return errors.New("no idle client")
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

	sendMsg := &tg.MessagesSendMessageRequest{
		Peer:     commonetInputPeerChannel,
		RandomID: rand.Int64(),
		ReplyTo: &tg.InputReplyToMessage{
			TopMsgID:     msgId,
			ReplyToMsgID: msgId,
		},
		Message: text,
	}

	if _, err = client.API().MessagesSendMessage(context.TODO(), sendMsg); err != nil {
		log.Println("failed to send message:", err)
		return err
	}

	return nil
}

func parseMp4VideoMetadata(filePath string) (width, height int, duration int32, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	mp4File, err := mp4.DecodeFile(file)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("decode mp4 failed: %w", err)
	}

	if mvhd := mp4File.Moov.Mvhd; mvhd != nil {
		timescale := mvhd.Timescale
		durationInTimescale := mvhd.Duration
		if timescale > 0 {
			duration = int32(durationInTimescale / uint64(timescale))
		}
	}

	for _, track := range mp4File.Moov.Traks {
		if track.Mdia.Hdlr.HandlerType == "vide" && track.Tkhd != nil {
			width = int(track.Tkhd.Width >> 16)
			height = int(track.Tkhd.Height >> 16)
			return width, height, duration, nil
		}
	}

	return 1920, 1080, 60, fmt.Errorf("no video track found in mp4 file")
}
