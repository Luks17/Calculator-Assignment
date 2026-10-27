// Package config loads process configuration from the environment.
package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Host           string   `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port           int      `env:"SERVER_PORT" envDefault:"8080"`
	AppEnv         string   `env:"APP_ENV" envDefault:"production"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envSeparator:","`
}

func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("load environment configuration: %w", err)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("SERVER_PORT must be an integer between 1 and 65535, got %d", cfg.Port)
	}
	origins := cfg.AllowedOrigins[:0]
	for _, origin := range cfg.AllowedOrigins {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	cfg.AllowedOrigins = origins
	return cfg, nil
}

func (c Config) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c Config) GetAppEnv() string { return c.AppEnv }

func (c Config) GetAllowedOrigins() []string { return c.AllowedOrigins }
