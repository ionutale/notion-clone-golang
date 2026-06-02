package config

import "github.com/caarlos0/env/v11"

type Config struct {
	DatabaseURL         string `env:"DATABASE_URL" required:"true"`
	Port                string `env:"PORT" envDefault:"8080"`
	DevMode             bool   `env:"DEV_MODE" envDefault:"false"`
	StorageEmulatorHost string `env:"STORAGE_EMULATOR_HOST" envDefault:""`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
