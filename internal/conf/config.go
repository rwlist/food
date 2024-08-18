package conf

import (
	"github.com/caarlos0/env/v6"
)

type App struct {
	TelegramBotToken string `env:"BOT_TOKEN"`
	PostgresURL      string `env:"POSTGRES_URL"`
	R2AccountID      string `env:"R2_ACCOUNT_ID"`
	R2KeyID          string `env:"R2_KEY_ID"`
	R2KeySecret      string `env:"R2_KEY_SECRET"`
	R2BucketName     string `env:"R2_BUCKET_NAME"`
	HttpBind         string `env:"HTTP_BIND" envDefault:":9090"`
}

func ParseEnv() (*App, error) {
	cfg := App{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
