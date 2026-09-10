package compiler

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchMatch represents a single search match in project
type SearchMatch struct {
	File    string
	Line    int
	Column  int
	Snippet string
}

// GetSearchRootDir determines the root directory to search (module root or file directory)
func GetSearchRootDir(targetPath string) string {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}
	if modRoot, hasMod := FindGoModuleRoot(absTarget); hasMod {
		return modRoot
	}
	fi, err := os.Stat(absTarget)
	if err == nil && !fi.IsDir() {
		return filepath.Dir(absTarget)
	}
	return absTarget
}

// CollectGoFiles gathers all .go files within rootDir (skipping .git, bin, vendor, hidden)
func CollectGoFiles(rootDir string) []string {
	var files []string
	_ = filepath.Walk(rootDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		name := fi.Name()
		if fi.IsDir() {
			if strings.HasPrefix(name, ".") || name == "bin" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(name, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// FindDefinitionInProject searches for the definition of symbol (func, type, const, var)
func FindDefinitionInProject(targetPath, symbol string) (string, int, int, bool) {
	if symbol == "" {
		return "", 0, 0, false
	}

	rootDir := GetSearchRootDir(targetPath)
	files := CollectGoFiles(rootDir)

	// Prefer current file first
	absTarget, _ := filepath.Abs(targetPath)
	orderedFiles := make([]string, 0, len(files))
	for _, f := range files {
		if f == absTarget {
			orderedFiles = append([]string{f}, orderedFiles...)
		} else {
			orderedFiles = append(orderedFiles, f)
		}
	}

	// Definition patterns
	// 1. func MyFunc( or func (r *Recv) MyFunc(
	escaped := regexp.QuoteMeta(symbol)
	funcPattern := regexp.MustCompile(`(?m)^\s*func\s+(?:\([^)]+\)\s+)?` + escaped + `\s*[\(\[]`)
	// 2. type MyType struct/interface/...
	typePattern := regexp.MustCompile(`(?m)^\s*type\s+` + escaped + `\s+`)
	// 3. const MyConst = or var MyVar =
	varPattern := regexp.MustCompile(`(?m)^\s*(?:const|var)\s+` + escaped + `\b`)

	patterns := []*regexp.Regexp{funcPattern, typePattern, varPattern}

	for _, p := range patterns {
		for _, file := range orderedFiles {
			f, err := os.Open(file)
			if err != nil {
				continue
			}
			scanner := bufio.NewScanner(f)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				loc := p.FindStringIndex(line)
				if loc != nil {
					f.Close()
					col := loc[0] + 1
					// Find exact symbol start column inside match
					symIdx := strings.Index(line[loc[0]:loc[1]], symbol)
					if symIdx >= 0 {
						col = loc[0] + symIdx + 1
					}
					return file, lineNum, col, true
				}
			}
			f.Close()
		}
	}

	return "", 0, 0, false
}

// SearchInProject searches for query in all .go files in project
func SearchInProject(targetPath, query string, caseSensitive bool) []SearchMatch {
	if query == "" {
		return nil
	}

	rootDir := GetSearchRootDir(targetPath)
	files := CollectGoFiles(rootDir)

	var matches []SearchMatch
	qLower := strings.ToLower(query)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			targetLine := line
			searchQ := query
			if !caseSensitive {
				targetLine = strings.ToLower(line)
				searchQ = qLower
			}

			colIdx := strings.Index(targetLine, searchQ)
			if colIdx >= 0 {
				colRunes := len([]rune(line[:colIdx])) + 1
				snippet := strings.TrimSpace(line)
				matches = append(matches, SearchMatch{
					File:    file,
					Line:    lineNum,
					Column:  colRunes,
					Snippet: snippet,
				})
			}
		}
		f.Close()
	}

	return matches
}

// FormatMatchDisplay formats a SearchMatch for Borland list dialogs
func FormatMatchDisplay(m SearchMatch, rootDir string) string {
	relPath, err := filepath.Rel(rootDir, m.File)
	if err != nil {
		relPath = filepath.Base(m.File)
	}
	return relPath
}
