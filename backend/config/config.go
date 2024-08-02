package config

type Config struct {
	DatabaseURL string
	TraceSQL    bool
	Debug       bool
	LogRequests bool

	Address string
	Port    uint16
}
