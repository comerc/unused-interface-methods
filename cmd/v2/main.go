// Соглашение по путям:
// Все пути в проекте относительные. Это базовое соглашение, которое применяется
// во всех пакетах и компонентах проекта.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/comerc/unused-interface-methods/pkg/config"
	"github.com/comerc/unused-interface-methods/pkg/stage0"
	"github.com/comerc/unused-interface-methods/pkg/stage1"
	"github.com/comerc/unused-interface-methods/pkg/stage2"
)

func main() {
	var (
		verbose = flag.Bool("v", false, "Verbose output")
		help    = flag.Bool("h", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Unused Interface Methods - finds unused interface methods")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  unused-interface-methods [flags] [path]")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -v    Verbose output")
		fmt.Println("  -h    Show this help")
		fmt.Println()
		fmt.Println("Config file:")
		fmt.Println("  Automatically looks for .unused-interface-methods.yml")
		fmt.Println("  Example ignore patterns: \"**/*_test.go\", \"test/**\", \"**/mock/**\"")
		fmt.Println()
		fmt.Println("Note: Generic interfaces are detected but not analyzed (warnings will be shown)")
		config.OsExit(0)
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		config.OsExit(1)
	}

	args := flag.Args()
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Analyzing directory: %s\n", dir)
	}

	// Stage 0: Загружаем AST всего проекта в память
	pkgs, err := stage0.LoadProject(cfg, *verbose, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		config.OsExit(1)
	}

	// Stage 1: Находим все используемые методы интерфейсов
	usedMethodsByPkg, err := stage1.FindUsedMethods(pkgs, cfg, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding used methods: %v\n", err)
		config.OsExit(1)
	}

	// Stage 2: Проверяем неиспользуемые методы через staticcheck
	err = stage2.FindUnusedMethods(pkgs, usedMethodsByPkg, cfg, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding unused methods: %v\n", err)
		config.OsExit(1)
	}
}
