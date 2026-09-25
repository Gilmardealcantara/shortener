package config

type Config struct {
	BaseURL     string
	PostgresDSN string
}

func New() *Config {
	pgDsn := "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable"
	return &Config{
		BaseURL:     "http://localhost:8080",
		PostgresDSN: pgDsn,
	}
}
