package i18n

import "strings"

const (
	LangZH = "zh"
	LangEN = "en"
)

var (
	ZH_MAP map[string]string
	EN_MAP map[string]string
)

func init() {
	ZH_MAP = map[string]string{
		// Command
		"start_message": StartMessageZH,
		"self_message":  SelfMessageZH,
		"help_message":  HelpMessageZH,

		// Magnet
		MagnetInvalidLinkMessageCode: MagnetInvalidLinkMessageZH,
		MagnetProcessingMessageCode:  MagnetProcessingMessageZH,
		MagnetErrorMessageCode:       MagnetErrorMessageZH,
		MagnetSuccessMessageCode:     MagnetSuccessMessageZH,

		// Download
		DownloadStartMessageCode:      DownloadStartMesssageZH,
		DownloadProcessingMessageCode: DownloadProcessingMessageZH,
		DownloadSuccessMessageCode:    DownloadSuccessMessageZH,
		DownloadFailedMessageCode:     DownloadFailedMessageZH,
	}
	EN_MAP = map[string]string{
		// Command
		"start_message": StartMessageEN,
		"self_message":  SelfMessageEN,
		"help_message":  HelpMessageEN,

		// Magnet
		MagnetInvalidLinkMessageCode: MagnetInvalidLinkMessageEN,
		MagnetProcessingMessageCode:  MagnetProcessingMessageEN,
		MagnetErrorMessageCode:       MagnetErrorMessageEN,
		MagnetSuccessMessageCode:     MagnetSuccessMessageEN,

		// Download
		DownloadStartMessageCode:      DownloadStartMesssageEN,
		DownloadProcessingMessageCode: DownloadProcessingMessageEN,
		DownloadSuccessMessageCode:    DownloadSuccessMessageEN,
		DownloadFailedMessageCode:     DownloadFailedMessageEN,
	}
}

func Text(key string, lang ...string) string {
	translationLang := LangZH
	if len(lang) > 0 {
		translationLang = lang[0]
	}

	switch translationLang {
	case LangZH:
		return ZH_MAP[key]
	case LangEN:
		return EN_MAP[key]
	default:
		return EN_MAP[key]
	}
}

func Replace(text string, placeholders map[string]string) string {
	for placeholder, value := range placeholders {
		text = strings.Replace(text, placeholder, value, 1)
	}
	return text
}
