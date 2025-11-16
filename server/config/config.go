package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	Port       string
	DBDSN      string
	StorageDir string
}

// Load загружает конфигурацию из .env файла и переменных окружения
func Load() (*Config, error) {
	// Пытаемся найти корень проекта и загрузить .env оттуда
	if root, err := findProjectRoot(); err == nil {
		envPath := filepath.Join(root, ".env")
		_ = godotenv.Load(envPath)
	} else {
		// Если не удалось найти корень, пробуем загрузить из разных мест
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
	}

	if cfg.DBDSN == "" {
		return nil, fmt.Errorf("DB_DSN is required (set DB_DSN environment variable or create .env file in project root)")
	}

	return cfg, nil
}

// findProjectRoot пытается найти корень проекта, ища go.mod файл
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

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
