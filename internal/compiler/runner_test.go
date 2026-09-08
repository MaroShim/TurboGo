package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompilerBuildAndRun(t *testing.T) {
	// Create a temporary valid Go file
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "main.go")
	code := `package main

import "fmt"

func main() {
	fmt.Println("TURBO GO OK")
}
`
	if err := os.WriteFile(srcFile, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}

	// 1. Test line counter
	lines := CountLines(srcFile)
	if lines < 5 {
		t.Errorf("expected at least 5 lines, got %d", lines)
	}

	// 2. Test Build Success
	bRes := Build(srcFile)
	if !bRes.Success {
		t.Fatalf("build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	// 3. Test Run
	rRes := RunBinary(bRes.BinaryPath, tmpDir, nil)
	if !rRes.Completed || rRes.ExitCode != 0 {
		t.Errorf("run failed, exit code %d, output: %s", rRes.ExitCode, rRes.Output)
	}
	if rRes.Output != "TURBO GO OK\n" {
		t.Errorf("unexpected output: %q", rRes.Output)
	}

	// 4. Test Build Failure and Error Parsing
	badDir := filepath.Join(tmpDir, "bad")
	_ = os.MkdirAll(badDir, 0755)
	badCode := `package main
func main() {
	undefinedVariable = 123
}
`
	badFile := filepath.Join(badDir, "bad.go")
	_ = os.WriteFile(badFile, []byte(badCode), 0644)
	badRes := Build(badFile)
	if badRes.Success {
		t.Errorf("expected build to fail for bad code")
	}
	if len(badRes.Errors) == 0 {
		t.Errorf("expected parsed errors from compiler output, got 0: %s", badRes.RawOutput)
	} else {
		errItem := badRes.Errors[0]
		if errItem.Line != 3 {
			t.Errorf("expected error at line 3, got %d", errItem.Line)
		}
	}
}

func TestCompilerMultiFileBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create main.go which calls Helper() from helper.go
	mainFile := filepath.Join(tmpDir, "main.go")
	mainCode := `package main

import "fmt"

func main() {
	fmt.Println(Helper())
}
`
	if err := os.WriteFile(mainFile, []byte(mainCode), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// 2. Create helper.go in the same directory (no go.mod)
	helperFile := filepath.Join(tmpDir, "helper.go")
	helperCode := `package main

func Helper() string {
	return "MULTIFILE_OK"
}
`
	if err := os.WriteFile(helperFile, []byte(helperCode), 0644); err != nil {
		t.Fatalf("failed to write helper.go: %v", err)
	}

	// 3. Count lines should count both files
	totalLines := CountLines(mainFile)
	if totalLines < 10 {
		t.Errorf("expected at least 10 lines across both files, got %d", totalLines)
	}

	// 4. Build from main.go
	bRes := Build(mainFile)
	if !bRes.Success {
		t.Fatalf("multi-file build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	// 5. Run binary and verify output
	rRes := RunBinary(bRes.BinaryPath, tmpDir, nil)
	if !rRes.Completed || rRes.ExitCode != 0 {
		t.Errorf("run failed: %v", rRes)
	}
	if rRes.Output != "MULTIFILE_OK\n" {
		t.Errorf("expected 'MULTIFILE_OK', got %q", rRes.Output)
	}
}

func TestCompilerModuleProjectBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create go.mod
	goModFile := filepath.Join(tmpDir, "go.mod")
	goModContent := "module testmod\n\ngo 1.20\n"
	if err := os.WriteFile(goModFile, []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// 2. Create main.go in sub-package cmd/app
	cmdDir := filepath.Join(tmpDir, "cmd", "app")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to mkdir cmd/app: %v", err)
	}

	mainFile := filepath.Join(cmdDir, "main.go")
	mainCode := `package main

import "fmt"

func main() {
	fmt.Println(GetMessage())
}
`
	_ = os.WriteFile(mainFile, []byte(mainCode), 0644)

	extraFile := filepath.Join(cmdDir, "msg.go")
	extraCode := `package main

func GetMessage() string {
	return "MODULE_OK"
}
`
	_ = os.WriteFile(extraFile, []byte(extraCode), 0644)

	// 3. Build from cmd/app/main.go
	bRes := Build(mainFile)
	if !bRes.Success {
		t.Fatalf("module build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	// 4. Run binary
	rRes := RunBinary(bRes.BinaryPath, cmdDir, nil)
	if !rRes.Completed || rRes.ExitCode != 0 {
		t.Errorf("run failed: %v", rRes)
	}
	if rRes.Output != "MODULE_OK\n" {
		t.Errorf("expected 'MODULE_OK', got %q", rRes.Output)
	}
}
