package model

type DownloadFileComment struct {
	InfoHash string `gorm:"column:info_hash;"`
	Index    int    `gorm:"column:index;"`
}
