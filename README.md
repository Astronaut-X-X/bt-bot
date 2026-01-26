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

2. 设置环境变量 `BOT_TOKEN`：
   - 在 Telegram 中搜索 `@BotFather`
   - 发送 `/newbot` 创建新 bot
   - 按照提示设置 bot 名称和用户名
   - 复制 BotFather 提供的 token
   - 设置环境变量：
   
   **Linux/macOS:**
   ```bash
   export BOT_TOKEN=你的_bot_token
   ```
   
   **Windows (PowerShell):**
   ```powershell
   $env:BOT_TOKEN="你的_bot_token"
   ```
   
   **Windows (CMD):**
   ```cmd
   set BOT_TOKEN=你的_bot_token
   ```

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
├── main.go         # Bot 主程序
├── go.mod          # Go 模块文件
├── go.sum          # 依赖校验文件（自动生成）
├── .gitignore      # Git 忽略文件
└── README.md       # 项目说明
```

## 技术栈

- Go 1.21+
- github.com/go-telegram-bot-api/telegram-bot-api/v5

