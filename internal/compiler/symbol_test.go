package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDefinitionInProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tg_test_def_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "main.go")
	file2 := filepath.Join(tempDir, "calc.go")

	code1 := `package main

func main() {
	println(Add(1, 2))
}
`
	code2 := `package main

// Add two numbers
func Add(a int, b int) int {
	return a + b
}

type MyStruct struct {
	val int
}
`
	if err := os.WriteFile(file1, []byte(code1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(code2), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Find Add definition from main.go
	f, line, col, ok := FindDefinitionInProject(file1, "Add")
	if !ok {
		t.Fatalf("expected to find definition for Add")
	}
	if filepath.Base(f) != "calc.go" {
		t.Errorf("expected calc.go, got %s", f)
	}
	if line != 4 {
		t.Errorf("expected line 4, got %d", line)
	}
	if col < 1 {
		t.Errorf("invalid col %d", col)
	}

	// Test 2: Find MyStruct definition
	f, line, _, ok = FindDefinitionInProject(file1, "MyStruct")
	if !ok {
		t.Fatalf("expected to find definition for MyStruct")
	}
	if line != 8 {
		t.Errorf("expected line 8, got %d", line)
	}

	// Test 3: Search in project
	matches := SearchInProject(file1, "Add", false)
	if len(matches) < 2 {
		t.Errorf("expected at least 2 matches for Add, got %d", len(matches))
	}
}
