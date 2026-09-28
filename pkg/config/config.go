package config

import "os"

type Config struct {
	BaseURL        string
	PostgresDSN    string
	RedisDSN       string
	OTelEndpoint   string // OTLP/HTTP host:port only — no scheme, no path, e.g. "otlp.nr-data.net"
	NewRelicAPIKey string // New Relic license key, set via NEW_RELIC_LICENSE_KEY env var
	EnebleOTel     bool   // Enable OpenTelemetry instrumentation
}

func New() *Config {
	pgDsn := "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable"
	return &Config{
		BaseURL:        "http://localhost:8080",
		PostgresDSN:    pgDsn,
		RedisDSN:       "redis://localhost:6379/0",
		OTelEndpoint:   "otlp.nr-data.net",
		NewRelicAPIKey: os.Getenv("NEW_RELIC_LICENSE_KEY"),
		EnebleOTel:     enableOTel(),
	}
}

func enableOTel() bool {
	if os.Getenv("ENABLE_OTEL") == "true" {
		return true
	}
	return false
}
