package common

import (
	"bt-bot/database"
	"bt-bot/database/model"
	"errors"

	"gorm.io/gorm"
)

func CheckDownloadMessage(infoHash string) (int64, bool, error) {
	var downloadMessage model.DownloadFileMessage
	err := database.DB.Where("info_hash = ?", infoHash).First(&downloadMessage).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return -1, false, err
	} else if err != nil {
		return 0, false, err
	}

	return downloadMessage.MessageID, true, nil
}

func RecordDownloadMessage(infoHash string, messageID int64) error {
	downloadMessage := model.DownloadFileMessage{
		InfoHash:  infoHash,
		MessageID: messageID,
	}
	return database.DB.Create(&downloadMessage).Error
}

func CheckDownloadComment(infoHash string, index int) (bool, error) {
	var downloadComment model.DownloadFileComment
	err := database.DB.Where("info_hash = ? AND file_index = ?", infoHash, index).First(&downloadComment).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func RecordDownloadComment(infoHash string, index int) error {
	downloadComment := model.DownloadFileComment{
		InfoHash:  infoHash,
		FileIndex: index,
	}
	return database.DB.Save(&downloadComment).Error
}
