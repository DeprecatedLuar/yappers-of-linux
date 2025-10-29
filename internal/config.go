package internal

import (
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type NotificationConfig struct {
	Events []string
	Urgent bool
}

type Config struct {
	Notifications string `toml:"notifications"`
	Model         string `toml:"model"`
	Device        string `toml:"device"`
	Language      string `toml:"language"`
	FastMode      bool   `toml:"fast_mode"`
	EnableTyping  bool   `toml:"enable_typing"`
	OutputFile    bool   `toml:"output_file"`
}

func ParseNotifications(notifStr string) NotificationConfig {
	notifStr = strings.TrimSpace(notifStr)

	// Handle disabled cases
	if notifStr == "" || notifStr == "false" || notifStr == "disabled" {
		return NotificationConfig{Events: []string{}, Urgent: false}
	}

	// Parse comma-separated values
	parts := strings.Split(notifStr, ",")
	events := []string{}
	urgent := false

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "urgent" {
			urgent = true
		} else if part == "start" || part == "pause" || part == "stop" {
			events = append(events, part)
		}
	}

	// Handle "urgent" alone as shorthand for "start,urgent"
	if len(events) == 0 && urgent {
		events = []string{"start"}
	}

	return NotificationConfig{Events: events, Urgent: urgent}
}

func (nc NotificationConfig) ShouldNotify(event string) bool {
	for _, e := range nc.Events {
		if e == event {
			return true
		}
	}
	return false
}

func LoadConfig() *Config {
	configDir, err := GetConfigDir()
	if err != nil {
		return &Config{
			Notifications: "urgent",
			Model:         "tiny",
			Device:        "cpu",
			Language:      "en",
			FastMode:      false,
			EnableTyping:  true,
			OutputFile:    false,
		}
	}

	configPath := filepath.Join(configDir, "config.toml")

	cfg := &Config{
		Notifications: "urgent",
		Model:         "tiny",
		Device:        "cpu",
		Language:      "en",
		FastMode:      false,
		EnableTyping:  true,
		OutputFile:    false,
	}
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return cfg
	}

	return cfg
}
