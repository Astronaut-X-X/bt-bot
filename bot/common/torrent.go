package common

import (
	"bt-bot/database"
	"bt-bot/database/model"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

func SaveTorrentInfo(infoHash string, info *metainfo.Info) error {
	torrentInfo := &model.TorrentInfo{
		InfoHash:    infoHash,
		Name:        info.Name,
		PieceLength: info.PieceLength,
		Pieces:      info.Pieces,
		NameUtf8:    info.NameUtf8,
		Length:      info.Length,
		IsDir:       info.IsDir(),
	}

	if err := database.DB.Save(torrentInfo).Error; err != nil {
		return err
	}

	torrentFiles := make([]model.TorrentFile, 0)
	for _, file := range info.Files {
		torrentFiles = append(torrentFiles, model.TorrentFile{
			InfoHash: infoHash,
			Length:   file.Length,
			Path:     strings.Join(file.Path, "/"),
			PathUtf8: strings.Join(file.PathUtf8, "/"),
		})
	}

	return database.DB.Save(torrentFiles).Error
}

func GetTorrentInfo(infoHash string) (*model.Torrent, error) {
	var torrentInfo model.TorrentInfo
	if err := database.DB.Where("info_hash = ?", infoHash).First(&torrentInfo).Error; err != nil {
		return nil, err
	}

	var torrentFiles []model.TorrentFile
	if err := database.DB.Where("info_hash = ?", infoHash).Find(&torrentFiles).Error; err != nil {
		return nil, err
	}

	return &model.Torrent{
		TorrentInfo: torrentInfo,
		Files:       torrentFiles,
	}, nil
}
