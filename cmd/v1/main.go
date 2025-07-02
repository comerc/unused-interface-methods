package main

import (
	"flag"
	"fmt"

	"github.com/comerc/unused-interface-methods/pkg/config"
	"github.com/comerc/unused-interface-methods/pkg/linter"
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
		fmt.Printf("Error loading config: %v\n", err)
		config.OsExit(1)
	}

	args := flag.Args()
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	if *verbose {
		fmt.Printf("Analyzing directory: %s\n", dir)
	}

	linter := linter.New(cfg, *verbose)

	err = linter.LoadPackages(dir)
	if err != nil {
		fmt.Printf("Error loading packages: %v\n", err)
		config.OsExit(1)
	}

	linter.ExtractInterfaceMethods()
	unused := !linter.FindUnusedMethods()
	if unused {
		config.OsExit(1)
	}
}
