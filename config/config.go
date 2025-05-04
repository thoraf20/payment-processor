package config

import (
	"fmt"
	
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPPort         			string `envconfig:"HTTP_PORT" default:"8080"`
	DatabaseURL     			string `envconfig:"DATABASE_URL" required:"true"`
	StripeAPIKey     			string `envconfig:"STRIPE_API_KEY" required:"true"`
	FlutterWaveAPIKey     string `envconfig:"FLUTTERWAVE_API_KEY" required:"true"`
	PayStackAPIKey     		string `envconfig:"PAYSTACK_API_KEY" required:"true"`

	Environment      			string `envconfig:"ENVIRONMENT" default:"development"`
	LogLevel         			string `envconfig:"LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}