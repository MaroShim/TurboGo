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
