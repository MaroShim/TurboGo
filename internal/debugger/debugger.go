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

// Debugger manages a Delve session
type Debugger struct {
	mu          sync.Mutex
	dlvCmd      *exec.Cmd
	rpcPort     int
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	activeBin   string
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
		rpcPort:     40455,
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
	if _, err := os.Stat(cand); err == nil {
		return cand, nil
	}
	return "", fmt.Errorf("dlv not found in PATH or GOPATH/bin")
}

// ToggleBreakpoint toggles a breakpoint at file:line
func (d *Debugger) ToggleBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.breakpoints[file]; !ok {
		d.breakpoints[file] = make(map[int]bool)
	}

	if d.breakpoints[file][line] {
		delete(d.breakpoints[file], line)
		return false
	} else {
		d.breakpoints[file][line] = true
		return true
	}
}

// HasBreakpoint checks if there is a breakpoint at file:line
func (d *Debugger) HasBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if lines, ok := d.breakpoints[file]; ok {
		return lines[line]
	}
	return false
}

// GetBreakpoints returns all breakpoints for a file
func (d *Debugger) GetBreakpoints(file string) []int {
	d.mu.Lock()
	defer d.mu.Unlock()
	var res []int
	if lines, ok := d.breakpoints[file]; ok {
		for l := range lines {
			res = append(res, l)
		}
	}
	return res
}

// StartSession launches dlv headless for the target binary and sets up breakpoints
func (d *Debugger) StartSession(binaryPath string, workDir string, currentFile string) error {
	dlvPath, err := FindDelve()
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.stopSessionLocked()

	d.activeBin = binaryPath
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
		return fmt.Errorf("failed to start dlv: %w", err)
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
		return fmt.Errorf("timeout connecting to Delve RPC server")
	}

	d.state = DebugState{
		Active:  true,
		Running: false,
	}

	// 1. Create Breakpoints in Delve
	for f, lines := range d.breakpoints {
		for l := range lines {
			var out CreateBreakpointOut
			in := CreateBreakpointIn{Breakpoint: DlvBreakpoint{File: f, Line: l}}
			_ = d.callRPCLocked("CreateBreakpoint", in, &out)
		}
	}

	// 2. Initial continue to run to first breakpoint or main
	_ = d.runCommandLocked("continue")

	return nil
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
			vars[i] = Variable{
				Name:  v.Name,
				Type:  v.Type,
				Value: v.Value,
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
	if d.state.Active {
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
