package i18n

const (
	DownloadAlreadyDownloadingMessageCode          = "download_already_downloading_message"
	DownloadDailyDownloadCountNotEnoughMessageCode = "download_daily_download_count_not_enough_message"
	DownloadFileDownloadSizeNotEnoughMessageCode   = "download_file_download_size_not_enough_message"

	DownloadStartMessageCode      = "download_start_message"
	DownloadProcessingMessageCode = "download_processing_message"
	DownloadSendFileMessageCode   = "download_send_file_message"
	DownloadSuccessMessageCode    = "download_success_message"
	DownloadFailedMessageCode     = "download_failed_message"

	DownloadMessagePlaceholderMagnet          = "{magnet}"
	DownloadMessagePlaceholderErrorMessage    = "{error_message}"
	DownloadMessagePlaceholderDownloadFiles   = "{download_files}"
	DownloadMessagePlaceholderPercent         = "{percent}"
	DownloadMessagePlaceholderBytesCompleted  = "{bytes_completed}"
	DownloadMessagePlaceholderTotalBytes      = "{total_bytes}"
	DownloadMessagePlaceholderDownloadChannel = "{download_channel}"
	DownloadMessagePlaceholderElapsedTime     = "{elapsed_time}"
)

const (
	DownloadAlreadyDownloadingMessageZH = "❌ 已经有一个在下载了，请稍后再试"
	DownloadAlreadyDownloadingMessageEN = "❌ Already downloading, please try again later"
)

const (
	DownloadDailyDownloadCountNotEnoughMessageZH = "❌ 每日下载数量不足，请明天再试"
	DownloadDailyDownloadCountNotEnoughMessageEN = "❌ Daily download count not enough, please try again tomorrow"
)

const (
	DownloadFileDownloadSizeNotEnoughMessageZH = "❌ 文件下载大小超过限制"
	DownloadFileDownloadSizeNotEnoughMessageEN = "❌ File download size exceeds the limit"
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

⚠️ 若资源过冷门，可能会等待较长时间或无法完成下载。

🔗 磁力链接: {magnet}
⏱️ 当前耗时: {elapsed_time}
💾 正在下载文件：
[{percent}({bytes_completed}/{total_bytes})] {download_files}
`

	DownloadProcessingMessageEN = `
⌛ Downloading file...

⚠️ If the resource is unpopular, it may take a long time or cannot be completed.

🔗 Magnet: {magnet}
⏱️ Elapsed time: {elapsed_time}
💾 Downloading:
[{percent}({bytes_completed}/{total_bytes})] {download_files}
`
)

const (
	DownloadSendFileMessageZH = `
⌛ 文件发送中...

🔗 Magent: {magnet}
💾 正在发送文件：
{download_files}
`

	DownloadSendFileMessageEN = `
⌛ Sending file...

🔗 Magnet: {magnet}
💾 Sending file:
{download_files}
`
)

// 文件下载成功：正常下载完成
const (
	DownloadSuccessMessageZH = `
✅ 文件下载成功

🔗 磁力链接: #{magnet}
💾 文件列表：
{download_files}

前往消息频道：{download_channel}
`

	DownloadSuccessMessageEN = `
✅ Download complete

🔗 Magnet: #{magnet}
💾 File list:
{download_files}

Go to channel: {download_channel}
`
)

// 文件下载失败（模板含：并发数限制/超时错误/取消下载）
const (
	DownloadFailedMessageZH = `
❌ 下载失败

⚠️ 错误信息: {error_message}
🔗 磁力链接: {magnet}
💾 下载文件：
{download_files}

`

	DownloadFailedMessageEN = `
❌ Download failed

⚠️ Error: {error_message}
🔗 Magnet: {magnet}
💾 Download file:
{download_files}
`
)
