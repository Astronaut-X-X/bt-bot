package i18n

const (
	SelfMessagePlaceholderUserName              = "{bot_user_name}"
	SelfMessagePlaceholderUUID                  = "{uuid}"
	SelfMessagePlaceholderLanguage              = "{language}"
	SelfMessagePlaceholderDailyDownloadRemain   = "{daily_download_remain}"
	SelfMessagePlaceholderAsyncDownloadQuantity = "{async_download_quantity}"
	SelfMessagePlaceholderDailyDownloadQuantity = "{daily_download_quantity}"
	SelfMessagePlaceholderFileDownloadSize      = "{file_download_size}"
)

const (
	SelfMessageZH = `
你好，{bot_user_name}！👋

个人消息：
唯一标识: {uuid} [请保管好，不要泄露给他人]
使用语言: {language}

使用限制：
- 剩余每日下载数量：{daily_download_remain}

权限信息：
- 并发下载数量：{async_download_quantity}
- 每日下载数量：{daily_download_quantity}
- 下载文件大小限制：{file_download_size}
`

	SelfMessageEN = `
Hello, {bot_user_name}! 👋

Personal message:
Unique identifier: {uuid} [Please keep it safe, do not leak to others]
Using language: {language}

Usage limit:
- Remaining daily download quantity: {daily_download_remain}

Permission information:
- Concurrent download quantity: {async_download_quantity}
- Daily download quantity: {daily_download_quantity}
- Download file size limit: {file_download_size}
`
)
