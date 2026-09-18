package debugger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MaroShim/tg/internal/compiler"
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

func TestDebuggerMissingDelveError(t *testing.T) {
	t.Setenv("PATH", "")
	t.Setenv("GOPATH", t.TempDir())

	dbg := NewDebugger()
	err := dbg.StartSession("dummy_bin", t.TempDir(), "sample.go")
	if err == nil {
		t.Fatalf("expected error when delve is missing, got nil")
	}
	if !strings.Contains(err.Error(), "delve (dlv) not found") || !strings.Contains(err.Error(), "go install github.com/go-delve/delve/cmd/dlv@latest") {
		t.Errorf("expected actionable error message with installation instruction, got: %v", err)
	}
}

func TestDebuggerSession(t *testing.T) {
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
	// Breakpoint in Fibonacci: line 12 (a, b = b, a+b)
	dbg.SetBreakpoint(fibPath, 12)
	// Breakpoint in main: line 23 (fib = Fibonacci(i))
	dbg.SetBreakpoint(fibPath, 23)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(fibPath), fibPath)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	for iter := 0; iter < 15; iter++ {
		st := dbg.GetState()
		t.Logf("Hit %2d: Line=%d, Func=%s, Active=%v, Exited=%v",
			iter, st.CurrentLine, st.CurrentFunc, st.Active, st.Exited)
		for _, v := range st.LocalVars {
			t.Logf("    %s = %s", v.Name, v.Value)
		}
		if st.Exited {
			break
		}
		err = dbg.Continue()
		if err != nil {
			t.Fatalf("continue failed at iter %d: %v", iter, err)
		}
	}
}

func TestDynamicBreakpointDuringSession(t *testing.T) {
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
	// Start with ONLY breakpoint at line 23
	dbg.SetBreakpoint(fibPath, 23)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(fibPath), fibPath)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	st0 := dbg.GetState()
	if st0.CurrentLine != 23 {
		t.Fatalf("expected initial break at line 23, got %d", st0.CurrentLine)
	}

	// Now dynamically set breakpoint on line 12 while session is ALIVE!
	dbg.SetBreakpoint(fibPath, 12)

	hitLine12 := false
	for iter := 0; iter < 10; iter++ {
		err = dbg.Continue()
		if err != nil {
			t.Fatalf("continue failed: %v", err)
		}
		st := dbg.GetState()
		if st.CurrentLine == 12 && st.CurrentFunc == "main.Fibonacci" {
			hitLine12 = true
			break
		}
	}

	if !hitLine12 {
		t.Errorf("expected dynamically added breakpoint at line 12 to be hit")
	}
}

func TestF7ContinuousTracing(t *testing.T) {
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
	// Breakpoints on 12 and 23
	dbg.SetBreakpoint(fibPath, 12)
	dbg.SetBreakpoint(fibPath, 23)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(fibPath), fibPath)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	hitLine12 := false
	hitLine23Count := 0

	for step := 0; step < 40; step++ {
		st := dbg.GetState()
		if st.Exited {
			break
		}
		// Must always stay within user project file, never leak to stdlib or runtime
		if filepath.Base(st.CurrentFile) != "main.go" {
			t.Errorf("step %d leaked to non-user file: %s (func=%s)", step, st.CurrentFile, st.CurrentFunc)
		}
		if st.CurrentLine == 12 {
			hitLine12 = true
		}
		if st.CurrentLine == 23 {
			hitLine23Count++
		}
		err = dbg.Step()
		if err != nil {
			t.Fatalf("Step error at step %d: %v", step, err)
		}
	}

	if !hitLine12 {
		t.Errorf("expected line 12 in Fibonacci to be traced into by F7")
	}
	if hitLine23Count < 3 {
		t.Errorf("expected line 23 in main to be hit multiple times during loop tracing, got %d", hitLine23Count)
	}
}

func TestMultiFileBreakpointsDelve(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.go")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.go")
	calcStats := filepath.Join(filepath.Dir(calcMain), "stats.go")

	bRes := compiler.BuildDebug(calcMain)
	if !bRes.Success {
		t.Fatalf("build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	dbg := NewDebugger()
	// Breakpoint in math.go:5 (Factorial)
	dbg.SetBreakpoint(calcMath, 5)
	// Breakpoint in main.go:14
	dbg.SetBreakpoint(calcMain, 14)
	// Breakpoint in stats.go:7 (Average)
	dbg.SetBreakpoint(calcStats, 7)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(calcMain), calcMain)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	// 1. First stop should be at math.go:5 inside Factorial()
	st1 := dbg.GetState()
	if !st1.Active || st1.CurrentLine != 5 || filepath.Base(st1.CurrentFile) != "math.go" {
		t.Fatalf("expected first break at math.go:5, got %s:%d (func=%s)", st1.CurrentFile, st1.CurrentLine, st1.CurrentFunc)
	}

	// 2. Continue to second breakpoint: main.go:14
	if err := dbg.Continue(); err != nil {
		t.Fatalf("continue to main.go failed: %v", err)
	}
	st2 := dbg.GetState()
	if !st2.Active || st2.CurrentLine != 14 || filepath.Base(st2.CurrentFile) != "main.go" {
		t.Fatalf("expected second break at main.go:14, got %s:%d (func=%s)", st2.CurrentFile, st2.CurrentLine, st2.CurrentFunc)
	}

	// 3. Continue to third breakpoint: stats.go:7 inside Average()
	if err := dbg.Continue(); err != nil {
		t.Fatalf("continue to stats.go failed: %v", err)
	}
	st3 := dbg.GetState()
	if !st3.Active || st3.CurrentLine != 7 || filepath.Base(st3.CurrentFile) != "stats.go" {
		t.Fatalf("expected third break at stats.go:7, got %s:%d (func=%s)", st3.CurrentFile, st3.CurrentLine, st3.CurrentFunc)
	}
}

func TestStepIntoMathGo(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.go")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}

	bRes := compiler.BuildDebug(calcMain)
	if !bRes.Success {
		t.Fatalf("build failed: %s", bRes.RawOutput)
	}
	defer os.Remove(bRes.BinaryPath)

	dbg := NewDebugger()
	dbg.SetBreakpoint(calcMain, 13)

	err = dbg.StartSession(bRes.BinaryPath, filepath.Dir(calcMain), calcMain)
	if err != nil {
		t.Fatalf("failed to start debug: %v", err)
	}
	defer dbg.Stop()

	st1 := dbg.GetState()
	t.Logf("Break at: File=%q, Line=%d, Func=%q", st1.CurrentFile, st1.CurrentLine, st1.CurrentFunc)

	err = dbg.Step()
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}
	st2 := dbg.GetState()
	t.Logf("After Step: File=%q, Line=%d, Func=%q", st2.CurrentFile, st2.CurrentLine, st2.CurrentFunc)
}






