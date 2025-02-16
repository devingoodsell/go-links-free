package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string `json:"port"`
	DatabaseURL      string `json:"database_url"`
	EnableOktaSSO    bool   `json:"enable_okta_sso"`
	OktaOrgURL       string `json:"okta_org_url,omitempty"`
	OktaClientID     string `json:"okta_client_id,omitempty"`
	OktaClientSecret string `json:"okta_client_secret,omitempty"`
	JWTSecret        string `json:"jwt_secret"`
}

func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load() // Ignore error from .env load

	// Required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	// Optional environment variables with defaults
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	enableOktaSSO := os.Getenv("ENABLE_OKTA_SSO") == "true"

	cfg := &Config{
		Port:          port,
		DatabaseURL:   dbURL,
		JWTSecret:     jwtSecret,
		EnableOktaSSO: enableOktaSSO,
	}

	if cfg.EnableOktaSSO {
		cfg.OktaOrgURL = os.Getenv("OKTA_ORG_URL")
		cfg.OktaClientID = os.Getenv("OKTA_CLIENT_ID")
		cfg.OktaClientSecret = os.Getenv("OKTA_CLIENT_SECRET")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
