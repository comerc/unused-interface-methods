package linter

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// InterfaceMethod представляет метод интерфейса
type InterfaceMethod struct {
	InterfaceName string
	MethodName    string
	Signature     string
	File          string
	Line          int
	Interface     *types.Interface // Добавляем информацию о типе интерфейса
}

// GenericWarning представляет предупреждение о дженерике
type GenericWarning struct {
	InterfaceName string
	File          string
	Line          int
	MethodCount   int
	TypeParams    string
}

// ConfigInterface определяет интерфейс для конфигурации
type ConfigInterface interface {
	ShouldIgnore(filePath string) bool
}

// UnusedMethodLinter анализирует Go-код на предмет неиспользуемых методов в интерфейсах
type UnusedMethodLinter struct {
	packages        []*packages.Package
	methods         []InterfaceMethod
	genericWarnings []GenericWarning
	verbose         bool
	config          ConfigInterface
}

func New(config ConfigInterface, verbose bool) *UnusedMethodLinter {
	return &UnusedMethodLinter{
		methods:         make([]InterfaceMethod, 0),
		genericWarnings: make([]GenericWarning, 0),
		verbose:         verbose,
		config:          config,
	}
}

// LoadPackages загружает пакеты с информацией о типах
func (l *UnusedMethodLinter) LoadPackages(dir string) error {
	// Настройка загрузки пакетов
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedModule,
		Dir: dir,
	}

	// Загружаем пакеты рекурсивно
	pattern := "./..."
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	// Фильтруем пакеты, исключая ненужные директории
	var filteredPkgs []*packages.Package
	for _, pkg := range pkgs {
		shouldIgnore := false
		for _, dir := range pkg.GoFiles {
			if l.config.ShouldIgnore(filepath.Dir(dir)) {
				shouldIgnore = true
				break
			}
		}

		if !shouldIgnore {
			// Проверяем ошибки в пакетах
			if len(pkg.Errors) > 0 {
				for _, err := range pkg.Errors {
					if l.verbose {
						fmt.Printf("Warning: %v\n", err)
					}

				}
			}
			filteredPkgs = append(filteredPkgs, pkg)
		} else if l.verbose {
			fmt.Printf("Excluding package: %s\n", pkg.PkgPath)
		}
	}

	l.packages = filteredPkgs
	return nil
}

// ExtractInterfaceMethods извлекает все методы интерфейсов
func (l *UnusedMethodLinter) ExtractInterfaceMethods() {
	for _, pkg := range l.packages {
		for _, file := range pkg.Syntax {
			if l.shouldSkipFile(pkg, file) {
				continue
			}

			filename := pkg.Fset.Position(file.Pos()).Filename
			if l.verbose {
				fmt.Printf("  Analyzing: %s\n", getRelativePath(filename))
			}

			l.ExtractInterfaceMethodsFromFile(pkg, file, filename)
		}
	}
}

// ExtractInterfaceMethodsFromFile извлекает методы интерфейсов из файла
func (l *UnusedMethodLinter) ExtractInterfaceMethodsFromFile(pkg *packages.Package, file *ast.File, filename string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if interfaceType, ok := x.Type.(*ast.InterfaceType); ok {
				// НОВАЯ ПРОВЕРКА: детектируем дженерики
				if l.isGenericInterface(x) {
					l.addGenericWarning(x, interfaceType, filename, pkg.Fset)
					return true // Пропускаем анализ дженерик-интерфейсов
				}

				// Получаем информацию о типе интерфейса
				if obj := pkg.TypesInfo.Defs[x.Name]; obj != nil {
					if namedType, ok := obj.Type().(*types.Named); ok {
						if iface, ok := namedType.Underlying().(*types.Interface); ok {
							l.extractMethodsFromInterface(pkg, x.Name.Name, interfaceType, iface, filename)
						}
					}
				}
			}
		}
		return true
	})
}

// isGenericInterface проверяет, является ли интерфейс дженериком
func (l *UnusedMethodLinter) isGenericInterface(typeSpec *ast.TypeSpec) bool {
	return typeSpec.TypeParams != nil && len(typeSpec.TypeParams.List) > 0
}

// addGenericWarning добавляет предупреждение о дженерик-интерфейсе
func (l *UnusedMethodLinter) addGenericWarning(typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType, filename string, fset *token.FileSet) {
	position := fset.Position(typeSpec.Pos())

	// Подсчитываем количество методов
	methodCount := 0
	for _, method := range interfaceType.Methods.List {
		if len(method.Names) > 0 {
			methodCount += len(method.Names)
		}
	}

	// Извлекаем информацию о типовых параметрах
	typeParams := l.getTypeParamsString(typeSpec.TypeParams)

	warning := GenericWarning{
		InterfaceName: typeSpec.Name.Name,
		File:          filename,
		Line:          position.Line,
		MethodCount:   methodCount,
		TypeParams:    typeParams,
	}

	l.genericWarnings = append(l.genericWarnings, warning)

	if l.verbose {
		fmt.Printf("⚠️  WARNING: Skipping generic interface '%s%s' at %s:%d (%d methods)\n",
			warning.InterfaceName, warning.TypeParams, getRelativePath(filename), warning.Line, warning.MethodCount)
	}
}

// getTypeParamsString извлекает строковое представление типовых параметров
func (l *UnusedMethodLinter) getTypeParamsString(typeParams *ast.FieldList) string {
	if typeParams == nil || len(typeParams.List) == 0 {
		return ""
	}

	var params []string
	for _, field := range typeParams.List {
		for _, name := range field.Names {
			paramStr := name.Name
			if field.Type != nil {
				paramStr += " " + l.typeToString(field.Type)
			}
			params = append(params, paramStr)
		}
	}

	return "[" + strings.Join(params, ", ") + "]"
}

// extractMethodsFromInterface извлекает методы из интерфейса
func (l *UnusedMethodLinter) extractMethodsFromInterface(pkg *packages.Package, interfaceName string,
	interfaceAST *ast.InterfaceType, interfaceType *types.Interface, filename string) {

	for _, method := range interfaceAST.Methods.List {
		if len(method.Names) == 0 {
			continue // Встроенный интерфейс
		}

		for _, name := range method.Names {
			position := pkg.Fset.Position(name.Pos())
			signature := l.getMethodSignature(method)

			l.methods = append(l.methods, InterfaceMethod{
				InterfaceName: interfaceName,
				MethodName:    name.Name,
				Signature:     signature,
				File:          filename,
				Line:          position.Line,
				Interface:     interfaceType,
			})
		}
	}
}

// getMethodSignature получает сигнатуру метода в виде строки
func (l *UnusedMethodLinter) getMethodSignature(method *ast.Field) string {
	if funcType, ok := method.Type.(*ast.FuncType); ok {
		var parts []string

		// Параметры
		if funcType.Params != nil {
			var params []string
			for _, param := range funcType.Params.List {
				paramType := l.typeToString(param.Type)
				if len(param.Names) > 0 {
					for _, name := range param.Names {
						params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
					}
				} else {
					params = append(params, paramType)
				}
			}
			parts = append(parts, fmt.Sprintf("(%s)", strings.Join(params, ", ")))
		} else {
			parts = append(parts, "()")
		}

		// Возвращаемые значения
		if funcType.Results != nil {
			var results []string
			for _, result := range funcType.Results.List {
				resultType := l.typeToString(result.Type)
				results = append(results, resultType)
			}
			if len(results) == 1 {
				parts = append(parts, results[0])
			} else {
				parts = append(parts, fmt.Sprintf("(%s)", strings.Join(results, ", ")))
			}
		}

		return strings.Join(parts, " ")
	}
	return ""
}

// typeToString преобразует AST-тип в строку
func (l *UnusedMethodLinter) typeToString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + l.typeToString(x.X)
	case *ast.SelectorExpr:
		return l.typeToString(x.X) + "." + x.Sel.Name
	case *ast.ArrayType:
		return "[]" + l.typeToString(x.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", l.typeToString(x.Key), l.typeToString(x.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		return "chan " + l.typeToString(x.Value)
	default:
		return "unknown"
	}
}

// FindUnusedMethods находит неиспользуемые методы
func (l *UnusedMethodLinter) FindUnusedMethods() bool {
	fmt.Println("DEBUG: Starting FindUnusedMethods")

	// Группируем методы по интерфейсам
	interfaceMap := make(map[string][]InterfaceMethod)
	for _, method := range l.methods {
		interfaceMap[method.InterfaceName] = append(interfaceMap[method.InterfaceName], method)
	}

	fmt.Printf("DEBUG: Found %d interfaces to check\n", len(interfaceMap))

	unusedCount := 0
	usedCount := 0
	skippedMethodsCount := 0

	interfaceNum := 0
	for interfaceName, methods := range interfaceMap {
		interfaceNum++
		fmt.Printf("DEBUG: Checking interface %d/%d: %s (%d methods)\n",
			interfaceNum, len(interfaceMap), interfaceName, len(methods))

		if l.verbose {
			fmt.Printf("Interface: %s\n", interfaceName)
		}

		methodNum := 0
		for _, method := range methods {
			methodNum++
			fmt.Printf("DEBUG: Checking method %d/%d: %s.%s\n",
				methodNum, len(methods), interfaceName, method.MethodName)

			used := l.isMethodUsed(method)
			if !used {
				fmt.Printf("UNUSED: %s.%s%s (%s:%d)\n",
					method.InterfaceName, method.MethodName, method.Signature,
					getRelativePath(method.File), method.Line)
				unusedCount++
			} else {
				if l.verbose {
					fmt.Printf("  USED: %s%s\n", method.MethodName, method.Signature)
				}
				usedCount++
			}
		}

	}

	// Подсчитываем пропущенные методы из дженериков
	for _, warning := range l.genericWarnings {
		skippedMethodsCount += warning.MethodCount
	}

	// Выводим предупреждения о дженериках
	l.printGenericWarnings()

	// Итоговая статистика
	fmt.Printf("\nDEBUG: Final stats - %d used, %d unused, %d total", usedCount, unusedCount, len(l.methods))
	if skippedMethodsCount > 0 {
		fmt.Printf(" (%d methods skipped due to generics)", skippedMethodsCount)
	}
	fmt.Println()

	return unusedCount == 0
}

// printGenericWarnings выводит предупреждения о дженериках
func (l *UnusedMethodLinter) printGenericWarnings() {
	if len(l.genericWarnings) == 0 {
		return
	}

	if !l.verbose {
		// Краткий режим - только общая статистика
		totalSkipped := 0
		for _, warning := range l.genericWarnings {
			totalSkipped += warning.MethodCount
		}

		if totalSkipped > 0 {
			fmt.Printf("\n⚠️  %d generic interface%s skipped (%d methods not analyzed)\n",
				len(l.genericWarnings),
				func() string {
					if len(l.genericWarnings) == 1 {
						return ""
					} else {
						return "s"
					}
				}(),
				totalSkipped)
		}
	} else {
		// Подробный режим - детальные предупреждения
		fmt.Println("\n⚠️  Generic Interface Warnings:")
		for _, warning := range l.genericWarnings {
			fmt.Printf("  - '%s%s' at %s:%d (%d methods skipped)\n",
				warning.InterfaceName, warning.TypeParams, getRelativePath(warning.File), warning.Line, warning.MethodCount)
		}
	}
}

// isMethodUsed проверяет, используется ли метод в коде с учетом типов
func (l *UnusedMethodLinter) isMethodUsed(method InterfaceMethod) bool {
	if l.verbose {
		fmt.Printf("    Checking usage of: %s.%s\n", method.InterfaceName, method.MethodName)
	}

	for _, pkg := range l.packages {
		for _, file := range pkg.Syntax {
			if l.shouldSkipFile(pkg, file) {
				continue
			}

			filename := pkg.Fset.Position(file.Pos()).Filename
			if l.verbose {
				fmt.Printf("      Checking file: %s\n", getRelativePath(filename))
			}

			if l.checkMethodUsageWithTypes(pkg, file, method) {
				if l.verbose {
					fmt.Printf("        Found usage in %s\n", getRelativePath(filename))
				}
				return true
			}
		}
	}

	if l.verbose {
		fmt.Printf("      No usage found\n")
	}
	return false
}

// checkMethodUsageWithTypes проверяет использование метода с учетом информации о типах
func (l *UnusedMethodLinter) checkMethodUsageWithTypes(pkg *packages.Package, file *ast.File, method InterfaceMethod) bool {
	if l.verbose {
		fmt.Printf("      DEBUG: Checking method usage for %s.%s in file %s\n",
			method.InterfaceName, method.MethodName, pkg.Fset.Position(file.Pos()).Filename)
	}

	found := false

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			// Проверяем вызовы методов (obj.method())
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == method.MethodName {
					if l.verbose {
						fmt.Printf("        DEBUG: Found method call %s\n", sel.Sel.Name)
					}
					// Получаем тип объекта, на котором вызывается метод
					if l.isMethodCallOnInterface(pkg, sel, method) {
						found = true
						return false
					}
				}
			}
		case *ast.SelectorExpr:
			// Проверяем обращения к методу без вызова (obj.method)
			if x.Sel.Name == method.MethodName {
				if l.verbose {
					fmt.Printf("        DEBUG: Found selector expression %s\n", x.Sel.Name)
				}
				// Проверяем, что это не поле структуры
				if ident, ok := x.X.(*ast.Ident); ok {
					if l.verbose {
						fmt.Printf("        DEBUG: Found identifier %s\n", ident.Name)
					}
					// Получаем тип идентификатора
					if obj := pkg.TypesInfo.ObjectOf(ident); obj != nil {
						if l.verbose {
							fmt.Printf("        DEBUG: Found object of type %T\n", obj)
						}
						// Проверяем, что это не поле структуры
						if _, ok := obj.(*types.Var); !ok {
							if l.isMethodCallOnInterface(pkg, x, method) {
								found = true
								return false
							}
						} else {
							if l.verbose {
								fmt.Printf("        DEBUG: Skipping field access\n")
							}
						}
					}
				} else {
					if l.verbose {
						fmt.Printf("        DEBUG: Non-identifier selector base: %T\n", x.X)
					}
					if l.isMethodCallOnInterface(pkg, x, method) {
						found = true
						return false
					}
				}
			}
		case *ast.Field:
			// Проверяем поля структур
			if ident, ok := x.Type.(*ast.Ident); ok {
				if l.verbose {
					fmt.Printf("        DEBUG: Found field with type %s\n", ident.Name)
				}
				// Проверяем, что это поле с типом нашего интерфейса
				if ident.Name == method.InterfaceName {
					if l.verbose {
						fmt.Printf("        DEBUG: Field type matches interface name\n")
					}
					// Получаем тип поля
					if obj := pkg.TypesInfo.ObjectOf(ident); obj != nil {
						if l.verbose {
							fmt.Printf("        DEBUG: Found field type object: %T\n", obj)
						}
						// Проверяем, что это тип
						if _, ok := obj.(*types.TypeName); ok {
							if l.verbose {
								fmt.Printf("        DEBUG: Field type is a type name\n")
							}
							// Проверяем, что это наш интерфейс
							if named, ok := obj.Type().(*types.Named); ok {
								if iface, ok := named.Underlying().(*types.Interface); ok {
									if l.verbose {
										fmt.Printf("        DEBUG: Found interface with %d methods\n", iface.NumMethods())
									}
									// Проверяем, что это именно тот интерфейс, который мы ищем
									if types.Identical(method.Interface, iface) {
										if l.verbose {
											fmt.Printf("        DEBUG: Interface types are identical\n")
										}
										// Проверяем, что метод используется
										for i := 0; i < iface.NumMethods(); i++ {
											if iface.Method(i).Name() == method.MethodName {
												if l.verbose {
													fmt.Printf("        DEBUG: Found method in interface\n")
												}
												found = true
												return false
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return !found
	})

	if l.verbose {
		if found {
			fmt.Printf("      DEBUG: Method usage found\n")
		} else {
			fmt.Printf("      DEBUG: Method usage not found\n")
		}
	}

	return found
}

// isMethodCallOnInterface проверяет, вызывается ли метод на нужном интерфейсе
func (l *UnusedMethodLinter) isMethodCallOnInterface(pkg *packages.Package, sel *ast.SelectorExpr, method InterfaceMethod) bool {
	if l.verbose {
		fmt.Printf("        DEBUG: Checking method call %s.%s\n", method.InterfaceName, method.MethodName)
	}

	// Получаем тип выражения слева от селектора
	exprType := pkg.TypesInfo.TypeOf(sel.X)
	if exprType == nil {
		if l.verbose {
			fmt.Printf("        DEBUG: No type info for expression\n")
		}
		return false
	}

	if l.verbose {
		fmt.Printf("        DEBUG: Expression type: %v\n", exprType.String())
	}

	// Убираем именованные типы
	if named, ok := exprType.(*types.Named); ok {
		if l.verbose {
			fmt.Printf("        DEBUG: Found named type: %s\n", named.Obj().Name())
		}
		// Проверяем, что это именно тот интерфейс, который мы ищем
		if named.Obj().Name() == method.InterfaceName {
			if l.verbose {
				fmt.Printf("        DEBUG: Direct match with interface name\n")
			}
			return true
		}
		exprType = named.Underlying()
	}

	// Проверяем, является ли тип интерфейсом
	iface, ok := exprType.(*types.Interface)
	if !ok {
		if l.verbose {
			fmt.Printf("        DEBUG: Not an interface type: %T\n", exprType)
		}
		return false
	}

	if l.verbose {
		fmt.Printf("        DEBUG: Found interface with %d methods\n", iface.NumMethods())
	}

	// Проверяем, что это именно тот интерфейс, который мы ищем
	if types.Identical(method.Interface, iface) {
		if l.verbose {
			fmt.Printf("        DEBUG: Interface types are identical\n")
		}
		return true
	}

	// Проверяем, содержит ли интерфейс нужный метод с той же сигнатурой
	for i := 0; i < iface.NumMethods(); i++ {
		ifaceMethod := iface.Method(i)
		if ifaceMethod.Name() == method.MethodName {
			// Находим метод в исходном интерфейсе
			for j := 0; j < method.Interface.NumMethods(); j++ {
				origMethod := method.Interface.Method(j)
				if origMethod.Name() == method.MethodName {
					// Сравниваем сигнатуры методов
					if types.Identical(ifaceMethod.Type(), origMethod.Type()) {
						if l.verbose {
							fmt.Printf("        DEBUG: Found matching method with identical signature\n")
						}
						return true
					}
				}
			}
		}
	}

	if l.verbose {
		fmt.Printf("        DEBUG: Interface types are different\n")
	}
	return false
}

// isSameInterface проверяет, являются ли два интерфейса одинаковыми
func (l *UnusedMethodLinter) isSameInterface(iface1, iface2 *types.Interface) bool {
	// Если количество методов разное, интерфейсы точно разные
	if iface1.NumMethods() != iface2.NumMethods() {
		return false
	}

	// Проверяем каждый метод из первого интерфейса
	for i := 0; i < iface1.NumMethods(); i++ {
		method1 := iface1.Method(i)
		found := false

		// Ищем метод с таким же именем во втором интерфейсе
		for j := 0; j < iface2.NumMethods(); j++ {
			method2 := iface2.Method(j)
			if method1.Name() == method2.Name() {
				// Нашли метод с таким же именем, проверяем сигнатуру
				sig1 := method1.Type().(*types.Signature)
				sig2 := method2.Type().(*types.Signature)

				// Сравниваем сигнатуры
				if types.Identical(sig1, sig2) {
					found = true
					break
				}
			}
		}

		// Если не нашли метод с таким же именем и сигнатурой, интерфейсы разные
		if !found {
			return false
		}
	}

	return true
}

// getRelativePath преобразует абсолютный путь в относительный от текущей директории
func getRelativePath(filePath string) string {
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

// shouldSkipFile проверяет, нужно ли пропустить файл при анализе
func (l *UnusedMethodLinter) shouldSkipFile(pkg *packages.Package, file *ast.File) bool {
	// Пропускаем тестовые пакеты
	if strings.HasSuffix(pkg.PkgPath, "_test") {
		return true
	}

	// Пропускаем файлы по конфигурации
	filename := pkg.Fset.Position(file.Pos()).Filename
	return l.config.ShouldIgnore(filename)
}
