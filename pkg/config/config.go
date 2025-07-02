package config

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

// OsExit используется для возможности мока os.Exit в тестах
var OsExit = os.Exit

// Config содержит настройки линтера
type Config struct {
	// Паттерны для игнорирования файлов и директорий
	Ignore []string `yaml:"ignore"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Ignore: []string{
			"**/*_test.go",
			"test/**",
			"**/*_mock.go",
			"**/mock/**",
			"**/mocks/**",
		},
	}
}

// LoadConfig загружает конфигурацию из файла или возвращает конфигурацию по умолчанию
func LoadConfig(configPath string) (*Config, error) {
	// Если путь не указан, ищем стандартные места
	if configPath == "" {
		configPath = findConfigFile()
	}

	// Если файл не найден, используем конфигурацию по умолчанию
	if configPath == "" {
		return DefaultConfig(), nil
	}

	// Проверяем существование файла
	_, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// findConfigFile ищет конфигурационный файл в стандартных местах
func findConfigFile() string {
	candidates := []string{
		".unused-interface-methods.yml",
		"unused-interface-methods.yml",
		".config/unused-interface-methods.yml",
		".unused-interface-methods.yaml",
		"unused-interface-methods.yaml",
		".config/unused-interface-methods.yaml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// ShouldIgnore проверяет, нужно ли игнорировать файл или директорию
func (c *Config) ShouldIgnore(filePath string) bool {
	// Нормализуем путь
	filePath = filepath.Clean(filePath)

	for _, pattern := range c.Ignore {
		if c.matchPattern(pattern, filePath) {
			return true
		}
	}

	return false
}

// matchPattern проверяет соответствие файла паттерну
func (c *Config) matchPattern(pattern, filePath string) bool {
	// Нормализуем путь
	filePath = filepath.Clean(filePath)

	// Используем doublestar для проверки соответствия
	matched, _ := doublestar.Match(pattern, filePath)
	return matched
}

// GetRelativePath преобразует путь в относительный от текущей директории
func GetRelativePath(filePath string) string {
	wd, err := os.Getwd()
	if err != nil {
		return filePath
	}

	rel, err := filepath.Rel(wd, filePath)
	if err != nil {
		return filePath
	}

	return rel
}
