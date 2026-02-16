package i18n

const (
	HelpMessagePlaceholderDownloadChannel = "{download_channel}"
	HelpMessagePlaceholderHelpChannel     = "{help_channel}"
)

const (
	HelpMessageZH = `
可用命令：
/start - 开始使用 bot
/magnet <磁力链接> - 解析磁力链接信息
/self - 个人消息
/help - 显示帮助信息

💡 提示：直接发送磁力链接也可以自动解析

Bot频道：
下载文件频道：{download_channel}
帮助反馈频道：{help_channel}		
`

	HelpMessageEN = `
Available commands:
/start - Start using bot
/magnet <magnet link> - Parse magnet link information
/self - Personal message
/help - Display help information

💡 Tip: Directly sending a magnet link can also automatically parse

Bot channel:
Download file channel: {download_channel}
Help feedback channel: {help_channel}		
`
)
