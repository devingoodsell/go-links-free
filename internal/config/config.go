package config

import (
	"os"
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
	cfg := &Config{
		Port:          getEnvOrDefault("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		EnableOktaSSO: os.Getenv("ENABLE_OKTA_SSO") == "true",
		JWTSecret:     os.Getenv("JWT_SECRET"),
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
