package model

type TorrentFile struct {
	InfoHash  string `gorm:"column:info_hash;"`
	FileIndex int    `gorm:"column:file_index;"`
	Length    int64  `gorm:"colume:length;"`
	Path      string `gorm:"column:path;"`
	PathUtf8  string `gorm:"column:path_utf8;"`
}
