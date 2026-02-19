package model

type DownloadFileMessage struct {
	InfoHash  string `gorm:"column:info_hash;"`
	MessageID int64  `gorm:"column:message_id;"`
}
