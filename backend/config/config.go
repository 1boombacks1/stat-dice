package config

import (
	"time"

	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppName   string                `yaml:"app_name" default:"Stat Dice"`
	StartPage templates.PageContent `yaml:"start_page" enum:"find,create,leaderboard,completed" default:"find"`

	DatabaseURL string `short:"c" yaml:"dsn" default:"postgres://d0c:d0c@localhost:5433/stat_dice" help:"database url"`
	TraceSQL    bool   `yaml:"trace_sql" help:"trace SQL statements"`

	Debug       bool `short:"d" yaml:"debug"`
	LogRequests bool `yaml:"log_requests" help:"logging http requests"`

	Host string `yaml:"host" default:"127.0.0.1" help:"listen http address"`
	Port uint16 `short:"p" yaml:"port" default:"8080" help:"port"`

	JWTKey      string        `yaml:"jwt_key"`
	JWTDuration time.Duration `yaml:"jwt_duration" default:"252h" help:"JWT expires duration - по дэфолту полторы недели"`

	BcryptCost int `yaml:"bcrypt_cost" default:"7"`
}

func MustParseYAML(path string, cfg *Config) {
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic(err)
	}
}
