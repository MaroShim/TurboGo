package debugger

import (
	"os"
	"path/filepath"
	"strings"
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

func TestFibonacciDebug(t *testing.T) {
	fibPath, err := filepath.Abs("../../examples/fibonacci/main.go")
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}

	bRes := compiler.BuildDebug(fibPath)
	if !bRes.Success {
		t.Fatalf("build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	dbg := NewDebugger()
	// Break at line 6: func Fibonacci(n int) uint64 {
	dbg.ToggleBreakpoint(fibPath, 6)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(fibPath), fibPath)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	st := dbg.GetState()
	if !st.Active || st.CurrentLine != 6 || st.CurrentFunc != "main.Fibonacci" {
		t.Fatalf("expected break at line 6 in main.Fibonacci, got line %d in %s (active=%v)", st.CurrentLine, st.CurrentFunc, st.Active)
	}

	foundN := false
	for _, v := range st.LocalVars {
		if v.Name == "n" {
			foundN = true
		}
		if strings.HasPrefix(v.Name, "~") {
			t.Errorf("compiler internal variable %q should be filtered out from watch list", v.Name)
		}
	}
	if !foundN {
		t.Errorf("expected argument 'n' to be present in LocalVars at line 6")
	}

	// Step to line 7: if n <= 1 {
	err = dbg.Next()
	if err != nil {
		t.Fatalf("next failed: %v", err)
	}
	stLine7 := dbg.GetState()
	if stLine7.CurrentLine != 7 {
		t.Errorf("expected line 7 after Next, got %d", stLine7.CurrentLine)
	}
	foundN7 := false
	for _, v := range stLine7.LocalVars {
		if v.Name == "n" {
			foundN7 = true
		}
	}
	if !foundN7 {
		t.Errorf("expected argument 'n' to be present in LocalVars at line 7")
	}
}

func TestMainLoopStepping(t *testing.T) {
	fibPath, err := filepath.Abs("../../examples/fibonacci/main.go")
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}

	bRes := compiler.BuildDebug(fibPath)
	if !bRes.Success {
		t.Fatalf("build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	dbg := NewDebugger()
	// Breakpoint in Fibonacci: line 6
	dbg.SetBreakpoint(fibPath, 6)
	// Breakpoint in main: line 23 (inside the loop body)
	dbg.SetBreakpoint(fibPath, 23)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(fibPath), fibPath)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	for iter := 0; iter < 10; iter++ {
		st := dbg.GetState()
		if !st.Active || st.Exited {
			t.Fatalf("expected active session at iter %d, got active=%v exited=%v", iter, st.Active, st.Exited)
		}
		if iter%2 == 0 {
			if st.CurrentLine != 23 || st.CurrentFunc != "main.main" {
				t.Errorf("iter %d: expected break at line 23 in main.main, got line %d in %s", iter, st.CurrentLine, st.CurrentFunc)
			}
		} else {
			if st.CurrentLine != 6 || st.CurrentFunc != "main.Fibonacci" {
				t.Errorf("iter %d: expected break at line 6 in main.Fibonacci, got line %d in %s", iter, st.CurrentLine, st.CurrentFunc)
			}
		}
		err = dbg.Continue()
		if err != nil {
			t.Fatalf("continue failed at iter %d: %v", iter, err)
		}
	}
}


