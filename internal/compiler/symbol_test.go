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

func TestFindDefinitionEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tg_test_edge_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file := filepath.Join(tempDir, "service.go")
	code := `package main

// func FakeFunc() should not be matched as real definition
var comment = "// func FakeFunc() string"

type Worker interface {
	DoWork() error
}

func (s *MyStruct) Compute(val int) int {
	return val * 2
}

const MaxLimit = 100
var GlobalCounter = 0
`
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Method receiver definition
	f, line, _, ok := FindDefinitionInProject(file, "Compute")
	if !ok || line != 10 {
		t.Errorf("expected Compute method at line 10, got %s:%d (ok=%v)", f, line, ok)
	}

	// 2. Interface definition
	_, line, _, ok = FindDefinitionInProject(file, "Worker")
	if !ok || line != 6 {
		t.Errorf("expected Worker interface at line 6, got line %d", line)
	}

	// 3. Const & Var definition
	_, line, _, ok = FindDefinitionInProject(file, "MaxLimit")
	if !ok || line != 14 {
		t.Errorf("expected MaxLimit at line 14, got line %d", line)
	}
	_, line, _, ok = FindDefinitionInProject(file, "GlobalCounter")
	if !ok || line != 15 {
		t.Errorf("expected GlobalCounter at line 15, got line %d", line)
	}

	// 4. Commented function should not be matched
	_, _, _, ok = FindDefinitionInProject(file, "FakeFunc")
	if ok {
		t.Errorf("FakeFunc inside comment should not be found as a definition")
	}

	// 5. Empty/blank search
	_, _, _, ok = FindDefinitionInProject(file, "   ")
	if ok {
		t.Errorf("blank symbol should return false")
	}
	emptyMatches := SearchInProject(file, "", false)
	if len(emptyMatches) != 0 {
		t.Errorf("empty query search should return empty matches")
	}
}
