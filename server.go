package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"bt-bot/handler"
	"bt-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
)

// Server Bot 服务器结构体
type Server struct {
	bot    *tgbotapi.BotAPI
	config *utils.Config
}

// NewServer 创建新的 Server 实例
func NewServer(config *utils.Config) (*Server, error) {
	var bot *tgbotapi.BotAPI
	var err error

	// 如果配置了代理，使用代理创建 bot 实例
	if config.Bot.Proxy != "" {
		bot, err = createBotWithProxy(config.Bot.Token, config.Bot.Proxy)
		if err != nil {
			return nil, err
		}
		log.Printf("使用代理连接: %s", config.Bot.Proxy)
	} else {
		// 不使用代理创建 bot 实例
		bot, err = tgbotapi.NewBotAPI(config.Bot.Token)
		if err != nil {
			return nil, err
		}
	}

	// 设置 debug 模式
	bot.Debug = config.Bot.Debug

	log.Printf("已授权为 %s", bot.Self.UserName)

	// 初始化缓存服务
	if err := handler.InitCache(config); err != nil {
		log.Printf("警告: 初始化缓存服务失败: %v", err)
	} else if config.Cache.Enabled {
		log.Printf("缓存服务已启用，缓存目录: %s", config.Cache.Dir)
	}

	return &Server{
		bot:    bot,
		config: config,
	}, nil
}

// createBotWithProxy 使用代理创建 bot 实例
func createBotWithProxy(token, proxyURL string) (*tgbotapi.BotAPI, error) {
	// 解析代理 URL
	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("解析代理 URL 失败: %w", err)
	}

	var httpClient *http.Client

	// 根据代理类型创建 HTTP 客户端
	switch proxyParsed.Scheme {
	case "http", "https":
		// HTTP/HTTPS 代理，跳过证书验证
		transport := &http.Transport{
			Proxy:           http.ProxyURL(proxyParsed),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	case "socks5":
		// SOCKS5 代理
		dialer, err := proxy.SOCKS5("tcp", proxyParsed.Host, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("创建 SOCKS5 代理失败: %w", err)
		}
		transport := &http.Transport{
			Dial: dialer.Dial,
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	default:
		return nil, fmt.Errorf("不支持的代理类型: %s (支持: http, https, socks5)", proxyParsed.Scheme)
	}

	// 使用自定义 HTTP 客户端创建 bot 实例
	// NewBotAPIWithClient 参数: token, apiEndpoint, client
	// 明确指定 API endpoint 为默认值
	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, fmt.Errorf("创建 bot 实例失败: %w", err)
	}

	return bot, nil
}

// Run 启动服务器并开始处理消息
func (s *Server) Run() error {
	// 创建更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = s.config.Bot.Timeout

	// 获取更新通道
	updates := s.bot.GetUpdatesChan(u)

	log.Println("Bot 服务器已启动，等待消息...")

	// 处理更新
	for update := range updates {
		// 处理回调查询（按钮点击）
		if update.CallbackQuery != nil {
			handler.CallbackQueryHandler(s.bot, update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		msg := update.Message

		// 处理命令
		if msg.IsCommand() {
			switch msg.Command() {
			case "start":
				handler.StartHandler(s.bot, msg)
			case "help":
				handler.HelpHandler(s.bot, msg)
			case "echo":
				handler.EchoHandler(s.bot, msg)
			case "magnet":
				handler.MagnetHandler(s.bot, msg)
			case "about":
				handler.AboutHandler(s.bot, msg)
			case "stop":
				handler.StopHandler(s.bot, msg)
			default:
				handler.UnknownHandler(s.bot, msg)
			}
			continue
		}

		// 处理普通文本消息
		// 检查是否包含磁力链接
		if containsMagnetLink(msg.Text) {
			handler.MagnetHandler(s.bot, msg)
		} else {
			handler.MessageHandler(s.bot, msg)
		}
	}

	return nil
}

// containsMagnetLink 检查文本是否包含磁力链接
func containsMagnetLink(text string) bool {
	return text != "" && strings.Contains(text, "magnet:")
}

// GetBot 获取 bot 实例
func (s *Server) GetBot() *tgbotapi.BotAPI {
	return s.bot
}
