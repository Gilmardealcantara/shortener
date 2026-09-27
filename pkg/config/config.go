package config

type Config struct {
	BaseURL      string
	PostgresDSN  string
	RedisDSN     string
	OTelEndpoint string // OTLP gRPC endpoint, e.g. "localhost:4317"
}

func New() *Config {
	pgDsn := "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable"
	return &Config{
		BaseURL:      "http://localhost:8080",
		PostgresDSN:  pgDsn,
		RedisDSN:     "redis://localhost:6379/0",
		OTelEndpoint: "localhost:4317",
	}
}
