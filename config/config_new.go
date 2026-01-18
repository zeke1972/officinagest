package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Database DatabaseConfig
	App      AppConfig
	Backup   BackupConfig
}

type DatabaseConfig struct {
	URI      string
	Name     string
	Timeout  time.Duration
	Username string
	Password string
}

type AppConfig struct {
	Name       string
	Version    string
	DebugMode  bool
	LogFile    string
	BackupPath string
}

type BackupConfig struct {
	Enabled  bool
	Interval time.Duration
	MaxFiles int
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".officina")
	os.MkdirAll(dataDir, 0755)

	return &Config{
		Database: DatabaseConfig{
			URI:      "mongodb://localhost:27017", // Default MongoDB URI
			Name:     "officina",
			Timeout:  5 * time.Second,
			Username: "",
			Password: "",
		},
		App: AppConfig{
			Name:       "Officina Manager",
			Version:    "2.0.0",
			DebugMode:  false,
			LogFile:    filepath.Join(dataDir, "debug.log"),
			BackupPath: filepath.Join(dataDir, "backups"),
		},
		Backup: BackupConfig{
			Enabled:  true,
			Interval: 24 * time.Hour,
			MaxFiles: 7,
		},
	}
}

func (c *Config) Validate() error {
	if c.Database.URI == "" {
		return fmt.Errorf("database URI non può essere vuoto")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name non può essere vuoto")
	}

	logDir := filepath.Dir(c.App.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("impossibile creare directory log: %w", err)
	}

	if c.Backup.Enabled && c.App.BackupPath == "" {
		return fmt.Errorf("backup path non può essere vuoto quando i backup sono abilitati")
	}

	if c.Backup.Enabled {
		if err := os.MkdirAll(c.App.BackupPath, 0755); err != nil {
			return fmt.Errorf("impossibile creare directory backup: %w", err)
		}
	}

	return nil
}

func LoadOrDefault() (*Config, error) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
