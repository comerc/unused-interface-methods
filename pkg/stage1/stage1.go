package stage1

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/types"
	"os"

	"github.com/comerc/unused-interface-methods/pkg/config"
	"github.com/comerc/unused-interface-methods/pkg/stage0"
)

// UsedMethod представляет используемый метод интерфейса
type UsedMethod struct {
	PkgPath       string           // путь к пакету
	InterfaceName string           // имя интерфейса
	MethodName    string           // имя метода
	Signature     *types.Signature // сигнатура для точного определения
}

// checkMethodUsage проверяет использование метода интерфейса в вызове o.i.Method()
func checkMethodUsage(info *types.Info, n ast.Node, verbose bool) (*UsedMethod, bool) {
	// Проверяем что это вызов функции
	call, ok := n.(*ast.CallExpr)
	if !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: узел не является вызовом функции: %T\n", n)
		}
		return nil, false
	}

	// Проверяем что вызывается метод: foo.Method()
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: вызов не является обращением к методу: %T\n", call.Fun)
		}
		return nil, false
	}

	// Получаем тип поля (i)
	fieldType := info.TypeOf(sel.X)
	if fieldType == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: не удалось определить тип для %v\n", sel.X)
		}
		return nil, false
	}

	// Проверяем что это интерфейс
	iface, ok := fieldType.Underlying().(*types.Interface)
	if !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: тип %v не является интерфейсом\n", fieldType)
		}
		return nil, false
	}

	// Получаем имя интерфейса
	named, ok := fieldType.(*types.Named)
	if !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: тип %v не является именованным типом\n", fieldType)
		}
		return nil, false
	}
	interfaceName := named.Obj().Name()

	// Получаем тип метода
	methodType := info.TypeOf(sel.Sel)
	if methodType == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: не удалось определить тип метода %s\n", sel.Sel.Name)
		}
		return nil, false
	}

	// Получаем сигнатуру метода
	signature, ok := methodType.(*types.Signature)
	if !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: тип %v не является сигнатурой метода\n", methodType)
		}
		return nil, false
	}

	// Проверяем что метод действительно принадлежит интерфейсу
	for i := 0; i < iface.NumMethods(); i++ {
		method := iface.Method(i)
		if method.Name() == sel.Sel.Name {
			// Проверяем сигнатуру
			if types.Identical(method.Type(), signature) {
				return &UsedMethod{
					InterfaceName: interfaceName,
					MethodName:    sel.Sel.Name,
					Signature:     signature,
				}, true
			}
		}
	}

	return nil, false
}

// FindUsedMethods находит все точно используемые методы интерфейсов в пакете
func FindUsedMethods(pkgs map[string]*stage0.Package, cfg *config.Config, verbose bool) (map[string][]*UsedMethod, error) {
	usedMethods := make(map[string][]*UsedMethod) // pkgPath -> usedMethods

	for pkgPath, pkg := range pkgs {
		if verbose {
			fmt.Fprintf(os.Stderr, "DEBUG: анализ пакета %s\n", pkgPath)
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

		// Создаем info для анализа типов
		info := &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}

		// Собираем все файлы пакета в слайс для анализа
		var files []*ast.File
		for _, file := range pkg.Files {
			files = append(files, file)
		}

		// Анализируем типы пакета
		_, err := conf.Check(pkgPath, pkg.Fset, files, info)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "DEBUG: ошибка проверки типов в пакете %s: %v\n", pkgPath, err)
			}
			continue // пропускаем пакет с ошибками
		}

		// Анализируем каждый файл в пакете
		for filePath, file := range pkg.Files {
			// Пропускаем файлы по конфигурации
			if cfg.ShouldIgnore(filePath) {
				if verbose {
					fmt.Fprintf(os.Stderr, "DEBUG: пропускаем файл %s\n", filePath)
				}
				continue
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "DEBUG: анализ файла %s\n", filePath)
			}

			ast.Inspect(file, func(n ast.Node) bool {
				method, ok := checkMethodUsage(info, n, verbose)
				if !ok {
					return true
				}

				method.PkgPath = pkgPath

				// Выводим результат в правильном формате
				fmt.Printf("OK: github.com/comerc/unused-interface-methods/test/data.%s.%s\n",
					method.InterfaceName,
					method.MethodName,
				)

				usedMethods[pkgPath] = append(usedMethods[pkgPath], method)
				return true
			})
		}
	}

	return usedMethods, nil
}
