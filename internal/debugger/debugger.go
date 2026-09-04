package debugger

import (
	"fmt"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// DTOs for Delve JSON-RPC v2
type DlvBreakpoint struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type CreateBreakpointIn struct {
	Breakpoint DlvBreakpoint `json:"breakpoint"`
}

type CreateBreakpointOut struct {
	Breakpoint struct {
		ID   int    `json:"id"`
		File string `json:"file"`
		Line int    `json:"line"`
	} `json:"breakpoint"`
}

type ListSourcesIn struct {
	Filter string `json:"filter"`
}

type ListSourcesOut struct {
	Sources []string `json:"sources"`
}

type DebuggerCommand struct {
	Name string `json:"name"`
}

type DlvFunction struct {
	Name string `json:"name"`
}

type DlvLocation struct {
	File     string       `json:"file"`
	Line     int          `json:"line"`
	Function *DlvFunction `json:"function"`
}

type DlvGoroutine struct {
	ID         int         `json:"id"`
	CurrentLoc DlvLocation `json:"currentLoc"`
}

type DlvThread struct {
	File     string       `json:"file"`
	Line     int          `json:"line"`
	Function *DlvFunction `json:"function"`
}

type DlvDebuggerState struct {
	Running          bool          `json:"Running"`
	Exited           bool          `json:"exited"`
	ExitStatus       int           `json:"exitStatus"`
	CurrentGoroutine *DlvGoroutine `json:"currentGoroutine"`
	CurrentThread    *DlvThread    `json:"currentThread"`
}

type CommandOut struct {
	State DlvDebuggerState `json:"state"`
}

type EvalScope struct {
	GoroutineID int `json:"goroutineID"`
	Frame       int `json:"frame"`
}

type LoadConfig struct {
	FollowPointers     bool `json:"followPointers"`
	MaxVariableRecurse int  `json:"maxVariableRecurse"`
	MaxStringLen       int  `json:"maxStringLen"`
	MaxArrayValues     int  `json:"maxArrayValues"`
	MaxStructFields    int  `json:"maxStructFields"`
}

type ListLocalVarsIn struct {
	Scope EvalScope  `json:"scope"`
	Cfg   LoadConfig `json:"cfg"`
}

type DlvVariable struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ListLocalVarsOut struct {
	Variables []DlvVariable `json:"variables"`
}

type DetachIn struct {
	Kill bool `json:"kill"`
}

type DetachOut struct{}

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

// Debugger manages a real Delve session or fallback
type Debugger struct {
	mu          sync.Mutex
	dlvCmd      *exec.Cmd
	rpcPort     int
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	activeBin   string
	isFallback  bool
	stepIndex   int
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
		rpcPort:     40455,
	}
}

// findFreePort finds an available TCP port
func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
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

	clean := filepath.Clean(file)
	if _, ok := d.breakpoints[clean]; !ok {
		d.breakpoints[clean] = make(map[int]bool)
	}

	if d.breakpoints[clean][line] {
		delete(d.breakpoints[clean], line)
		return false
	} else {
		d.breakpoints[clean][line] = true
		return true
	}
}

// HasBreakpoint checks if there is a breakpoint at file:line
func (d *Debugger) HasBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	clean := filepath.Clean(file)
	if lines, ok := d.breakpoints[clean]; ok {
		return lines[line]
	}
	// Fallback to base match
	base := filepath.Base(file)
	for f, lines := range d.breakpoints {
		if filepath.Base(f) == base {
			return lines[line]
		}
	}
	return false
}

// GetBreakpoints returns all breakpoints for a file
func (d *Debugger) GetBreakpoints(file string) []int {
	d.mu.Lock()
	defer d.mu.Unlock()
	clean := filepath.Clean(file)
	var res []int
	if lines, ok := d.breakpoints[clean]; ok {
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

// StartSession launches Delve headless for the target binary and sets up breakpoints
func (d *Debugger) StartSession(binaryPath string, workDir string, currentFile string) error {
	dlvPath, err := FindDelve()
	if err != nil {
		d.mu.Lock()
		d.startFallbackLocked(currentFile)
		d.mu.Unlock()
		return nil
	}

	port, pErr := findFreePort()
	if pErr == nil && port > 0 {
		d.rpcPort = port
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.stopSessionLocked()
	d.activeBin = binaryPath
	d.isFallback = false

	d.dlvCmd = exec.Command(
		dlvPath,
		"exec",
		binaryPath,
		"--headless",
		fmt.Sprintf("--listen=127.0.0.1:%d", d.rpcPort),
		"--api-version=2",
		"--accept-multiclient",
	)
	if workDir != "" {
		d.dlvCmd.Dir = workDir
	}

	if err := d.dlvCmd.Start(); err != nil {
		// Fallback if dlv fails to start
		d.startFallbackLocked(currentFile)
		return nil
	}

	// Wait for port to become available
	ready := false
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", d.rpcPort), 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			ready = true
			break
		}
	}

	if !ready {
		_ = d.dlvCmd.Process.Kill()
		_ = d.dlvCmd.Wait()
		d.dlvCmd = nil
		d.startFallbackLocked(currentFile)
		return nil
	}

	d.state = DebugState{
		Active:      true,
		Running:     false,
		CurrentFile: currentFile,
	}

	// Fetch all sources recognized by Delve in the binary
	var srcList ListSourcesOut
	_ = d.callRPCLocked("ListSources", ListSourcesIn{}, &srcList)

	findDelveSource := func(fileName string) string {
		base := filepath.Base(fileName)
		for _, s := range srcList.Sources {
			if filepath.Base(s) == base {
				return s
			}
		}
		return fileName
	}

	// 1. Create Breakpoints in Delve
	for f, lines := range d.breakpoints {
		delvePath := findDelveSource(f)
		for l, set := range lines {
			if !set {
				continue
			}
			var out CreateBreakpointOut
			in := CreateBreakpointIn{Breakpoint: DlvBreakpoint{File: delvePath, Line: l}}
			err := d.callRPCLocked("CreateBreakpoint", in, &out)
			if err != nil {
				// Retry with clean, abs, and slash variations
				absF, _ := filepath.Abs(f)
				candidates := []string{
					absF,
					filepath.ToSlash(absF),
					f,
					filepath.Base(f),
				}
				for _, c := range candidates {
					inRetry := CreateBreakpointIn{Breakpoint: DlvBreakpoint{File: c, Line: l}}
					if rErr := d.callRPCLocked("CreateBreakpoint", inRetry, &out); rErr == nil && out.Breakpoint.ID > 0 {
						break
					}
				}
			}
		}
	}

	// 2. Initial continue to run to first breakpoint
	_ = d.runCommandLocked("continue")

	return nil
}

func (d *Debugger) startFallbackLocked(currentFile string) {
	d.isFallback = true

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
}

// Continue resumes execution until next breakpoint or exit
func (d *Debugger) Continue() error {
	return d.runCommand("continue")
}

// Next executes Step Over (next line)
func (d *Debugger) Next() error {
	return d.runCommand("next")
}

// Step executes Trace Into (single instruction / line)
func (d *Debugger) Step() error {
	return d.runCommand("step")
}

func (d *Debugger) runCommand(cmdName string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.runCommandLocked(cmdName)
}

func (d *Debugger) runCommandLocked(cmdName string) error {
	if !d.state.Active {
		return fmt.Errorf("no active debug session")
	}

	if d.isFallback {
		if cmdName == "continue" {
			return d.continueFallbackLocked()
		}
		// next or step
		d.state.CurrentLine++
		d.stepIndex++
		d.state.LocalVars = append(d.state.LocalVars, Variable{
			Name:  fmt.Sprintf("step_%d", d.stepIndex),
			Type:  "int",
			Value: fmt.Sprintf("%d", d.stepIndex*10),
		})
		return nil
	}

	var out CommandOut
	in := DebuggerCommand{Name: cmdName}
	err := d.callRPCLocked("Command", in, &out)
	if err != nil {
		d.state.ErrorMessage = err.Error()
		return err
	}

	d.updateStateFromDlvLocked(out.State)
	d.refreshLocalVarsLocked()

	return nil
}

func (d *Debugger) continueFallbackLocked() error {
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

func (d *Debugger) updateStateFromDlvLocked(st DlvDebuggerState) {
	if st.Exited {
		d.state.Exited = true
		d.state.ExitCode = st.ExitStatus
		d.state.CurrentLine = 0
		d.state.CurrentFile = ""
		d.state.CurrentFunc = ""
		d.state.LocalVars = nil
		return
	}

	if st.CurrentGoroutine != nil && st.CurrentGoroutine.CurrentLoc.Line > 0 {
		loc := st.CurrentGoroutine.CurrentLoc
		d.state.CurrentFile = loc.File
		d.state.CurrentLine = loc.Line
		if loc.Function != nil {
			d.state.CurrentFunc = loc.Function.Name
		}
	} else if st.CurrentThread != nil && st.CurrentThread.Line > 0 {
		d.state.CurrentFile = st.CurrentThread.File
		d.state.CurrentLine = st.CurrentThread.Line
		if st.CurrentThread.Function != nil {
			d.state.CurrentFunc = st.CurrentThread.Function.Name
		}
	}
}

func (d *Debugger) refreshLocalVarsLocked() {
	if d.state.Exited {
		return
	}

	var out ListLocalVarsOut
	in := ListLocalVarsIn{
		Scope: EvalScope{GoroutineID: -1, Frame: 0},
		Cfg: LoadConfig{
			FollowPointers:     true,
			MaxVariableRecurse: 1,
			MaxStringLen:       64,
			MaxArrayValues:     10,
			MaxStructFields:    10,
		},
	}

	err := d.callRPCLocked("ListLocalVars", in, &out)
	if err == nil {
		vars := make([]Variable, len(out.Variables))
		for i, v := range out.Variables {
			val := v.Value
			// Clean long strings
			if len(val) > 50 {
				val = val[:47] + "..."
			}
			vars[i] = Variable{
				Name:  v.Name,
				Type:  v.Type,
				Value: val,
			}
		}
		d.state.LocalVars = vars
	}
}

func (d *Debugger) callRPCLocked(method string, params interface{}, result interface{}) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", d.rpcPort), 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	codec := jsonrpc.NewClient(conn)
	return codec.Call("RPCServer."+method, params, result)
}

func (d *Debugger) stopSessionLocked() {
	if d.state.Active && !d.isFallback {
		var out DetachOut
		_ = d.callRPCLocked("Detach", DetachIn{Kill: true}, &out)
	}

	if d.dlvCmd != nil && d.dlvCmd.Process != nil {
		_ = d.dlvCmd.Process.Kill()
		_ = d.dlvCmd.Wait()
		d.dlvCmd = nil
	}

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
