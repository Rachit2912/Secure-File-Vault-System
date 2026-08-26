package storage

import (
	"fmt"

	"backend/internal/config"
	"github.com/supabase-community/supabase-go"
)

var Client *supabase.Client

func Init() error {
	if config.AppConfig.SupabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL is not configured")
	}

	if config.AppConfig.SupabaseServiceKey == "" {
		return fmt.Errorf("SUPABASE_SERVICE_KEY is not configured")
	}

	var err error

	Client, err = supabase.NewClient(
		config.AppConfig.SupabaseURL,
		config.AppConfig.SupabaseServiceKey,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	return nil
}
