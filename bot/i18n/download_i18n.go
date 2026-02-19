package i18n

const (
	DownloadStartMessageCode      = "download_start_message"
	DownloadProcessingMessageCode = "download_processing_message"
	DownloadSuccessMessageCode    = "download_success_message"
	DownloadFailedMessageCode     = "download_failed_message"

	DownloadMessagePlaceholderMagnet         = "{magnet}"
	DownloadMessagePlaceholderErrorMessage   = "{error_message}"
	DownloadMessagePlaceholderDownloadFiles  = "{download_files}"
	DownloadMessagePlaceholderPercent        = "{percent}"
	DownloadMessagePlaceholderBytesCompleted = "{bytes_completed}"
	DownloadMessagePlaceholderTotalBytes     = "{total_bytes}"
)

const (
	DownloadStartMesssageZH = `
⌛ 准备开始下载文件...
🔗 磁力链接: {magnet}
`

	DownloadStartMesssageEN = `
⌛ Preparing to download file...
🔗 Magnet: {magnet}
`
)

// 文件下载中
const (
	DownloadProcessingMessageZH = `
⌛ 文件下载中...
🔗 Magent: {magnet}
💾 正在下载文件：
[{percent}%({bytes_completed}/{total_bytes})] {download_files}
`

	DownloadProcessingMessageEN = `
⌛ Downloading file...
🔗 Magnet: {magnet}
💾 Downloading:
[{percent}%({bytes_completed}/{total_bytes})] {download_files}
`
)

// 文件下载成功：正常下载完成
const (
	DownloadSuccessMessageZH = `
✅ 文件下载成功
🔗 磁力链接: {magnet}
💾 文件列表：
{download_files}

前往消息频道：{download_channel}
`

	DownloadSuccessMessageEN = `
✅ Download complete
🔗 Magnet: {magnet}
💾 File list:
{download_files}

Go to channel: {download_channel}
`
)

// 文件下载失败（模板含：并发数限制/超时错误/取消下载）
const (
	DownloadFailedMessageZH = `
❎ 下载失败
ℹ 错误信息: {error_message}
🔗 磁力链接: {magnet}
💾 下载文件：
{download_files}

`

	DownloadFailedMessageEN = `
❎ Download failed
ℹ Error: {error_message}
🔗 Magnet: {magnet}
💾 Download file:
{download_files}
`
)
