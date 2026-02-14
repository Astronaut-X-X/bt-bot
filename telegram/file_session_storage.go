package telegram

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram"
)

const sessionDir = "sessions"
const sessionFile = "%s.json"

func init() {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		fmt.Printf("创建会话目录失败: %v\n", err)
		os.Exit(1)
	}
}

func SessionExists(uuid string) bool {
	_, err := os.Stat(GetSessionFile(uuid))
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func GetSessionFile(uuid string) string {
	return filepath.Join(sessionDir, fmt.Sprintf(sessionFile, uuid))
}

func GetSessionStorage(uuid string) *telegram.FileSessionStorage {
	return &telegram.FileSessionStorage{
		Path: GetSessionFile(uuid),
	}
}
