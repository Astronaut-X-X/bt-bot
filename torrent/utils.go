package torrent

import (
	"strings"
)

// ExtractMagnetLink 从文本中提取磁力链接
func ExtractMagnetLink(text string) string {
	if text == "" {
		return ""
	}

	// 如果是命令，提取参数
	if strings.HasPrefix(text, "/magnet") {
		parts := strings.Fields(text)
		if len(parts) > 1 {
			text = strings.Join(parts[1:], " ")
		} else {
			return ""
		}
	}

	// 查找磁力链接
	if strings.HasPrefix(text, "magnet:") {
		// 提取完整的磁力链接（到行尾或空格）
		spaceIndex := strings.Index(text, " ")
		if spaceIndex > 0 {
			return text[:spaceIndex]
		}
		return text
	}

	// 尝试从文本中查找磁力链接
	start := strings.Index(text, "magnet:")
	if start >= 0 {
		remaining := text[start:]
		spaceIndex := strings.Index(remaining, " ")
		if spaceIndex > 0 {
			return remaining[:spaceIndex]
		}
		return remaining
	}

	return ""
}

func ExtractTorrentInfoHash(magnetLink string) string {
	// magnetLink 格式: magnet:?xt=urn:btih:<infohash>[&...]
	if !strings.HasPrefix(magnetLink, "magnet:") {
		return ""
	}
	startIdx := strings.Index(magnetLink, "xt=urn:btih:")
	if startIdx == -1 {
		return ""
	}
	// 找到 xt=urn:btih: 后面的部分
	substr := magnetLink[startIdx+len("xt=urn:btih:"):]
	endIdx := strings.IndexAny(substr, "&")
	var infoHash string
	if endIdx != -1 {
		infoHash = substr[:endIdx]
	} else {
		infoHash = substr
	}
	// infohash 可能带有大写, 我们统一转为小写
	infoHash = strings.ToLower(infoHash)
	return infoHash
}
