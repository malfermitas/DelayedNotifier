package config

import (
	"os"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Email    EmailConfig    `mapstructure:"email"`
}

type ServerConfig struct {
	Port     string `mapstructure:"port"`
	EnableUI bool   `mapstructure:"enable_ui"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

type TelegramConfig struct {
	Token       string
	BotUsername string
}

type EmailConfig struct {
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort int    `mapstructure:"smtp_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

func Load(path string, pathEnv string) (*Config, error) {
	c := config.New()

	c.EnableEnv("APP")
	if err := c.LoadConfigFiles(path); err != nil {
		return nil, err
	}

	if err := c.LoadEnvFiles(pathEnv); err != nil {
		return nil, err
	}

	cfg := Config{
		Server: ServerConfig{EnableUI: true},
	}
	if err := c.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.Telegram.Token = os.Getenv("TELEGRAM_TOKEN")
	cfg.Telegram.BotUsername = os.Getenv("TELEGRAM_BOT_USERNAME")

	return &cfg, nil
}
