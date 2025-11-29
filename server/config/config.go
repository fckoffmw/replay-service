package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBDSN      string
	StorageDir string
	LogLevel   string
}

func (c Config) String() string {
	return fmt.Sprintf("{  PORT=%s,  DBDSN=%s,  STORAGE_DIR=%s,  LOG_LEVEL=%s  }",
		c.Port, c.DBDSN, c.StorageDir, c.LogLevel)
}
func Load() (*Config, error) {
	if root, err := findProjectRoot(); err == nil {
		envPath := filepath.Join(root, ".env")
		_ = godotenv.Load(envPath)
	} else {
		envPaths := []string{
			".env",
			"../../.env",
			"../../../.env",
		}
		for _, path := range envPaths {
			if err := godotenv.Load(path); err == nil {
				break
			}
		}
	}

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBDSN:      getEnv("DB_DSN", ""),
		StorageDir: getEnv("STORAGE_DIR", "./storage"),
		LogLevel:   getEnv("LOG_LEVEL", "debug"),
	}

	if cfg.DBDSN == "" {
		return nil, fmt.Errorf("DB_DSN is required (set DB_DSN environment variable or create .env file in project root)")
	}

	return cfg, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
