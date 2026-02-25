package i18n

const (
	StartMessageCode = "start_message"

	StartMessagePlaceholderUserName           = "{bot_user_name}"
	StartMessagePlaceholderDownloadChannel    = "{download_channel}"
	StartMessagePlaceholderHelpChannel        = "{help_channel}"
	StartMessagePlaceholderCooperationContact = "{cooperation_contact}"
)

const (
	StartMessageZH = `
Hi,  {bot_user_name}
欢迎使用 BtBot 🤖 

🔍 功能介绍：
- 解析 magnet 链接
- 下载出的解析文件

⌨️ 使用方式：
直接发送 magent 即可开始解析
如：magnet:?xt=urn:btih:E7FC73D9E20697C6C440203F5884EF52F9E4BD28

免责声明：
- 只提供解析下载功能，下载内容与本Bot无关
- 不存储内容，只提供下载，请自行判断内容真实性与合规性
- 违规内容请在帮助反馈频道反馈，我们会及时处理

Bot频道：
下载文件频道：{download_channel}
帮助反馈频道：{help_channel}	

合作联系：{cooperation_contact}
`

	StartMessageEN = `
Hi,  {bot_user_name}   

Welcome to BtBot 🤖 

🔍 Function introduction:
- Parse magnet links
- Download parsed files

⌨️ Usage:
Send magnet to start parsing
如：magnet:?xt=urn:btih:E7FC73D9E20697C6C440203F5884EF52F9E4BD28

Disclaimer:
- Only provide parsing and download functionality, the content of the downloaded content is not related to this Bot
- Do not store content, only provide download, please judge the authenticity and legality of the content yourself
- If you find any illegal content, please feedback in the help feedback channel, we will handle it in time

Bot channel:
Download file channel: {download_channel}
Help feedback channel: {help_channel}	

Cooperation contact: {cooperation_contact}
`
)
