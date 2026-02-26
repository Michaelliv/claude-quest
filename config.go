package main

import (
	"encoding/json"
	"os"
)

// Config holds user configuration
type Config struct {
	// Equipped accessories
	Hat   string `json:"hat,omitempty"`
	Cape  string `json:"cape,omitempty"`
	Item  string `json:"item,omitempty"` // Held item

	// Display settings
	Scale      int  `json:"scale"`
	Fullscreen bool `json:"fullscreen"`
	Debug      bool `json:"debug"`

	// Audio settings
	SoundEnabled bool    `json:"sound_enabled"`
	Volume       float32 `json:"volume"`

	// Theme/background
	Background string `json:"background"`

	// Language support
	Korean bool `json:"-"` // --korean flag, not persisted
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Scale:        3,
		Fullscreen:   false,
		Debug:        false,
		SoundEnabled: true,
		Volume:       0.7,
		Background:   "study",
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) *Config {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist, use defaults
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		// Invalid JSON, use defaults
		return config
	}

	return config
}

// Save writes the config to a JSON file
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
