package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CompileError holds parsed compiler diagnostic information
type CompileError struct {
	File    string
	Line    int
	Column  int
	Level   string // "error" or "warning"
	Message string
}

// BuildResult holds outcome of go build
type BuildResult struct {
	Success       bool
	LinesCompiled int
	Duration      time.Duration
	Errors        []CompileError
	ErrorCount    int
	WarningCount  int
	BinaryPath    string
	RawOutput     string
}

// RunResult holds outcome of program execution
type RunResult struct {
	Output    string
	ExitCode  int
	Duration  time.Duration
	Completed bool
}

// Error regex pattern: file.go:line:col: message
var errRegex = regexp.MustCompile(`(?m)^([^:\n\r]+):(\d+):(\d+):\s*(.+)$`)

// CountLines counts total lines of go code in the specified target
func CountLines(targetPath string) int {
	total := 0
	info, err := os.Stat(targetPath)
	if err != nil {
		return 0
	}

	if !info.IsDir() {
		content, err := os.ReadFile(targetPath)
		if err == nil {
			total += bytes.Count(content, []byte("\n")) + 1
		}
		return total
	}

	_ = filepath.Walk(targetPath, func(path string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			if err == nil {
				total += bytes.Count(content, []byte("\n")) + 1
			}
		}
		return nil
	})
	return total
}

// Build compiles the target Go file or package
func Build(targetPath string) *BuildResult {
	start := time.Now()
	res := &BuildResult{}

	lines := CountLines(targetPath)
	res.LinesCompiled = lines

	dir := filepath.Dir(targetPath)
	if dir == "" {
		dir = "."
	}

	ext := ""
	if os.PathSeparator == '\\' {
		ext = ".exe"
	}

	// Create temp binary output
	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("turbogo_bin_%d%s", time.Now().UnixNano(), ext))
	res.BinaryPath = tmpBin

	cmd := exec.Command("go", "build", "-o", tmpBin, filepath.Base(targetPath))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.RawOutput = outBuf.String()

	if err == nil {
		res.Success = true
		res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput)
		return res
	}

	res.Success = false
	res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput)
	if res.ErrorCount == 0 && len(res.Errors) > 0 {
		res.ErrorCount = len(res.Errors)
	} else if res.ErrorCount == 0 {
		res.ErrorCount = 1
	}

	return res
}

// BuildDebug compiles the target Go file or package with debugger flags (-N -l)
func BuildDebug(targetPath string) *BuildResult {
	start := time.Now()
	res := &BuildResult{}

	lines := CountLines(targetPath)
	res.LinesCompiled = lines

	dir := filepath.Dir(targetPath)
	if dir == "" {
		dir = "."
	}

	ext := ""
	if os.PathSeparator == '\\' {
		ext = ".exe"
	}
	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("turbogo_dbg_%d%s", time.Now().UnixNano(), ext))
	res.BinaryPath = tmpBin

	cmd := exec.Command("go", "build", "-gcflags=all=-N -l", "-o", tmpBin, filepath.Base(targetPath))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.RawOutput = outBuf.String()

	if err == nil {
		res.Success = true
		res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput)
		return res
	}

	res.Success = false
	res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput)
	if res.ErrorCount == 0 && len(res.Errors) > 0 {
		res.ErrorCount = len(res.Errors)
	} else if res.ErrorCount == 0 {
		res.ErrorCount = 1
	}

	return res
}

func parseDiagnostics(rawOutput string) ([]CompileError, int, int) {
	var errs []CompileError
	errCount := 0
	warnCount := 0

	matches := errRegex.FindAllStringSubmatch(rawOutput, -1)
	for _, m := range matches {
		if len(m) == 5 {
			lineNum, _ := strconv.Atoi(m[2])
			colNum, _ := strconv.Atoi(m[3])
			msg := strings.TrimSpace(m[4])
			level := "error"
			if strings.HasPrefix(strings.ToLower(msg), "warning") {
				level = "warning"
				warnCount++
			} else {
				errCount++
			}
			errs = append(errs, CompileError{
				File:    m[1],
				Line:    lineNum,
				Column:  colNum,
				Level:   level,
				Message: msg,
			})
		}
	}
	return errs, errCount, warnCount
}

// Run executes the binary or target file
func RunBinary(binPath string, workDir string, args []string) *RunResult {
	start := time.Now()
	cmd := exec.Command(binPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return &RunResult{
		Output:    outBuf.String(),
		ExitCode:  exitCode,
		Duration:  duration,
		Completed: true,
	}
}
