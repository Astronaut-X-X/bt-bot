# Telegram Bot

一个基础的 Telegram Bot 项目，使用 Go 语言开发。

## 功能

- `/start` - 开始使用 bot
- `/help` - 显示帮助信息
- `/echo <消息>` - 回显你的消息
- `/about` - 关于这个 bot
- 普通消息回复

## 前置要求

- Go 1.21 或更高版本

## 安装

1. 克隆或下载项目后，安装依赖：
```bash
go mod download
```

2. 配置 Bot Token：
   - 在 Telegram 中搜索 `@BotFather`
   - 发送 `/newbot` 创建新 bot
   - 按照提示设置 bot 名称和用户名
   - 复制 BotFather 提供的 token
   - 创建配置文件：
   ```bash
   cp config.yaml.example config.yaml
   ```
   - 编辑 `config.yaml` 文件，将 `your_bot_token_here` 替换为你的 bot token

## 运行

```bash
go run main.go
```

或者先编译再运行：
```bash
go build -o bt-bot
./bt-bot
```

## 项目结构

```
bt-bot/
├── main.go              # Bot 主程序入口
├── server.go            # Bot 服务器结构体
├── config.yaml          # 配置文件（需要自行创建）
├── config.yaml.example  # 配置文件示例
├── go.mod               # Go 模块文件
├── go.sum               # 依赖校验文件（自动生成）
├── handler/             # 消息处理模块
│   ├── startHandler.go  # /start 命令处理
│   ├── helpHandler.go   # /help 命令处理
│   ├── echoHandler.go   # /echo 命令处理
│   ├── aboutHandler.go  # /about 命令处理
│   ├── messageHandler.go # 普通消息处理
│   └── unknownHandler.go # 未知命令处理
├── service/             # 服务模块
│   ├── torrent.go       # 磁力链接解析服务
│   └── torrent_test.go  # 磁力链接解析测试
├── utils/               # 工具包
│   └── config.go        # 配置加载模块
├── .gitignore           # Git 忽略文件
└── README.md            # 项目说明
```

## 配置文件说明

配置文件 `config.yaml` 包含以下配置项：

- `bot.token`: Bot Token（从 @BotFather 获取）
- `bot.debug`: 是否启用调试模式（true/false）
- `bot.timeout`: 更新超时时间（秒）
- `bot.proxy`: 代理地址（可选，用于解决网络连接问题）
  - HTTP/HTTPS 代理格式: `http://127.0.0.1:7890`
  - SOCKS5 代理格式: `socks5://127.0.0.1:1080`

### 配置代理

如果遇到网络连接超时问题（如 `i/o timeout`），可以在 `config.yaml` 中配置代理：

```yaml
bot:
  token: "your_bot_token"
  debug: false
  timeout: 60
  proxy: "http://127.0.0.1:7890"  # 根据你的代理设置修改
```

常见的代理配置：
- Clash: `http://127.0.0.1:7890` 或 `socks5://127.0.0.1:7891`
- V2Ray: `socks5://127.0.0.1:1080`
- 其他代理工具请查看其配置的端口

## 技术栈

- Go 1.21+
- github.com/go-telegram-bot-api/telegram-bot-api/v5
- github.com/anacrolix/torrent (用于磁力链接解析)
- gopkg.in/yaml.v3
- golang.org/x/net (用于代理支持)

## 功能模块

### 磁力链接解析服务

`service/torrent.go` 提供了磁力链接解析功能：

- `ParseMagnetLink(magnetLink string)`: 解析磁力链接，返回文件信息、大小、tracker 等
- `ParseTorrentFile(torrentPath string)`: 解析 torrent 文件

使用示例：
```go
service, err := service.NewTorrentService()
if err != nil {
    log.Fatal(err)
}
defer service.Close()

info, err := service.ParseMagnetLink("magnet:?xt=urn:btih:...")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("名称: %s\n", info.Name)
fmt.Printf("总大小: %d 字节\n", info.TotalLength)
fmt.Printf("文件数: %d\n", len(info.Files))
```

运行测试：
```bash
go test ./service -v
```

## 常见问题

### 网络连接超时

如果遇到 `dial tcp ... i/o timeout` 错误，说明无法连接到 Telegram API。解决方法：

1. **配置代理**：在 `config.yaml` 中添加 `proxy` 配置项
2. **检查网络**：确认你的网络可以访问 Telegram API
3. **安装依赖**：运行 `go mod tidy` 确保所有依赖已下载

