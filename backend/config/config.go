package config

import (
	"context"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	TurnstileApi = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type GormConfig struct {
	DSN string `yaml:"dsn"`
}

type CaptchaConfig struct {
	Enable    bool   `yaml:"enable"`
	SecretKey string `yaml:"secret_key"`
}

type JwtConfig struct {
	Realm string `yaml:"realm"`
	Key   string `yaml:"key"`
}

type Config struct {
	GormConfig GormConfig    `yaml:"gorm_config"`
	ReCaptcha  CaptchaConfig `yaml:"captcha"`
	HttpPort   string        `yaml:"http_port"`
	JwtConfig  JwtConfig     `yaml:"jwt"`
}

func NewConfig(ctx context.Context) *Config {
	logger := logrus.WithContext(ctx)

	logger.Debug("loading config")

	config := &Config{}
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		logger.Fatal("your config is lost!")
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		logger.Fatal("parse yaml failed: ", err)
	}

	return config
}
