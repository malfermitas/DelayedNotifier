package config

import (
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
	Port string `mapstructure:"port"`
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
	Token  string `mapstructure:"token"`
	ChatID int64  `mapstructure:"chat_id"`
}

type EmailConfig struct {
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort int    `mapstructure:"smtp_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

func Load(path string) (*Config, error) {
	c := config.New()

	c.EnableEnv("APP")

	if path != "" {
		if err := c.LoadConfigFiles(path); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := c.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
