package telegram

import "os"

type SessionFile struct {
	Session string `json:"session"`
}

func NewSessionFile(session string) *SessionFile {
	return &SessionFile{Session: session}
}

func (s *SessionFile) Save(filename string) error {
	return os.WriteFile(filename, []byte(s.Session), 0644)
}
