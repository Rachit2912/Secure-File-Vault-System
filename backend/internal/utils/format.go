package utils

import "fmt"

// fn. for converting bytes into human-readable KB/MB/GB string
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	// KB, MB, GB, TB...
	pre := "KMGTPE"[exp : exp+1]
	return fmt.Sprintf("%.1f %sB", float64(b)/float64(div), pre)
}
