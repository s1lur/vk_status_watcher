package main

import (
	"context"
	"time"
    "github.com/heetch/confita"
    "github.com/heetch/confita/backend/env"
    "github.com/heetch/confita/backend/file"
)

type Config struct {
    AccountID int `config:"account_id,requied" yaml:"accountId"`
    ChatID int `config:"chat_id,required" yaml:"chatId"`
    TGToken string `config:"tg_token,required"`
    VKToken string `config:"vk_token,required"`
	PollingInterval time.Duration `config:"polling_interval" yaml:"pollingInterval"`
}

func LoadConfig(ctx context.Context, yamlPath string) (*Config, error) {
	cfg := &Config{
		PollingInterval: time.Hour,
	}
	err := confita.NewLoader(
		env.NewBackend(),
		file.NewBackend(yamlPath),
	).Load(ctx, cfg)
	return cfg, err
}