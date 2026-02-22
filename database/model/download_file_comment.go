package model

type DownloadFileComment struct {
	InfoHash  string `gorm:"column:info_hash;"`
	FileIndex int    `gorm:"column:file_index;"`
}
