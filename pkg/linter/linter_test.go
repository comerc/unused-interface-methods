package linter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"

	"github.com/comerc/unused-interface-methods/pkg/config"
)

func TestUnusedMethodLinter(t *testing.T) {
	// Создаем и настраиваем линтер
	linter := &UnusedMethodLinter{
		methods: make([]InterfaceMethod, 0),
		verbose: true,
		config:  config.DefaultConfig(),
	}

	// Загружаем и анализируем пакеты
	err := linter.LoadPackages("../../test/data")
	assert.NoError(t, err, "LoadPackages() failed")

	// Извлекаем методы
	linter.ExtractInterfaceMethods()

	// Проверяем методы из interfaces.go
	t.Run("interfaces.go", func(t *testing.T) {
		// Проверяем методы интерфейса DirectProcessor
		directProcessor := findInterface(linter.methods, "DirectProcessor")
		assert.NotNil(t, directProcessor, "DirectProcessor interface not found")

		// Проверяем что ProcessDirect используется в DirectProcessor
		processDirect := findMethod(directProcessor, "ProcessDirect")
		assert.NotNil(t, processDirect, "ProcessDirect method not found in DirectProcessor")
		if processDirect != nil {
			isUsed := linter.isMethodUsed(*processDirect)
			assert.True(t, isUsed, "DirectProcessor.ProcessDirect should be used")
		}
	})
}

// Вспомогательные функции для тестов
func findInterface(methods []InterfaceMethod, name string) *types.Interface {
	for _, method := range methods {
		if method.InterfaceName == name {
			return method.Interface
		}
	}
	return nil
}

func findMethod(iface *types.Interface, name string) *InterfaceMethod {
	if iface == nil {
		return nil
	}
	for i := 0; i < iface.NumMethods(); i++ {
		if iface.Method(i).Name() == name {
			return &InterfaceMethod{
				InterfaceName: name,
				MethodName:    name,
				Interface:     iface,
			}
		}
	}
	return nil
}

// TestIsSameInterface проверяет корректность сравнения интерфейсов
func TestIsSameInterface(t *testing.T) {
	linter := &UnusedMethodLinter{}

	// Создаем тестовые интерфейсы
	makeMethod := func(name string) *types.Func {
		return types.NewFunc(token.NoPos, nil, name, types.NewSignature(nil, nil, nil, false))
	}

	tests := []struct {
		name     string
		iface1   *types.Interface
		iface2   *types.Interface
		expected bool
	}{
		{
			name:     "empty interfaces",
			iface1:   types.NewInterface(nil, nil),
			iface2:   types.NewInterface(nil, nil),
			expected: true,
		},
		{
			name: "same methods",
			iface1: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
				makeMethod("Method2"),
			}, nil),
			iface2: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
				makeMethod("Method2"),
			}, nil),
			expected: true,
		},
		{
			name: "different order",
			iface1: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
				makeMethod("Method2"),
			}, nil),
			iface2: types.NewInterface([]*types.Func{
				makeMethod("Method2"),
				makeMethod("Method1"),
			}, nil),
			expected: true,
		},
		{
			name: "different methods",
			iface1: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
				makeMethod("Method2"),
			}, nil),
			iface2: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
				makeMethod("Method3"),
			}, nil),
			expected: false,
		},
		{
			name: "different count",
			iface1: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
				makeMethod("Method2"),
			}, nil),
			iface2: types.NewInterface([]*types.Func{
				makeMethod("Method1"),
			}, nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linter.isSameInterface(tt.iface1, tt.iface2)
			assert.Equal(t, tt.expected, result, "unexpected result for test case %s", tt.name)
		})
	}
}

func TestFindUnusedMethods(t *testing.T) {
	tests := []struct {
		name             string
		verbose          bool
		expectedCode     int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:         "verbose mode",
			verbose:      true,
			expectedCode: 0,
			shouldContain: []string{
				"UNUSED:",
				"\n  USED:",
				"WARNING:",
				"Analyzing:",
				"Skipping generic interface",
			},
			shouldNotContain: []string{},
		},
		{
			name:         "non-verbose mode",
			verbose:      false,
			expectedCode: 0,
			shouldContain: []string{
				"UNUSED:",
			},
			shouldNotContain: []string{
				"\n  USED:",
				"WARNING:",
				"Analyzing:",
				"Skipping generic interface",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл для вывода
			tmpfile, err := os.CreateTemp("", "test_output")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Создаем и настраиваем линтер
			linter := &UnusedMethodLinter{
				methods:         make([]InterfaceMethod, 0),
				genericWarnings: make([]GenericWarning, 0),
				verbose:         tt.verbose,
				config:          config.DefaultConfig(),
			}

			// Загружаем и анализируем пакеты
			err = linter.LoadPackages("../../test/data")
			assert.NoError(t, err, "LoadPackages() failed")

			// Перехватываем stdout
			oldStdout := os.Stdout
			os.Stdout = tmpfile
			defer func() {
				os.Stdout = oldStdout
			}()

			// Извлекаем методы
			linter.ExtractInterfaceMethods()

			// Подменяем os.Exit на мок
			var exitCode int
			oldOsExit := config.OsExit
			config.OsExit = func(code int) {
				exitCode = code
			}
			defer func() {
				config.OsExit = oldOsExit
			}()

			// Ищем неиспользуемые методы
			linter.FindUnusedMethods()

			// Закрываем файл для записи и открываем для чтения
			tmpfile.Close()
			content, err := os.ReadFile(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			output := string(content)

			// Проверяем код выхода
			assert.Equal(t, tt.expectedCode, exitCode, "unexpected exit code")

			// Проверяем наличие ожидаемых строк
			for _, str := range tt.shouldContain {
				assert.Contains(t, output, str, "missing expected output: %s", str)
			}

			// Проверяем отсутствие нежелательных строк
			for _, str := range tt.shouldNotContain {
				assert.NotContains(t, output, str, "unexpected output found: %s", str)
			}
		})
	}
}

func TestIsMethodCallOnInterface_PointerHandling(t *testing.T) {
	// Загружаем тестовые данные
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "../../test/data/interfaces.go", nil, 0)
	assert.NoError(t, err)

	// Создаем информацию о типах
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	// Настраиваем конфигурацию с импортером
	conf := &types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("../../test/data", fset, []*ast.File{f}, info)
	assert.NoError(t, err)

	// Создаем линтер
	linter := &UnusedMethodLinter{
		methods:         make([]InterfaceMethod, 0),
		genericWarnings: make([]GenericWarning, 0),
		verbose:         true,
		config:          config.DefaultConfig(),
	}

	// Создаем интерфейс для проверки
	pointerIface := pkg.Scope().Lookup("PointerHandler").Type().Underlying().(*types.Interface)

	pointerMethod := InterfaceMethod{
		InterfaceName: "PointerHandler",
		MethodName:    "HandlePointer",
		Interface:     pointerIface,
	}

	// Находим все вызовы методов HandlePointer
	var interfaceCalls []*ast.SelectorExpr
	var directCalls []*ast.SelectorExpr

	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok && s.Sel.Name == "HandlePointer" {
			// Проверяем, является ли это прямым вызовом на структуре
			if ident, ok := s.X.(*ast.Ident); ok && ident.Name == "s" {
				directCalls = append(directCalls, s)
			} else {
				interfaceCalls = append(interfaceCalls, s)
			}
		}
		return true
	})

	// Создаем тестовый пакет для packages.Package
	testPkg := &packages.Package{
		Types:     pkg,
		TypesInfo: info,
	}

	// Проверяем вызовы через интерфейс
	fmt.Printf("\nTesting interface HandlePointer calls:\n")
	for i, sel := range interfaceCalls {
		result := linter.isMethodCallOnInterface(testPkg, sel, pointerMethod)
		assert.True(t, result, "HandlePointer call %d should be considered an interface call", i)
	}

	// Проверяем прямые вызовы
	fmt.Printf("\nTesting direct HandlePointer calls:\n")
	for i, sel := range directCalls {
		result := linter.isMethodCallOnInterface(testPkg, sel, pointerMethod)
		assert.False(t, result, "Direct HandlePointer call %d should not be considered an interface call", i)
	}
}

// TestGetTypeParamsString_Empty проверяет обработку пустых параметров типа
func TestGetTypeParamsString_Empty(t *testing.T) {
	linter := &UnusedMethodLinter{}

	// Проверяем nil
	assert.Empty(t, linter.getTypeParamsString(nil), "should return empty string for nil params")

	// Проверяем пустой список
	emptyList := &ast.FieldList{List: []*ast.Field{}}
	assert.Empty(t, linter.getTypeParamsString(emptyList), "should return empty string for empty list")
}

// TestIsMethodCallOnInterface_NoTypeInfo проверяет поведение функции
// isMethodCallOnInterface когда TypesInfo не содержит информацию о типе
func TestIsMethodCallOnInterface_NoTypeInfo(t *testing.T) {
	// Перехватываем stdout для проверки отладочного сообщения
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Создаем линтер
	linter := &UnusedMethodLinter{
		methods:         make([]InterfaceMethod, 0),
		genericWarnings: make([]GenericWarning, 0),
		verbose:         true,
		config:          config.DefaultConfig(),
	}

	// Создаем пустой интерфейс для теста
	iface := types.NewInterface(nil, nil)

	method := InterfaceMethod{
		InterfaceName: "TestInterface",
		MethodName:    "TestMethod",
		Interface:     iface,
	}

	// Создаем пустой TypesInfo (без информации о типах)
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	// Создаем тестовый пакет
	testPkg := &packages.Package{
		TypesInfo: info,
	}

	// Создаем селектор с неразрешенным идентификатором
	sel := &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "undefinedVar", // Неразрешенный идентификатор
		},
		Sel: &ast.Ident{
			Name: "TestMethod",
		},
	}

	// Проверяем, что функция возвращает false и выводит отладочное сообщение
	result := linter.isMethodCallOnInterface(testPkg, sel, method)

	// Закрываем pipe и восстанавливаем stdout
	w.Close()
	output := make([]byte, 1024)
	_, _ = r.Read(output)
	os.Stdout = oldStdout

	// Проверяем результат
	assert.False(t, result, "Method call with no type info should return false")
}

// TestShouldSkipFile проверяет корректность определения пропускаемых файлов
func TestShouldSkipFile(t *testing.T) {
	// Создаем FileSet для позиций
	fset := token.NewFileSet()
	mainFileNode := fset.AddFile("main.go", -1, 1000)
	mainFileNode.SetLines([]int{0, 100})
	testFileNode := fset.AddFile("main_test.go", -1, 1000)
	testFileNode.SetLines([]int{0, 100})
	mockFileNode := fset.AddFile("mock/mock.go", -1, 1000)
	mockFileNode.SetLines([]int{0, 100})

	// Создаем тестовые файлы
	mainFile := &ast.File{
		Name:    &ast.Ident{Name: "main.go"},
		Package: mainFileNode.Pos(1),
	}

	testFile := &ast.File{
		Name:    &ast.Ident{Name: "main_test.go"},
		Package: testFileNode.Pos(1),
	}

	mockFile := &ast.File{
		Name:    &ast.Ident{Name: "mock.go"},
		Package: mockFileNode.Pos(1),
	}

	// Создаем тестовые пакеты
	mainPkg := &packages.Package{
		PkgPath: "example.com/myapp",
		Fset:    fset,
	}

	testPkg := &packages.Package{
		PkgPath: "example.com/myapp_test",
		Fset:    fset,
	}

	// Создаем линтер
	linter := &UnusedMethodLinter{
		verbose: true,
		config:  config.DefaultConfig(),
	}

	// Проверяем различные случаи
	tests := []struct {
		name     string
		pkg      *packages.Package
		file     *ast.File
		expected bool
	}{
		{
			name:     "regular file",
			pkg:      mainPkg,
			file:     mainFile,
			expected: false,
		},
		{
			name:     "test package",
			pkg:      testPkg,
			file:     testFile,
			expected: true,
		},
		{
			name:     "file in ignored directory",
			pkg:      mainPkg,
			file:     mockFile,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linter.shouldSkipFile(tt.pkg, tt.file)
			assert.Equal(t, tt.expected, result, "неожиданный результат для случая %s", tt.name)
		})
	}
}

// TestIsMethodUsed_SkipFiles проверяет пропуск файлов в isMethodUsed
func TestIsMethodUsed_SkipFiles(t *testing.T) {
	// Создаем FileSet для позиций
	fset := token.NewFileSet()
	mainFileNode := fset.AddFile("main.go", -1, 1000)
	mainFileNode.SetLines([]int{0, 100})
	testFileNode := fset.AddFile("main_test.go", -1, 1000)
	testFileNode.SetLines([]int{0, 100})
	mockFileNode := fset.AddFile("mock/mock.go", -1, 1000)
	mockFileNode.SetLines([]int{0, 100})

	// Создаем тестовые файлы
	mainFile := &ast.File{
		Name:    &ast.Ident{Name: "main.go"},
		Package: mainFileNode.Pos(1),
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: &ast.Ident{Name: "TestMethod"},
				Type: &ast.FuncType{},
			},
		},
	}

	testFile := &ast.File{
		Name:    &ast.Ident{Name: "main_test.go"},
		Package: testFileNode.Pos(1),
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: &ast.Ident{Name: "TestMethod"},
				Type: &ast.FuncType{},
			},
		},
	}

	mockFile := &ast.File{
		Name:    &ast.Ident{Name: "mock.go"},
		Package: mockFileNode.Pos(1),
		Decls: []ast.Decl{
			&ast.FuncDecl{
				Name: &ast.Ident{Name: "TestMethod"},
				Type: &ast.FuncType{},
			},
		},
	}

	// Создаем тестовые пакеты
	mainPkg := &packages.Package{
		PkgPath: "example.com/myapp",
		Fset:    fset,
		Syntax:  []*ast.File{mainFile},
	}

	testPkg := &packages.Package{
		PkgPath: "example.com/myapp_test",
		Fset:    fset,
		Syntax:  []*ast.File{testFile},
	}

	mockPkg := &packages.Package{
		PkgPath: "example.com/myapp/mock",
		Fset:    fset,
		Syntax:  []*ast.File{mockFile},
	}

	// Создаем линтер
	linter := &UnusedMethodLinter{
		verbose: true,
		config:  config.DefaultConfig(),
		packages: []*packages.Package{
			mainPkg,
			testPkg,
			mockPkg,
		},
	}

	// Создаем тестовый интерфейс и метод
	iface := types.NewInterface(nil, nil)
	method := InterfaceMethod{
		InterfaceName: "TestInterface",
		MethodName:    "TestMethod",
		Interface:     iface,
	}

	// Проверяем использование метода
	result := linter.isMethodUsed(method)

	// Метод не должен быть найден, так как все файлы должны быть пропущены
	assert.False(t, result, "метод не должен быть найден в пропущенных файлах")
}

// mockConfig реализует интерфейс конфигурации для тестирования
type mockConfig struct {
	ignoredFiles map[string]bool
}

func (c *mockConfig) ShouldIgnore(filePath string) bool {
	return c.ignoredFiles[filePath]
}

func TestExtractInterfaceMethods_SkipTestPackagesAndIgnoredFiles(t *testing.T) {
	// Создаем мок конфигурации
	mockCfg := &mockConfig{
		ignoredFiles: map[string]bool{
			"ignored.go": true,
			"main.go":    false,
		},
	}

	// Создаем линтер с мок-конфигурацией
	linter := &UnusedMethodLinter{
		methods:         make([]InterfaceMethod, 0),
		genericWarnings: make([]GenericWarning, 0),
		verbose:         true,
		config:          mockCfg,
	}

	// Создаем FileSet и файлы
	fset := token.NewFileSet()
	mainFile := &ast.File{
		Name: &ast.Ident{Name: "main.go"},
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{Name: "MainInterface"},
						Type: &ast.InterfaceType{
							Methods: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{{Name: "MainMethod"}},
										Type:  &ast.FuncType{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ignoredFile := &ast.File{
		Name: &ast.Ident{Name: "ignored.go"},
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{Name: "IgnoredInterface"},
						Type: &ast.InterfaceType{
							Methods: &ast.FieldList{
								List: []*ast.Field{
									{
										Names: []*ast.Ident{{Name: "IgnoredMethod"}},
										Type:  &ast.FuncType{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Создаем пакеты и их типы
	mainPkgTypes := types.NewPackage("example.com/myapp", "myapp")
	testPkgTypes := types.NewPackage("example.com/myapp_test", "myapp_test")

	mainPkg := &packages.Package{
		PkgPath: "example.com/myapp",
		Types:   mainPkgTypes,
		Fset:    fset,
		TypesInfo: &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
		},
		Syntax: []*ast.File{mainFile},
	}

	testPkg := &packages.Package{
		PkgPath: "example.com/myapp_test",
		Types:   testPkgTypes,
		Fset:    fset,
		TypesInfo: &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
		},
		Syntax: []*ast.File{ignoredFile},
	}

	// Регистрируем типы интерфейсов
	mainInterfaceType := types.NewInterfaceType(
		[]*types.Func{
			types.NewFunc(token.NoPos, mainPkgTypes, "MainMethod",
				types.NewSignature(nil, nil, nil, false)),
		},
		nil,
	)
	mainInterfaceType.Complete()
	mainIdent := mainFile.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name
	mainPkg.TypesInfo.Defs[mainIdent] = types.NewTypeName(token.NoPos, mainPkgTypes, "MainInterface",
		types.NewNamed(types.NewTypeName(token.NoPos, mainPkgTypes, "MainInterface", nil), mainInterfaceType, nil))

	ignoredInterfaceType := types.NewInterfaceType(
		[]*types.Func{
			types.NewFunc(token.NoPos, testPkgTypes, "IgnoredMethod",
				types.NewSignature(nil, nil, nil, false)),
		},
		nil,
	)
	ignoredInterfaceType.Complete()
	ignoredIdent := ignoredFile.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name
	testPkg.TypesInfo.Defs[ignoredIdent] = types.NewTypeName(token.NoPos, testPkgTypes, "IgnoredInterface",
		types.NewNamed(types.NewTypeName(token.NoPos, testPkgTypes, "IgnoredInterface", nil), ignoredInterfaceType, nil))

	linter.packages = []*packages.Package{mainPkg, testPkg}

	// Запускаем извлечение методов
	linter.ExtractInterfaceMethods()

	// Проверяем результаты
	foundMainMethod := false
	foundIgnoredMethod := false

	for _, method := range linter.methods {
		switch {
		case method.InterfaceName == "MainInterface" && method.MethodName == "MainMethod":
			foundMainMethod = true
		case method.InterfaceName == "IgnoredInterface" && method.MethodName == "IgnoredMethod":
			foundIgnoredMethod = true
		}
	}

	assert.True(t, foundMainMethod, "метод из основного файла должен быть извлечен")
	assert.False(t, foundIgnoredMethod, "метод из игнорируемого файла не должен быть извлечен")
}

func TestLoadPackages_ExcludeIgnoredDirs(t *testing.T) {
	// Создаем временную директорию для тестов
	tmpDir := t.TempDir()

	// Создаем структуру директорий и файлов
	mainDir := filepath.Join(tmpDir, "main")
	mockDir := filepath.Join(tmpDir, "mock")
	err := os.MkdirAll(mainDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(mockDir, 0755)
	assert.NoError(t, err)

	// Создаем тестовые файлы с валидным Go-кодом
	mainFile := filepath.Join(mainDir, "main.go")
	mockFile := filepath.Join(mockDir, "mock.go")
	err = os.WriteFile(mainFile, []byte(`package main

type MainInterface interface {
	DoSomething()
}
`), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(mockFile, []byte(`package mock

type MockInterface interface {
	DoSomething()
}
`), 0644)
	assert.NoError(t, err)

	// Создаем go.mod файл для корректной работы go/packages
	goModContent := []byte("module test\n\ngo 1.21\n")
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), goModContent, 0644)
	assert.NoError(t, err)

	// Перехватываем вывод для проверки отладочных сообщений
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Создаем мок конфигурации
	mockCfg := &mockConfig{
		ignoredFiles: map[string]bool{
			mockDir: true,
		},
	}

	// Создаем линтер с мок-конфигурацией
	linter := &UnusedMethodLinter{
		methods:         make([]InterfaceMethod, 0),
		genericWarnings: make([]GenericWarning, 0),
		verbose:         true,
		config:          mockCfg,
	}

	// Загружаем пакеты
	err = linter.LoadPackages(tmpDir)
	assert.NoError(t, err)

	// Восстанавливаем stdout и читаем перехваченный вывод
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	// Проверяем, что пакет mock был исключен
	for _, pkg := range linter.packages {
		for _, file := range pkg.GoFiles {
			assert.NotContains(t, file, mockDir, "файл из игнорируемой директории не должен быть загружен")
		}
	}

	// Проверяем, что основной пакет был загружен
	foundMainPackage := false
	for _, pkg := range linter.packages {
		for _, file := range pkg.GoFiles {
			if strings.Contains(file, mainDir) {
				foundMainPackage = true
				break
			}
		}
	}
	assert.True(t, foundMainPackage, "основной пакет должен быть загружен")
}
