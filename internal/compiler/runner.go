package compiler

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

// Error regex pattern: file.go:line:col: message (supports Windows drive letters C:\...)
var errRegex = regexp.MustCompile(`(?m)^((?:[a-zA-Z]:)?[^:\n\r]+):(\d+):(\d+):\s*(.+)$`)

// FindGoModuleRoot searches upwards from targetPath to find the directory containing go.mod
func FindGoModuleRoot(targetPath string) (string, bool) {
	dir := targetPath
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		dir = filepath.Dir(dir)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", false
	}

	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// GetPackageName reads the package clause from a Go source file
func GetPackageName(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

// CountLines counts total lines of go code in the specified target or package
func CountLines(targetPath string) int {
	total := 0
	info, err := os.Stat(targetPath)
	if err != nil {
		return 0
	}

	if info.IsDir() {
		_ = filepath.Walk(targetPath, func(path string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				content, err := os.ReadFile(path)
				if err == nil {
					total += bytes.Count(content, []byte("\n")) + 1
				}
			}
			return nil
		})
		return total
	}

	// For a file target: count all relevant .go files in its package
	absTarget, _ := filepath.Abs(targetPath)
	dir := filepath.Dir(absTarget)

	if _, hasMod := FindGoModuleRoot(absTarget); hasMod {
		// Module project: count all .go files in current package directory
		_ = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				content, err := os.ReadFile(path)
				if err == nil {
					total += bytes.Count(content, []byte("\n")) + 1
				}
			}
			return nil
		})
		if total > 0 {
			return total
		}
	}

	// Same directory multi-file package
	entries, err := os.ReadDir(dir)
	if err == nil {
		pkgName := GetPackageName(absTarget)
		fileCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
				fPath := filepath.Join(dir, entry.Name())
				if pkgName == "" || GetPackageName(fPath) == pkgName {
					content, err := os.ReadFile(fPath)
					if err == nil {
						total += bytes.Count(content, []byte("\n")) + 1
						fileCount++
					}
				}
			}
		}
		if fileCount > 0 {
			return total
		}
	}

	content, err := os.ReadFile(targetPath)
	if err == nil {
		total = bytes.Count(content, []byte("\n")) + 1
	}
	return total
}

// GetBuildArgs determines build arguments and working directory for single or multi-file projects
func GetBuildArgs(targetPath string, tmpBin string, extraFlags ...string) (workDir string, args []string) {
	absTarget, _ := filepath.Abs(targetPath)
	dir := filepath.Dir(absTarget)
	if dir == "" {
		dir = "."
	}

	_, hasMod := FindGoModuleRoot(absTarget)

	args = []string{"build"}
	args = append(args, extraFlags...)
	args = append(args, "-o", tmpBin)

	if hasMod {
		// Go module project:
		// Working directory is package directory, target is "." to build all package files
		workDir = dir
		args = append(args, ".")
		return workDir, args
	}

	// Non-module directory: collect all .go files in same directory with same package
	pkgName := GetPackageName(absTarget)
	entries, err := os.ReadDir(dir)
	if err == nil && pkgName != "" {
		var samePkgFiles []string
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
				fPath := filepath.Join(dir, name)
				if GetPackageName(fPath) == pkgName {
					samePkgFiles = append(samePkgFiles, name)
				}
			}
		}
		if len(samePkgFiles) > 1 {
			workDir = dir
			args = append(args, samePkgFiles...)
			return workDir, args
		}
	}

	// Single file fallback
	workDir = dir
	args = append(args, filepath.Base(absTarget))
	return workDir, args
}

// Build compiles the target Go file, multi-file package, or module
func Build(targetPath string) *BuildResult {
	start := time.Now()
	res := &BuildResult{}

	lines := CountLines(targetPath)
	res.LinesCompiled = lines

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("turbogo_bin_%d%s", time.Now().UnixNano(), ext))
	res.BinaryPath = tmpBin

	workDir, args := GetBuildArgs(targetPath, tmpBin)
	cmd := exec.Command("go", args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.RawOutput = outBuf.String()

	if err == nil {
		res.Success = true
		res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput, workDir)
		return res
	}

	res.Success = false
	res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput, workDir)
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

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("turbogo_dbg_%d%s", time.Now().UnixNano(), ext))
	res.BinaryPath = tmpBin

	workDir, args := GetBuildArgs(targetPath, tmpBin, "-gcflags=all=-N -l")
	cmd := exec.Command("go", args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.RawOutput = outBuf.String()

	if err == nil {
		res.Success = true
		res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput, workDir)
		return res
	}

	res.Success = false
	res.Errors, res.ErrorCount, res.WarningCount = parseDiagnostics(res.RawOutput, workDir)
	if res.ErrorCount == 0 && len(res.Errors) > 0 {
		res.ErrorCount = len(res.Errors)
	} else if res.ErrorCount == 0 {
		res.ErrorCount = 1
	}

	return res
}

func parseDiagnostics(rawOutput string, workDir string) ([]CompileError, int, int) {
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
			fPath := m[1]
			if !filepath.IsAbs(fPath) && workDir != "" {
				fPath = filepath.Clean(filepath.Join(workDir, fPath))
			}
			errs = append(errs, CompileError{
				File:    fPath,
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
