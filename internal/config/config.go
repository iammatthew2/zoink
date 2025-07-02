package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the minimal configuration for Zoink
type Config struct {
	DatabasePath    string   `json:"database_path"`
	ExcludePatterns []string `json:"exclude_patterns"`
	// Optional user overrides (only present if customized)
	MaxResults int     `json:"max_results,omitempty"`
	Threshold  float64 `json:"threshold,omitempty"`
}

// Default returns a config with minimal required settings
func Default() *Config {
	return &Config{
		DatabasePath:    "", // Will be set to config dir + "zoink.db"
		ExcludePatterns: []string{".git", "node_modules", "__pycache__", ".vscode", ".idea"},
		// MaxResults and Threshold are 0 (not set) - flags provide defaults
	}
}

// GetConfigDir returns the cross-platform config directory
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err == nil {
		return filepath.Join(configDir, "zoink"), nil
	}

	// Fallback to home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "zoink"), nil
}

// Load reads config from the standard location, creating defaults if it doesn't exist
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return Default(), nil // Return defaults if we can't get config dir
	}

	configPath := filepath.Join(configDir, "config.json")

	// If config doesn't exist, return defaults (don't create file yet)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := Default()
		cfg.DatabasePath = filepath.Join(configDir, "zoink.db")
		return cfg, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		cfg := Default()
		cfg.DatabasePath = filepath.Join(configDir, "zoink.db")
		return cfg, nil // Return defaults on read error
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		cfg = *Default()
		cfg.DatabasePath = filepath.Join(configDir, "zoink.db")
		return &cfg, nil // Return defaults on parse error
	}

	// Ensure database path is set
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = filepath.Join(configDir, "zoink.db")
	}

	return &cfg, nil
}

// Save writes the config to the standard location
func (c *Config) Save() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
