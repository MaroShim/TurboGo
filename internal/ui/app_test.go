package ui

import (
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestAppMultiFileDebugging(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.go")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.go")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// 1. Set a breakpoint in math.go:5
	if err := app.editor.LoadFile(calcMath); err != nil {
		t.Fatalf("failed to load math.go: %v", err)
	}
	app.ToggleBreakpoint(5)

	// 2. Switch back to main.go and set a breakpoint at line 14
	if err := app.editor.LoadFile(calcMain); err != nil {
		t.Fatalf("failed to load main.go: %v", err)
	}
	app.ToggleBreakpoint(14)

	// Verify FileBreakpoints recorded both files
	if len(app.editor.FileBreakpoints[filepath.Clean(calcMath)]) == 0 {
		t.Errorf("expected math.go breakpoints to be stored in FileBreakpoints")
	}

	// 3. Start debugging from main.go
	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	// 4. Since Factorial() is called first in main(), Delve stops at math.go:5
	// Verify editor switched to math.go automatically
	cleanMath := filepath.Clean(calcMath)
	if filepath.Clean(app.editor.FilePath) != cleanMath {
		t.Fatalf("expected editor to switch to math.go, but is at %s", app.editor.FilePath)
	}
	if app.editor.CurrentIP != 5 {
		t.Fatalf("expected CurrentIP to be 5 in math.go, got %d", app.editor.CurrentIP)
	}

	// 5. Continue execution: should hit main.go:14
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue failed: %v", err)
	}

	cleanMain := filepath.Clean(calcMain)
	if filepath.Clean(app.editor.FilePath) != cleanMain {
		t.Fatalf("expected editor to switch back to main.go, but is at %s", app.editor.FilePath)
	}
	if app.editor.CurrentIP != 14 {
		t.Fatalf("expected CurrentIP to be 14 in main.go, got %d", app.editor.CurrentIP)
	}
}

func TestAppF7StepIntoMathGo(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.go")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.go")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// Set a single breakpoint at line 13 of main.go (fact := Factorial(n))
	app.ToggleBreakpoint(13)

	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	t.Logf("Stopped initially: File=%s, Line=%d, CurrentIP=%d", app.editor.FilePath, app.debugger.GetState().CurrentLine, app.editor.CurrentIP)

	if app.editor.CurrentIP != 13 {
		t.Fatalf("expected CurrentIP 13, got %d", app.editor.CurrentIP)
	}

	// Press F7 (DebugStepInto) on line 13
	err = app.DebugStepInto()
	if err != nil {
		t.Fatalf("DebugStepInto failed: %v", err)
	}

	st := app.debugger.GetState()
	t.Logf("After F7: Debugger State File=%s, Line=%d, Func=%s", st.CurrentFile, st.CurrentLine, st.CurrentFunc)
	t.Logf("After F7: Editor FilePath=%s, CurrentIP=%d", app.editor.FilePath, app.editor.CurrentIP)

	cleanMath := filepath.Clean(calcMath)
	if filepath.Clean(app.editor.FilePath) != cleanMath {
		t.Fatalf("expected editor to switch to math.go, but is at %s", app.editor.FilePath)
	}
	if app.editor.CurrentIP != 4 {
		t.Fatalf("expected CurrentIP 4 in math.go, got %d", app.editor.CurrentIP)
	}
}

