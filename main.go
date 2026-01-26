package main

import (
	"log"

	"bt-bot/utils"
)

func main() {
	// 加载配置文件
	config, err := utils.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 创建服务器实例
	server, err := NewServer(config)
	if err != nil {
		log.Fatal("创建服务器失败:", err)
	}

	// 启动服务器
	if err := server.Run(); err != nil {
		log.Fatal("服务器运行失败:", err)
	}
}
