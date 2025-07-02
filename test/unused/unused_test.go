package unused

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/comerc/unused-interface-methods/pkg/config"
	"github.com/comerc/unused-interface-methods/pkg/stage0"
	"github.com/comerc/unused-interface-methods/pkg/stage1"
	"github.com/comerc/unused-interface-methods/pkg/stage2"
)

var testOutput string

func TestMain(m *testing.M) {
	// Загружаем конфиг и очищаем Ignore
	cfg, err := config.LoadConfig("")
	if err != nil {
		panic(err)
	}
	cfg.Ignore = nil

	// Перехватываем stdout для сбора вывода
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Получаем путь к тестовым данным
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Error getting working directory: %v\n", err))
	}
	testDataDir := filepath.Join(wd, "..", "data")

	// Повторяем логику из main.go
	pkgs, err := stage0.LoadProject(cfg, true, testDataDir)
	if err != nil {
		panic(fmt.Sprintf("Error loading project: %v\n", err))
	}

	usedMethodsByPkg, err := stage1.FindUsedMethods(pkgs, cfg, true)
	if err != nil {
		panic(fmt.Sprintf("Error finding used methods: %v\n", err))
	}

	err = stage2.FindUnusedMethods(pkgs, usedMethodsByPkg, cfg, true)
	if err != nil {
		panic(fmt.Sprintf("Error finding unused methods: %v\n", err))
	}

	// Закрываем pipe и читаем вывод
	w.Close()
	var out bytes.Buffer
	io.Copy(&out, r)
	os.Stdout = oldStdout

	// Фильтруем отладочные сообщения
	lines := strings.Split(out.String(), "\n")
	var filteredLines []string
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "DEBUG:") || strings.HasPrefix(line, "Analyzing") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	testOutput = strings.Join(filteredLines, "\n")

	os.Exit(m.Run())
}
