package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig       `mapstructure:"app"`
	Api      ApiConfig       `mapstructure:"api"`
	Telegram TelegramConfig  `mapstructure:"telegram"`
	Channels []ChannelConfig `mapstructure:"channels"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
}

// ApiConfig holds API-level configuration
type ApiConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	CertLocation string `mapstructure:"cert_location"`
	KeyLocation  string `mapstructure:"key_location"`
}

// TelegramConfig holds Telegram-level configuration
type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatId   int64  `mapstructure:"chat_id"`
}

// ChannelConfig holds Channel-level configuration
type ChannelConfig struct {
	Name           string `mapstructure:"name"`
	Needle         string `mapstructure:"needle"`
	TelegramToken  string `mapstructure:"telegram_token"`
	TelegramChatId int64  `mapstructure:"telegram_chat_id"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// Enable environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {

	v.SetDefault("app.name", "loki-notification")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_level", "info")

	v.SetDefault("api.host", "0.0.0.0")
	v.SetDefault("api.port", "7777")
	v.SetDefault("api.cert_location", "")
	v.SetDefault("api.key_location", "")

	v.SetDefault("telegram.bot_token", "")
	v.SetDefault("telegram.chat_id", 0)

	v.SetDefault("channels", []ChannelConfig{})
}
