package model

type TorrentInfo struct {
	InfoHash    string `gorm:"column:info_hash;type:varchar(255);primaryKey"`
	PieceLength int64  `gorm:"column:piece_length;type:int64"`
	Pieces      []byte `gorm:"column:pieces;type:blob"`
	Name        string `gorm:"column:name;type:varchar(255)"`
	NameUtf8    string `gorm:"column:name_utf8;type:varchar(255)"`
	Length      int64  `gorm:"column:length;type:int64"`
	IsDir       bool   `gorm:"column:is_dir"`
}
