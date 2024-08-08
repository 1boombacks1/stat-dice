package config

import "time"

type Config struct {
	DatabaseURL string `yaml:"dsn" env-default:"postgres://root:root@localhost:5433/statDice"`
	TraceSQL    bool   `yaml:"trace_sql"`

	Debug       bool `yaml:"debug"`
	LogRequests bool `yaml:"log_requests"`

	Address string `yaml:"address" env-default:"127.0.0.1"`
	Port    uint16 `yaml:"port" env-default:"8080"`

	JWTKey      string        `yaml:"jwt_key"`
	JWTDuration time.Duration `yaml:"jwt_duration" default-env:"252h" env-description:"по дэфолту это полторы недели"`

	BcryptCost int `yaml:"bcrypt_cost" env-default:"42"`
}
