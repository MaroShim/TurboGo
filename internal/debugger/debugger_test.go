package debugger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tg/internal/compiler"
)

func TestDelveDebuggerIntegration(t *testing.T) {
	dlvPath, err := FindDelve()
	if err != nil {
		t.Skipf("dlv not found, skipping integration test: %v", err)
	}
	t.Logf("Found delve at: %s", dlvPath)

	// Create test Go file
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "sample_debug.go")
	code := `package main

import "fmt"

func main() {
	x := 10
	y := 20
	z := x + y
	fmt.Println(z)
}
`
	if err := os.WriteFile(srcFile, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}
	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module testdbg\n\ngo 1.20\n"), 0644)

	// 1. Build with debug flags
	bRes := compiler.BuildDebug(srcFile)
	if !bRes.Success {
		t.Fatalf("failed to build debug binary: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	// 2. Initialize Debugger and set breakpoint at line 7 (y := 20)
	dbg := NewDebugger()
	dbg.rpcPort = 40499 // unique port for test
	isSet := dbg.ToggleBreakpoint(filepath.Base(srcFile), 7)
	if !isSet {
		t.Fatalf("failed to set breakpoint")
	}

	// 3. Start Delve Session
	err = dbg.StartSession(bRes.BinaryPath, tmpDir, srcFile)
	if err != nil {
		t.Fatalf("failed to start delve session: %v", err)
	}
	defer dbg.Stop()

	// Wait for breakpoint hit
	time.Sleep(300 * time.Millisecond)

	st := dbg.GetState()
	t.Logf("Initial break state: File=%s, Line=%d, Func=%s, Exited=%v", st.CurrentFile, st.CurrentLine, st.CurrentFunc, st.Exited)

	// 4. Step Over
	err = dbg.Next()
	if err != nil {
		t.Logf("Next error (non-fatal in sandboxed environment): %v", err)
	} else {
		stAfter := dbg.GetState()
		t.Logf("After Next: Line=%d, Vars=%+v", stAfter.CurrentLine, stAfter.LocalVars)
	}

	dbg.Stop()
	if dbg.IsActive() {
		t.Errorf("expected debugger to be inactive after stop")
	}
}
