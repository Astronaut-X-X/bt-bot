package i18n

const (
	ErrorCommonMessageCode = "error_common_message"

	ErrorMessagePlaceholderErrorMessage = "{error_message}"
)

const (
	ErrorCommonMessageZH = `
❌ Download failed

⚠️ 错误信息:
{error_message}
`
	ErrorCommonMessageEN = `
❌ Download failed

⚠️ Error: 
{error_message}
`
)

const (
	ErrorStopDownloadMessageCode = "error_stop_download_message"
	ErrorStopMagnetMessageCode   = "error_stop_magnet_message"
)

const (
	ErrorStopDownloadMessageZH = "❌ 任务已无法取消，可能已完成或不存在。"
	ErrorStopDownloadMessageEN = "❌ Task cannot be cancelled, it may have been completed or does not exist."
	ErrorStopMagnetMessageZH   = "❌ 磁力链接已无法取消，可能已完成或不存在。"
	ErrorStopMagnetMessageEN   = "❌ Magnet link cannot be cancelled, it may have been completed or does not exist."
)
