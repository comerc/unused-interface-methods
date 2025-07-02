package try

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckCode(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	goFiles := []string{
		filepath.Join(wd, "..", "..", "test", "data", "interfaces.go"),
		filepath.Join(wd, "..", "..", "test", "data", "generics.go"),
	}

	// Проверяем код через staticcheck
	unusedMethods, err := CheckCode(goFiles)
	if err != nil {
		t.Fatalf("Failed to check code: %v", err)
	}

	// Проверяем, что метод Process интерфейса StringHandler помечен как неиспользуемый
	found := false
	for _, method := range unusedMethods {
		if method == "StringHandler.Process" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected StringHandler.Process to be marked as unused")
	}

	t.Logf("Found unused methods: %v", unusedMethods)
}
