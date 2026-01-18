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
            Path:    filepath.Join(dataDir, "officina.db"),
            Timeout: 1 * time.Second,
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
    if c.Database.Path == "" {
        return fmt.Errorf("database path non può essere vuoto")
    }

    dbDir := filepath.Dir(c.Database.Path)
    if err := os.MkdirAll(dbDir, 0755); err != nil {
        return fmt.Errorf("impossibile creare directory database: %w", err)
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
