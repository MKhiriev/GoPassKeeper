package config

import (
	"fmt"
	"time"
)

// ClientApp содержит настройки приложения для клиента
type ClientApp struct {
	HashKey string
}

// ClientAdapter содержит настройки сетевых подключений
type ClientAdapter struct {
	HTTPAddress    string
	GRPCAddress    string
	RequestTimeout time.Duration
}

// ClientDB содержит настройки подключения к БД
type ClientDB struct {
	DSN string
}

// ClientStorage группирует настройки хранилищ
type ClientStorage struct {
	DB ClientDB
}

type ClientWorkers struct {
	SyncInterval time.Duration
}

// ClientConfig — основная структура конфигурации клиента
type ClientConfig struct {
	App     ClientApp
	Adapter ClientAdapter
	Storage ClientStorage
	Workers ClientWorkers
}

func GetClientConfig() (*ClientConfig, error) {
	cfg, err := GetStructuredConfig()
	if err != nil {
		return nil, fmt.Errorf("error get structured config: %w", err)
	}

	clientCfg := &ClientConfig{
		App: ClientApp{
			HashKey: cfg.App.HashKey,
		},
		Adapter: ClientAdapter{
			HTTPAddress:    cfg.Adapter.HTTPAddress,
			GRPCAddress:    cfg.Adapter.GRPCAddress,
			RequestTimeout: cfg.Adapter.RequestTimeout,
		},
		Storage: ClientStorage{
			DB: ClientDB{
				DSN: cfg.Storage.DB.DSN,
			},
		},
		Workers: ClientWorkers{SyncInterval: cfg.Workers.SyncInterval},
	}

	return clientCfg, clientCfg.validate()
}
