package utils

import (
	"backend/internal/config"
)

// GetUserQuotaBytes returns the per-user quota in bytes.
// Reads USER_QUOTA_MB from env, defaults to 10 MB if unset or invalid.
func GetUserQuotaBytes() int64 {
	mb := config.AppConfig.UserQuotaMB
	if mb < 0 {
		mb = 10 //  default: 10MB
	}
	return int64(mb) * 1024 * 1024
}
