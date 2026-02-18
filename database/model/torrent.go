package model

type Torrent struct {
	TorrentInfo
	Files []TorrentFile
}

func (info *Torrent) TotalLength() (ret int64) {
	if info.IsDir {
		for _, fi := range info.Files {
			ret += fi.Length
		}
	} else {
		ret = info.Length
	}
	return
}
