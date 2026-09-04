package debugger

import (
	"os"
	"path/filepath"
	"testing"

	"tg/internal/compiler"
)

func TestDebuggerBreakpoints(t *testing.T) {
	dbg := NewDebugger()
	file := "main.go"

	if dbg.HasBreakpoint(file, 10) {
		t.Errorf("breakpoint should not exist yet")
	}

	set := dbg.ToggleBreakpoint(file, 10)
	if !set || !dbg.HasBreakpoint(file, 10) {
		t.Errorf("expected breakpoint at line 10 to be set")
	}

	set = dbg.ToggleBreakpoint(file, 10)
	if set || dbg.HasBreakpoint(file, 10) {
		t.Errorf("expected breakpoint at line 10 to be removed")
	}
}

func TestDebuggerSessionAndFallback(t *testing.T) {
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

	bRes := compiler.BuildDebug(srcFile)
	if !bRes.Success {
		t.Fatalf("failed to build debug binary: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	dbg := NewDebugger()
	isSet := dbg.ToggleBreakpoint(filepath.Base(srcFile), 7)
	if !isSet {
		t.Fatalf("failed to set breakpoint")
	}

	err := dbg.StartSession(bRes.BinaryPath, tmpDir, srcFile)
	if err != nil {
		t.Fatalf("failed to start debug session: %v", err)
	}
	defer dbg.Stop()

	st := dbg.GetState()
	if !st.Active || st.CurrentLine != 7 {
		t.Errorf("expected break at line 7, got line %d (active=%v)", st.CurrentLine, st.Active)
	}

	// Step Over
	err = dbg.Next()
	if err != nil {
		t.Errorf("unexpected Next error: %v", err)
	}
	stAfter := dbg.GetState()
	if stAfter.CurrentLine != 8 {
		t.Errorf("expected line 8 after Next, got %d", stAfter.CurrentLine)
	}

	dbg.Stop()
	if dbg.IsActive() {
		t.Errorf("expected debugger to be inactive after stop")
	}
}
