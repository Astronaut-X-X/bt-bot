package model

import "time"

type TorrentFile struct {
	InfoHash         string    `gorm:"column:info_hash;type:varchar(255);primaryKey"`
	Path             string    `gorm:"column:path;type:varchar(255)"`
	Size             int64     `gorm:"column:size;type:int64"`
	Downloaded       bool      `gorm:"column:downloaded;type:bool;default:false"`
	DownloadedAt     time.Time `gorm:"column:downloaded_at;type:datetime"`
	DownloadedSpeed  int64     `gorm:"column:downloaded_speed;type:int64"`
	DownloadedTime   int64     `gorm:"column:downloaded_time;type:int64"`
	DownloadedStatus string    `gorm:"column:downloaded_status;type:varchar(255)"`
}
