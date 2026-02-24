package i18n

const (
	MagnetAlreadyParsingMessageCode = "magnet_already_parsing_message"

	MagnetInvalidLinkMessageCode       = "magnet_invalid_link_message"
	MagnetMessagePlaceholderMagnetLink = "{magnet_link}"

	MagnetProcessingMessageCode         = "magnet_processing_message"
	MagnetMessagePlaceholderElapsedTime = "{elapsed_time}"

	MagnetErrorMessageCode               = "magnet_error_message"
	MagnetMessagePlaceholderErrorMessage = "{error_message}"
	MagnetMessagePlaceholderTimeout      = "{timeout}"

	MagnetSuccessMessageCode          = "magnet_success_message"
	MagnetMessagePlaceholderFileName  = "{file_name}"
	MagnetMessagePlaceholderFileSize  = "{file_size}"
	MagnetMessagePlaceholderFileCount = "{file_count}"
	MagnetMessagePlaceholderFileList  = "{file_list}"
)

const (
	MagnetAlreadyParsingMessageZH = "❌ 已经有一个在解析了，请稍后再试"
	MagnetAlreadyParsingMessageEN = "❌ Already parsing, please try again later"
)

const (
	MagnetInvalidLinkMessageZH = `
❌ 磁力链接格式错误。

🧲 磁力链接：{magnet_link}
请发送磁力链接或使用命令：/magnet <磁力链接>
`
	MagnetInvalidLinkMessageEN = `
❌ No valid magnet link found.

🧲 Magnet link: {magnet_link}
Please send a magnet link or use the command: /magnet <magnet link>
`
)

const (
	MagnetProcessingMessageZH = `
⏳ 正在解析磁力链接，请稍候...

🧲 磁力链接：{magnet_link}
⏱️ 当前耗时：{elapsed_time}
`
	MagnetProcessingMessageEN = `
⏳ Parsing magnet link, please wait...

🧲 Magnet link: {magnet_link}
⏱️ Current elapsed time: {elapsed_time}

🧲 Magnet link: {magnet_link}
Current elapsed time: {elapsed_time}
`
)

const (
	MagnetErrorMessageZH = `
❌ 解析失败: 

⚠️ 错误信息: {error_message}
🧲 磁力链接：{magnet_link}

⚠️ 可能原因：
• 网络连接问题
• 磁力链接无效
• 超时（{timeout}分钟）
`
	MagnetErrorMessageEN = `
❌ Parsing failed: 

⚠️ Error: {error_message}
🧲 Magnet link: {magnet_link}

⚠️ Possible reasons:
• Network connection problem
• Invalid magnet link
• Timeout ({timeout} minutes)
`
)

const (
	MagnetSuccessMessageZH = `
✅ 解析成功

🧲 磁力链接：{magnet_link}
📄 文件名：{file_name}
📦 文件大小：{file_size}
🗃️ 文件数量：{file_count}
📋 文件列表：

{file_list}

📥 选择文件下载：
`
	MagnetSuccessMessageEN = `
✅ Parsing successful

🧲 Magnet Link: {magnet_link}
📄 File name: {file_name}
📦 File size: {file_size}
🗃️ File count: {file_count}
📋 File list: 

{file_list}

📥 Select file to download:
`
)
