package ui

import (
	"os"
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

func TestMultiFileFullLifecycleRegression(t *testing.T) {
	calcMain, err := filepath.Abs("../../examples/calc/main.go")
	if err != nil {
		t.Fatalf("failed to resolve calc main path: %v", err)
	}
	calcMath := filepath.Join(filepath.Dir(calcMain), "math.go")
	calcStats := filepath.Join(filepath.Dir(calcMain), "stats.go")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, calcMain)
	defer app.StopDebugging()

	// 1. Initial breakpoint at main.go:13 (fact := Factorial(n))
	app.ToggleBreakpoint(13)

	bRes, err := app.StartDebugging()
	if err != nil {
		t.Fatalf("StartDebugging failed: %v (build: %s)", err, bRes.RawOutput)
	}

	// 2. Initial stop at main.go:13
	if app.editor.FileName != "main.go" || app.editor.CurrentIP != 13 {
		t.Fatalf("expected initial stop at main.go:13, got %s:%d", app.editor.FileName, app.editor.CurrentIP)
	}

	// 3. F7 Step Into Factorial() -> math.go:4
	if err := app.DebugStepInto(); err != nil {
		t.Fatalf("DebugStepInto failed: %v", err)
	}
	if app.editor.FileName != "math.go" || app.editor.CurrentIP != 4 || filepath.Clean(app.editor.FilePath) != filepath.Clean(calcMath) {
		t.Fatalf("expected F7 to step into math.go:4 (%s), got %s:%d", calcMath, app.editor.FilePath, app.editor.CurrentIP)
	}

	// Verify local variable 'n' is present in watch state
	st := app.debugger.GetState()
	foundN := false
	for _, v := range st.LocalVars {
		if v.Name == "n" && v.Value == "7" {
			foundN = true
		}
	}
	if !foundN {
		t.Errorf("expected argument n=7 in watch locals, got %v", st.LocalVars)
	}

	// 4. F8 Step Over inside math.go -> line 5
	if err := app.DebugStepOver(); err != nil {
		t.Fatalf("DebugStepOver failed: %v", err)
	}
	if app.editor.FileName != "math.go" || app.editor.CurrentIP != 5 {
		t.Fatalf("expected F8 to step to math.go:5, got %s:%d", app.editor.FileName, app.editor.CurrentIP)
	}

	// 5. Dynamically set breakpoint on main.go:14 and continue to return to main.go
	app.debugger.SetBreakpoint(calcMain, 14)
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue to main.go:14 failed: %v", err)
	}
	if app.editor.FileName != "main.go" || app.editor.CurrentIP != 14 {
		t.Fatalf("expected return stop at main.go:14, got %s:%d", app.editor.FileName, app.editor.CurrentIP)
	}

	// 6. Dynamically set breakpoint in stats.go:7 and continue
	app.debugger.SetBreakpoint(calcStats, 7)
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue to stats.go:7 failed: %v", err)
	}
	if app.editor.FileName != "stats.go" || app.editor.CurrentIP != 7 {
		t.Fatalf("expected jump to stats.go:7, got %s:%d", app.editor.FileName, app.editor.CurrentIP)
	}

	// Remove breakpoint so it doesn't hit again during Variance/StdDev calculations
	app.debugger.RemoveBreakpoint(calcStats, 7)

	// 7. Continue to normal process exit
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue to exit failed: %v", err)
	}
	if app.debugger.IsActive() {
		t.Errorf("expected debug session to be inactive after exit")
	}
	if app.editor.CurrentIP != 0 {
		t.Errorf("expected CurrentIP=0 after exit, got %d", app.editor.CurrentIP)
	}
}

func TestUntitledScratchBuffer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tg_untitled_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()

	app := NewAppWithScreen(s, "")
	app.workDir = tempDir

	scratchFile := app.EnsureScratchBuffer()
	app.editor.FilePath = scratchFile
	app.editor.FileName = "NONAME00.GO"
	app.editor.IsUntitled = true

	if !app.editor.IsUntitled {
		t.Errorf("expected IsUntitled to be true")
	}
	if app.editor.FileName != "NONAME00.GO" {
		t.Errorf("expected FileName to be NONAME00.GO, got %s", app.editor.FileName)
	}
	if _, err := os.Stat(scratchFile); err != nil {
		t.Errorf("expected scratchFile %s to exist on disk: %v", scratchFile, err)
	}

	// Verify SaveFile preserves NONAME00.GO
	if err := app.editor.SaveFile(); err != nil {
		t.Errorf("SaveFile failed: %v", err)
	}
	if app.editor.FileName != "NONAME00.GO" {
		t.Errorf("expected FileName to remain NONAME00.GO after SaveFile, got %s", app.editor.FileName)
	}

	// Verify SaveAs sets real file
	savedPath := filepath.Join(tempDir, "real_main.go")
	if err := app.editor.SaveAs(savedPath); err != nil {
		t.Errorf("SaveAs failed: %v", err)
	}
	if app.editor.IsUntitled {
		t.Errorf("expected IsUntitled to be false after SaveAs")
	}
	if app.editor.FileName != "real_main.go" {
		t.Errorf("expected FileName to be real_main.go, got %s", app.editor.FileName)
	}

	// Verify Stop cleans up scratchDir
	app.Stop()
	if _, err := os.Stat(app.scratchDir); !os.IsNotExist(err) {
		t.Errorf("expected scratchDir %s to be deleted after Stop, err=%v", app.scratchDir, err)
	}
}

type mockModalDialog struct {
	visible bool
}

func (m *mockModalDialog) Draw(s tcell.Screen, w, h int) {}
func (m *mockModalDialog) IsVisible() bool { return m.visible }

func TestConfirmSaveDialogModalVisibility(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()

	app := NewAppWithScreen(s, "")
	mockDlg := &mockModalDialog{visible: false}
	app.SetConfirmSaveDialog(mockDlg)

	if app.HasModalVisible() {
		t.Errorf("expected HasModalVisible to be false")
	}

	mockDlg.visible = true
	if !app.HasModalVisible() {
		t.Errorf("expected HasModalVisible to be true when confirmSaveDialog is visible")
	}
}




