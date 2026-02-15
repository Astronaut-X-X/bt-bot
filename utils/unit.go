package utils

import "fmt"

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
