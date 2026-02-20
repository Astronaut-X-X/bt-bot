package utils

import "fmt"

func FormatPercentage(completed, total int64) string {
	if completed == 0 || total == 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", float64(completed)*100/float64(total))
}

func FormatBytesToSizeString(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"K", "M", "G", "T"}
	return fmt.Sprintf("%.2f %s", float64(size)/float64(div), units[exp])
}
