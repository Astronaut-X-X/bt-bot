package model

type Torrent struct {
	PieceLength int64  `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length,omitempty"`
	Private     *bool  `bencode:"private,omitempty"`
	Source      string `bencode:"source,omitempty"`
}
