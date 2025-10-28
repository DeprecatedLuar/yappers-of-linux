package internal

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type NotificationMode string

const (
	NotificationNormal  NotificationMode = "true"
	NotificationDisable NotificationMode = "false"
	NotificationUrgent  NotificationMode = "urgent"
)

type Config struct {
	Notifications NotificationMode `toml:"notifications"`
	Model         string           `toml:"model"`
	Device        string           `toml:"device"`
	Language      string           `toml:"language"`
	FastMode      bool             `toml:"fast_mode"`
}

func LoadConfig() *Config {
	configDir, err := GetConfigDir()
	if err != nil {
		return &Config{
			Notifications: NotificationUrgent,
			Model:         "tiny",
			Device:        "cpu",
			Language:      "en",
			FastMode:      false,
		}
	}

	configPath := filepath.Join(configDir, "config.toml")

	cfg := &Config{
		Notifications: NotificationUrgent,
		Model:         "tiny",
		Device:        "cpu",
		Language:      "en",
		FastMode:      false,
	}
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return cfg
	}

	return cfg
}
