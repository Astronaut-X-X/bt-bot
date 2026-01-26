package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构体
type Config struct {
	Bot BotConfig `yaml:"bot"`
}

// BotConfig Bot 配置
type BotConfig struct {
	Token   string `yaml:"token"`
	Debug   bool   `yaml:"debug"`
	Timeout int    `yaml:"timeout"`
	Proxy   string `yaml:"proxy"` // 代理地址，格式: http://host:port 或 socks5://host:port
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证必要配置
	if config.Bot.Token == "" || config.Bot.Token == "your_bot_token_here" {
		return nil, fmt.Errorf("配置错误: bot.token 未设置或使用默认值")
	}

	return &config, nil
}

