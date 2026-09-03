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
	badCode := `package main
func main() {
	undefinedVariable = 123
}
`
	badFile := filepath.Join(tmpDir, "bad.go")
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
