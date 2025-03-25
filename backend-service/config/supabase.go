package config

import (
	"os"

	"github.com/supabase-community/supabase-go"
)

func NewSupabaseClient() *supabase.Client {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		panic(err)
	}

	return client
}
