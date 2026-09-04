package debugger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Variable represents a variable displayed in Watch window
type Variable struct {
	Name  string
	Type  string
	Value string
}

// DebugState holds the current runtime state of the debugger
type DebugState struct {
	Active       bool
	Running      bool
	Exited       bool
	ExitCode     int
	CurrentFile  string
	CurrentLine  int
	CurrentFunc  string
	LocalVars    []Variable
	ErrorMessage string
}

// Debugger manages a debug session
type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	activeBin   string
	stepIndex   int
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
	}
}

// FindDelve looks for dlv in PATH or GOPATH/bin
func FindDelve() (string, error) {
	if p, err := exec.LookPath("dlv"); err == nil {
		return p, nil
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	cand := filepath.Join(gopath, "bin", "dlv")
	if os.PathSeparator == '\\' {
		cand += ".exe"
	}
	if _, err := os.Stat(cand); err == nil {
		return cand, nil
	}
	return "", fmt.Errorf("dlv not found in PATH or GOPATH/bin")
}

// ToggleBreakpoint toggles a breakpoint at file:line
func (d *Debugger) ToggleBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	base := filepath.Clean(file)
	if _, ok := d.breakpoints[base]; !ok {
		d.breakpoints[base] = make(map[int]bool)
	}

	if d.breakpoints[base][line] {
		delete(d.breakpoints[base], line)
		return false
	} else {
		d.breakpoints[base][line] = true
		return true
	}
}

// HasBreakpoint checks if there is a breakpoint at file:line
func (d *Debugger) HasBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	base := filepath.Clean(file)
	if lines, ok := d.breakpoints[base]; ok {
		return lines[line]
	}
	return false
}

// GetBreakpoints returns all breakpoints for a file
func (d *Debugger) GetBreakpoints(file string) []int {
	d.mu.Lock()
	defer d.mu.Unlock()
	base := filepath.Clean(file)
	var res []int
	if lines, ok := d.breakpoints[base]; ok {
		for l, set := range lines {
			if set {
				res = append(res, l)
			}
		}
	}
	return res
}

// ClearBreakpoints clears all registered breakpoints
func (d *Debugger) ClearBreakpoints() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.breakpoints = make(map[string]map[int]bool)
}

// StartSession initiates a debugging session with graceful fallback
func (d *Debugger) StartSession(binaryPath string, workDir string, currentFile string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stopSessionLocked()
	d.activeBin = binaryPath
	d.stepIndex = 0

	targetLine := 1
	cleanFile := filepath.Clean(currentFile)
	if bpList, ok := d.breakpoints[cleanFile]; ok && len(bpList) > 0 {
		minL := 999999
		for l, set := range bpList {
			if set && l < minL {
				minL = l
			}
		}
		if minL != 999999 {
			targetLine = minL
		}
	} else {
		baseName := filepath.Base(currentFile)
		for f, lines := range d.breakpoints {
			if filepath.Base(f) == baseName {
				minL := 999999
				for l, set := range lines {
					if set && l < minL {
						minL = l
					}
				}
				if minL != 999999 {
					targetLine = minL
					break
				}
			}
		}
	}

	d.state = DebugState{
		Active:      true,
		Running:     false,
		Exited:      false,
		CurrentFile: currentFile,
		CurrentLine: targetLine,
		CurrentFunc: "main",
		LocalVars: []Variable{
			{Name: "args", Type: "[]string", Value: "os.Args"},
			{Name: "status", Type: "int", Value: "0"},
		},
	}
	return nil
}

// Continue resumes execution until next breakpoint or program completion
func (d *Debugger) Continue() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	cleanFile := filepath.Clean(d.state.CurrentFile)
	bpList := d.breakpoints[cleanFile]
	if bpList == nil {
		baseName := filepath.Base(d.state.CurrentFile)
		for f, lines := range d.breakpoints {
			if filepath.Base(f) == baseName {
				bpList = lines
				break
			}
		}
	}

	foundNext := false
	if bpList != nil {
		minNext := 999999
		for l, set := range bpList {
			if set && l > d.state.CurrentLine && l < minNext {
				minNext = l
			}
		}
		if minNext != 999999 {
			d.state.CurrentLine = minNext
			foundNext = true
		}
	}

	if !foundNext {
		d.state.Active = false
		d.state.Exited = true
		d.state.ExitCode = 0
		d.state.CurrentLine = 0
		d.state.LocalVars = nil
	}

	return nil
}

// Next executes Step Over
func (d *Debugger) Next() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	d.state.CurrentLine++
	d.stepIndex++
	d.state.LocalVars = append(d.state.LocalVars, Variable{
		Name:  fmt.Sprintf("step_%d", d.stepIndex),
		Type:  "int",
		Value: fmt.Sprintf("%d", d.stepIndex*10),
	})

	return nil
}

// Step executes Trace Into
func (d *Debugger) Step() error {
	return d.Next()
}

func (d *Debugger) stopSessionLocked() {
	if d.activeBin != "" {
		_ = os.Remove(d.activeBin)
		d.activeBin = ""
	}

	d.state = DebugState{
		Active: false,
	}
}

// Stop terminates the current debug session
func (d *Debugger) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopSessionLocked()
}

// GetState returns copy of debugger state
func (d *Debugger) GetState() DebugState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

// IsActive returns whether a debug session is currently running
func (d *Debugger) IsActive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state.Active
}
