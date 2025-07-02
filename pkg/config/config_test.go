package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	cfg := DefaultConfig()

	testCases := []struct {
		path string
		want bool
	}{
		{"foo_test.go", true},                                     // **/*_test.go
		{"internal/order/order_test.go", true},                    // **/*_test.go
		{filepath.Join("test", "utils", "helper.go"), true},       // test/**
		{filepath.Join("service", "mocks", "user_mock.go"), true}, // **/mocks/**
		{filepath.Join("service", "mockups", "data.go"), false},   // не должен игнорировать
		{filepath.Join("cmd", "main.go"), false},                  // не должен игнорировать
	}

	for _, tc := range testCases {
		got := cfg.ShouldIgnore(tc.path)
		if got != tc.want {
			t.Errorf("ShouldIgnore(%s) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Сохраняем и восстанавливаем текущую директорию
	startDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(startDir)

	// Создаем временную директорию для тестов
	tmpDir := t.TempDir()

	t.Run("config not found in start dir", func(t *testing.T) {
		// Проверяем, что в начальной директории нет конфига
		if _, err := os.Stat(filepath.Join(startDir, ".unused-interface-methods.yml")); err == nil {
			t.Fatal("config file exists in start dir, test cannot proceed")
		}

		// Должны получить конфиг по умолчанию
		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		want := DefaultConfig()
		if !reflect.DeepEqual(cfg, want) {
			t.Errorf("LoadConfig() = %v, want %v", cfg, want)
		}
	})

	t.Run("config found in temp dir", func(t *testing.T) {
		// Переходим во временную директорию
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Создаем конфиг во временной директории
		content := []byte(`ignore:
  - "vendor/**"
  - "**/*.pb.go"`)
		if err := os.WriteFile(".unused-interface-methods.yml", content, 0644); err != nil {
			t.Fatal(err)
		}

		// Проверяем что файл существует
		if _, err := os.Stat(".unused-interface-methods.yml"); err != nil {
			t.Fatal("config file was not created")
		}

		// Загружаем конфиг
		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		want := &Config{
			Ignore: []string{
				"vendor/**",
				"**/*.pb.go",
			},
		}
		if !reflect.DeepEqual(cfg, want) {
			t.Errorf("LoadConfig() = %v, want %v", cfg, want)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		// Переходим во временную директорию
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Создаем некорректный конфиг
		content := []byte(`ignore: [`)
		if err := os.WriteFile(".unused-interface-methods.yml", content, 0644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadConfig("")
		if err == nil {
			t.Error("LoadConfig() error = nil, want error for invalid yaml")
		}
	})

	t.Run("explicit config path", func(t *testing.T) {
		// Создаем конфиг в нестандартном месте
		content := []byte(`ignore:
  - "custom/**"`)
		customPath := filepath.Join(tmpDir, "custom.yml")
		if err := os.WriteFile(customPath, content, 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadConfig(customPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		want := &Config{
			Ignore: []string{
				"custom/**",
			},
		}
		if !reflect.DeepEqual(cfg, want) {
			t.Errorf("LoadConfig() = %v, want %v", cfg, want)
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		// Переходим во временную директорию
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// Создаем поддиректорию без прав на чтение
		noAccessDir := filepath.Join(tmpDir, "noaccess")
		if err := os.Mkdir(noAccessDir, 0700); err != nil {
			t.Fatal(err)
		}

		// Создаем конфиг
		content := []byte(`ignore:
  - "vendor/**"`)
		configPath := filepath.Join(noAccessDir, ".unused-interface-methods.yml")
		if err := os.WriteFile(configPath, content, 0644); err != nil {
			t.Fatal(err)
		}

		// Убираем права на чтение у директории
		if err := os.Chmod(noAccessDir, 0000); err != nil {
			t.Fatal(err)
		}

		// Пытаемся загрузить конфиг
		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("LoadConfig() error = nil, want error for permission denied")
		}

		// Восстанавливаем права для очистки
		os.Chmod(noAccessDir, 0700)
	})
}
