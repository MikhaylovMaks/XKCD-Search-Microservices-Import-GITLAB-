package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel   string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	BotToken   string `yaml:"bot_token" env:"BOT_TOKEN" env-default:""`
	APIAddress string `yaml:"api_address" env:"API_ADDRESS" env-default:"http://api:8080"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
