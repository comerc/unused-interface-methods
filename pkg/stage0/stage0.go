package stage0

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"github.com/comerc/unused-interface-methods/pkg/config"
)

// Package представляет пакет с его файлами и FileSet
type Package struct {
	Fset  *token.FileSet
	Files map[string]*ast.File
	Info  *types.Info
}

// LoadProject загружает AST всего проекта в память
func LoadProject(cfg *config.Config, verbose bool, pkgPath string) (map[string]*Package, error) {
	pkgs := make(map[string]*Package)
	fset := token.NewFileSet() // Один FileSet для всех файлов

	// Обходим все файлы в проекте
	err := filepath.WalkDir(pkgPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем файлы по конфигурации
		if cfg.ShouldIgnore(path) {
			if verbose {
				fmt.Fprintf(os.Stderr, "DEBUG: пропускаем файл %s\n", path)
			}
			return nil
		}

		// Пропускаем директории и не-.go файлы
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Парсим файл
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "DEBUG: ошибка парсинга файла %s: %v\n", path, err)
			}
			return nil
		}

		// Получаем путь к пакету
		dir := filepath.Dir(path)
		relDir, err := filepath.Rel(pkgPath, dir)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "DEBUG: ошибка получения относительного пути: %v\n", err)
			}
			return nil
		}
		fullPkgPath := "github.com/comerc/unused-interface-methods/" + relDir

		// Создаем пакет если его еще нет
		if _, ok := pkgs[fullPkgPath]; !ok {
			pkgs[fullPkgPath] = &Package{
				Fset:  fset,
				Files: make(map[string]*ast.File),
				Info: &types.Info{
					Types: make(map[ast.Expr]types.TypeAndValue),
					Defs:  make(map[*ast.Ident]types.Object),
					Uses:  make(map[*ast.Ident]types.Object),
				},
			}
		}

		// Добавляем файл в пакет
		pkgs[fullPkgPath].Files[path] = file

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка обхода файлов: %v", err)
	}

	// Анализируем типы для каждого пакета
	for pkgPath, pkg := range pkgs {
		// Собираем все файлы пакета в слайс для анализа
		var files []*ast.File
		for _, file := range pkg.Files {
			files = append(files, file)
		}

		// Создаем конфигурацию для анализа типов
		conf := types.Config{
			Importer: importer.Default(),
			Error: func(err error) {
				if verbose {
					fmt.Fprintf(os.Stderr, "DEBUG: ошибка анализа типов: %v\n", err)
				}
			},
		}

		// Анализируем типы пакета
		_, err := conf.Check(pkgPath, pkg.Fset, files, pkg.Info)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "DEBUG: ошибка проверки типов в пакете %s: %v\n", pkgPath, err)
			}
			continue // пропускаем пакет с ошибками
		}
	}

	return pkgs, nil
}
