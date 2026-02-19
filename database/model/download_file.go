package model

type DownloadFile struct {
	MessageID int64  `gorm:"column:message_id;"`
	InfoHash  string `gorm:"column:info_hash;"`
	Index     int    `gorm:"column:index;"`
	Path      string `gorm:"column:path;"`
}
