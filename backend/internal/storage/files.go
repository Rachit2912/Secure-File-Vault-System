package storage

import (
	"fmt"
	"io"

	"backend/internal/config"
	storage_go "github.com/supabase-community/storage-go"
)

// UploadFile uploads file content to Supabase Storage at the specified relativePath.
func UploadFile(relativePath string, data io.Reader, contentType string) error {
	if Client == nil || Client.Storage == nil {
		return fmt.Errorf("supabase storage client is not initialized")
	}

	bucket := config.AppConfig.SupabaseBucket
	var options []storage_go.FileOptions
	if contentType != "" {
		options = append(options, storage_go.FileOptions{
			ContentType: &contentType,
		})
	}

	_, err := Client.Storage.UploadFile(bucket, relativePath, data, options...)
	if err != nil {
		return fmt.Errorf("failed to upload file to supabase storage: %w", err)
	}

	return nil
}

// DownloadFile retrieves raw file bytes from Supabase Storage at the specified relativePath.
func DownloadFile(relativePath string) ([]byte, error) {
	if Client == nil || Client.Storage == nil {
		return nil, fmt.Errorf("supabase storage client is not initialized")
	}

	bucket := config.AppConfig.SupabaseBucket
	data, err := Client.Storage.DownloadFile(bucket, relativePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from supabase storage: %w", err)
	}

	return data, nil
}

// DeleteFile removes a file from Supabase Storage at the specified relativePath.
func DeleteFile(relativePath string) error {
	if Client == nil || Client.Storage == nil {
		return fmt.Errorf("supabase storage client is not initialized")
	}

	bucket := config.AppConfig.SupabaseBucket
	_, err := Client.Storage.RemoveFile(bucket, []string{relativePath})
	if err != nil {
		return fmt.Errorf("failed to delete file from supabase storage: %w", err)
	}

	return nil
}
