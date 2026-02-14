package telegram

import "sync"

var _globalClientMutex sync.Map

func LoadGolbalClient() {
	uuids := GetAllSessionUUIDs()
	for _, uuid := range uuids {
		client := NewClient(uuid)
		if client != nil {
			_globalClientMutex.Store(uuid, client)
		}
	}
}

func StopAllGlobalClient() {
	_globalClientMutex.Range(func(key, value any) bool {
		client := value.(*Client)
		client.Stop()
		return true
	})

	_globalClientMutex.Clear()
}

func GetIdleGlobalClient() *Client {
	var client *Client = nil
	_globalClientMutex.Range(func(key, value any) bool {
		if client == nil && value.(*Client).IsIdle() {
			client = value.(*Client)
			return false
		}
		return true
	})

	return client
}
