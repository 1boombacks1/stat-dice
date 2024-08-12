package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DatabaseURL string `short:"c" yaml:"dsn" default:"postgres://d0c:d0c@localhost:5433/stat_dice" help:"database url"`
	TraceSQL    bool   `yaml:"trace_sql" help:"trace SQL statements"`

	Debug       bool `short:"d" yaml:"debug"`
	LogRequests bool `yaml:"log_requests" help:"logging http requests"`

	Address string `short:"h" yaml:"address" default:"127.0.0.1" help:"listen http address"`
	Port    uint16 `short:"p" yaml:"port" default:"8080" help:"port"`

	JWTKey      string        `yaml:"jwt_key"`
	JWTDuration time.Duration `yaml:"jwt_duration" default:"252h" help:"JWT expires duration - по дэфолту полторы недели"`

	BcryptCost int `yaml:"bcrypt_cost" default:"42"`
}

func MustParseYAML(path string, cfg *Config) {
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic(err)
	}
}
