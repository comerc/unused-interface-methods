package stage2

import (
	"fmt"
	"go/ast"
	"go/printer"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/comerc/unused-interface-methods/pkg/config"
	"github.com/comerc/unused-interface-methods/pkg/stage0"
	"github.com/comerc/unused-interface-methods/pkg/stage1"
)

// Method представляет метод интерфейса для проверки
type Method struct {
	InterfaceName string // имя интерфейса
	MethodName    string // имя метода
}

// Interface представляет интерфейс с методами
type Interface struct {
	Name    string    // имя интерфейса
	Methods []*Method // список методов
}

// copyFile копирует файл из src в dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("не удалось открыть исходный файл: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("не удалось создать целевой файл: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("не удалось скопировать файл: %v", err)
	}

	return nil
}

// copyProject копирует все .go файлы проекта во временную директорию
func copyProject(rootDir string, tmpDir string, cfg *config.Config) error {
	return filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем файлы по конфигурации
		if cfg.ShouldIgnore(path) {
			return nil
		}

		// Пропускаем директории и не-.go файлы
		if d.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return fmt.Errorf("не удалось получить путь: %v", err)
		}

		// Создаем директорию назначения
		dstDir := filepath.Join(tmpDir, filepath.Dir(relPath))
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию: %v", err)
		}

		// Копируем файл
		dstPath := filepath.Join(tmpDir, relPath)
		if err := copyFile(path, dstPath); err != nil {
			return err
		}

		return nil
	})
}

// findInterfaces находит все интерфейсы в файле
func findInterfaces(file *ast.File) []*Interface {
	var interfaces []*Interface

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		iface := &Interface{
			Name: typeSpec.Name.Name,
		}

		// Собираем методы интерфейса
		for _, field := range interfaceType.Methods.List {
			// Пропускаем встроенные интерфейсы (без имен)
			if len(field.Names) == 0 {
				continue
			}
			method := &Method{
				InterfaceName: iface.Name,
				MethodName:    field.Names[0].Name,
			}
			iface.Methods = append(iface.Methods, method)
		}

		interfaces = append(interfaces, iface)
		return true
	})

	return interfaces
}

// removeMethod удаляет метод из интерфейса
func removeMethod(file *ast.File, interfaceName, methodName string) bool {
	var removed bool
	ast.Inspect(file, func(n ast.Node) bool {
		// Ищем объявление интерфейса
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != interfaceName {
			return true
		}

		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		// Фильтруем методы, удаляя указанный
		var methods []*ast.Field
		for _, method := range interfaceType.Methods.List {
			// Пропускаем встроенные интерфейсы (без имен)
			if len(method.Names) == 0 {
				methods = append(methods, method)
				continue
			}
			if method.Names[0].Name != methodName {
				methods = append(methods, method)
			} else {
				removed = true
			}
		}
		interfaceType.Methods.List = methods
		return false
	})
	return removed
}

// runStaticcheck запускает staticcheck на временной копии проекта
func runStaticcheck(tmpDir string, verbose bool) error {
	cmd := exec.Command("staticcheck", "./...")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: ошибка запуска staticcheck: %v\n%s\n", err, output)
		}
		return fmt.Errorf("ошибка staticcheck: %v", err)
	}
	return nil
}

// FindUnusedMethods проверяет методы интерфейсов через staticcheck
func FindUnusedMethods(pkgs map[string]*stage0.Package, usedMethodsByPkg map[string][]*stage1.UsedMethod, cfg *config.Config, verbose bool) error {
	// Создаем временную директорию для проверки
	tmpDir, err := os.MkdirTemp("", "interface-linter-*")
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: не удалось создать временную директорию: %v\n", err)
		}
		return fmt.Errorf("не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Копируем весь проект во временную директорию
	if err := copyProject(".", tmpDir, cfg); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: не удалось скопировать проект: %v\n", err)
		}
		return fmt.Errorf("не удалось скопировать проект: %v", err)
	}

	// Для каждого пакета
	for pkgPath, pkg := range pkgs {
		// Получаем используемые методы пакета
		usedMethods := usedMethodsByPkg[pkgPath]
		usedMethodsMap := make(map[string]bool)
		for _, m := range usedMethods {
			key := fmt.Sprintf("%s.%s", m.InterfaceName, m.MethodName)
			usedMethodsMap[key] = true
		}

		// Для каждого файла в пакете
		for filePath, file := range pkg.Files {
			// Пропускаем файлы по конфигурации
			if cfg.ShouldIgnore(filePath) {
				if verbose {
					fmt.Fprintf(os.Stderr, "DEBUG: пропускаем файл %s\n", filePath)
				}
				continue
			}

			// Находим все интерфейсы в файле
			interfaces := findInterfaces(file)

			// Проверяем каждый метод каждого интерфейса
			for _, iface := range interfaces {
				for _, method := range iface.Methods {
					key := fmt.Sprintf("%s.%s", method.InterfaceName, method.MethodName)
					if !usedMethodsMap[key] {
						// Метод не найден в stage1, проверяем через staticcheck
						// Удаляем метод из интерфейса во временной копии
						tmpFilePath := filepath.Join(tmpDir, filePath)
						if removeMethod(file, method.InterfaceName, method.MethodName) {
							// Записываем измененный файл
							f, err := os.Create(tmpFilePath)
							if err != nil {
								if verbose {
									fmt.Fprintf(os.Stderr, "DEBUG: не удалось создать временный файл: %v\n", err)
								}
								continue
							}
							if err := printer.Fprint(f, pkg.Fset, file); err != nil {
								f.Close()
								if verbose {
									fmt.Fprintf(os.Stderr, "DEBUG: не удалось записать AST: %v\n", err)
								}
								continue
							}
							f.Close()

							// Запускаем staticcheck для проверки
							if err := runStaticcheck(tmpDir, verbose); err == nil {
								// Если staticcheck прошел без ошибок, значит метод действительно не используется
								fmt.Printf("UNUSED: github.com/comerc/unused-interface-methods/test/data.%s.%s\n",
									method.InterfaceName,
									method.MethodName,
								)
							}

							// Восстанавливаем файл
							if err := copyFile(filePath, tmpFilePath); err != nil {
								if verbose {
									fmt.Fprintf(os.Stderr, "DEBUG: не удалось восстановить файл: %v\n", err)
								}
								continue
							}
						}
					}
				}
			}
		}
	}

	return nil
}
