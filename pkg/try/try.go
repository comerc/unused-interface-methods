package try

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CheckCode(goFiles []string) ([]string, error) {
	tmpDir, err := os.MkdirTemp("", "interface-check")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Создаем временные копии файлов с сохранением структуры директорий
	var tmpFiles []string
	for _, file := range goFiles {
		// Создаем все необходимые поддиректории
		dstFile := filepath.Join(tmpDir, filepath.Base(file))
		err := copyFile(file, dstFile)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file: %v", err)
		}
		tmpFiles = append(tmpFiles, filepath.Base(file))
	}

	var unusedMethods []string

	for _, file := range tmpFiles {
		content, err := readFileToString(filepath.Join(tmpDir, file))
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file, err)
		}

		interfaces := findInterfaces(content)
		fmt.Printf("Found interfaces in %s: %v\n", file, interfaces)

		for _, iface := range interfaces {
			methods := findMethods(content, iface)
			fmt.Printf("Found methods in %s: %v\n", iface, methods)

			for _, method := range methods {
				newContent := removeMethod(content, iface, method)
				err = writeStringToFile(filepath.Join(tmpDir, file), newContent)
				if err != nil {
					return nil, fmt.Errorf("failed to write modified file: %v", err)
				}

				fmt.Printf("Checking %s.%s:\n", iface, method)
				output, err := runStaticcheck(tmpDir)
				fmt.Println(output)

				if err == nil {
					unusedMethods = append(unusedMethods, fmt.Sprintf("%s.%s", iface, method))
				}

				// TODO: можно не восстанавливать начальное состоятие,
				// пока удаление метода не ломает staticcheck
				err = writeStringToFile(filepath.Join(tmpDir, file), content)
				if err != nil {
					return nil, fmt.Errorf("failed to restore file: %v", err)
				}
			}
		}
	}

	return unusedMethods, nil
}

func copyFile(src, dst string) error {
	content, err := readFileToString(src)
	if err != nil {
		return err
	}
	return writeStringToFile(dst, content)
}

func readFileToString(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	reader := bufio.NewReader(file)
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func writeStringToFile(path string, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return err
	}

	return writer.Flush()
}

func findInterfaces(content string) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %v\n", err)
		return nil
	}

	var interfaces []string
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				interfaces = append(interfaces, typeSpec.Name.Name)
			}
		}
		return true
	})

	return interfaces
}

func findMethods(content, interfaceName string) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %v\n", err)
		return nil
	}

	var methods []string
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == interfaceName {
				if ifaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					for _, method := range ifaceType.Methods.List {
						// Если это обычный метод (не встроенный интерфейс)
						if len(method.Names) > 0 {
							methods = append(methods, method.Names[0].Name)
						}
					}
				}
			}
		}
		return true
	})

	return methods
}

func removeMethod(content, interfaceName, methodName string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inInterface := false
	skipNext := false

	for _, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}

		if strings.Contains(line, "type "+interfaceName+" interface") {
			inInterface = true
		}

		if inInterface && strings.Contains(line, methodName+"(") {
			skipNext = strings.Count(line, "(") > strings.Count(line, ")")
			continue
		}

		result = append(result, line)

		if inInterface && strings.Contains(line, "}") {
			inInterface = false
		}
	}

	return strings.Join(result, "\n")
}

func runStaticcheck(dir string) (string, error) {
	cmd := exec.Command("staticcheck", ".")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
