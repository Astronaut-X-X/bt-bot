package telegram

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func GetAllSessionUUIDs() []string {
	files, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil
	}

	uuids := make([]string, 0, len(files))
	for _, file := range files {
		uuids = append(uuids, strings.TrimSuffix(file.Name(), ".json"))
	}
	return uuids
}
