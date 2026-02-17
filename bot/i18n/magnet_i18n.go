package i18n

const (
	MagnetInvalidLinkMessageCode       = "magnet_invalid_link_message"
	MagnetMessagePlaceholderMagnetLink = "{magnet_link}"

	MagnetProcessingMessageCode               = "magnet_processing_message"
	MagnetMessagePlaceholderProcessingMessage = ""

	MagnetErrorMessageCode               = "magnet_error_message"
	MagnetMessagePlaceholderErrorMessage = "{error_message}"

	MagnetSuccessMessageCode          = "magnet_success_message"
	MagnetMessagePlaceholderFileName  = "{file_name}"
	MagnetMessagePlaceholderFileSize  = "{file_size}"
	MagnetMessagePlaceholderFileCount = "{file_count}"
	MagnetMessagePlaceholderFileList  = "{file_list}"
)

const (
	MagnetInvalidLinkMessageZH = `
❌ 磁力链接格式错误。
磁力链接：{magnet_link}
请发送磁力链接或使用命令：/magnet <磁力链接>
`
	MagnetInvalidLinkMessageEN = `
❌ No valid magnet link found.
Magnet link: {magnet_link}
Please send a magnet link or use the command: /magnet <magnet link>
`
)

const (
	MagnetProcessingMessageZH = `
⏳ 正在解析磁力链接，请稍候...
`
	MagnetProcessingMessageEN = `
⏳ Parsing magnet link, please wait...
`
)

const (
	MagnetErrorMessageZH = `
❌ 解析失败: {error_message}
可能的原因：
• 网络连接问题
• 磁力链接无效
• 超时（3分钟）
`
	MagnetErrorMessageEN = `
❌ Parsing failed: {error_message}
Possible reasons:
• Network connection problem
• Invalid magnet link
• Timeout (3 minutes)
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
`
	MagnetSuccessMessageEN = `
✅ Parsing successful

🧲 Magnet Link: {magnet_link}
📄 File name: {file_name}
📦 File size: {file_size}
🗃️ File count: {file_count}
📋 File list: 
{file_list}
`
)
