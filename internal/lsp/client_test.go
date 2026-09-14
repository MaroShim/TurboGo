package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPathURIRoundtrip(t *testing.T) {
	testPaths := []string{
		"/Users/maro/Projects/tg/main.go",
		"/tmp/test_file.go",
		"/var/log/test.go",
	}

	for _, p := range testPaths {
		uri := PathToURI(p)
		if !strings.HasPrefix(uri, "file://") {
			t.Fatalf("expected URI prefix file://, got %s", uri)
		}
		back := URIToPath(uri)
		if back != p {
			t.Errorf("expected %s, got %s", p, back)
		}
	}
}

func TestHoverSnippetCleaning(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "```go\nfunc Println(a ...any) (n int, err error)\n```\nPrintln formats using the default formats",
			expected: "func Println(a ...any) (n int, err error)",
		},
		{
			input:    "```rust\npub fn calculate(val: i32) -> i32\n```",
			expected: "pub fn calculate(val: i32) -> i32",
		},
		{
			input:    "type MyStruct struct",
			expected: "type MyStruct struct",
		},
	}

	for _, tc := range tests {
		got := cleanHoverSnippet(tc.input)
		if got != tc.expected {
			t.Errorf("cleanHoverSnippet(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestMockLSPClientInteraction(t *testing.T) {
	// Create piped readers and writers to emulate a full LSP server process
	serverInR, clientInW := io.Pipe()   // client writes to clientInW, server reads from serverInR
	clientOutR, serverOutW := io.Pipe() // server writes to serverOutW, client reads from clientOutR

	var wg sync.WaitGroup
	wg.Add(1)

	// Mock Server Goroutine
	go func() {
		defer wg.Done()
		defer serverOutW.Close()
		defer serverInR.Close()

		reader := bufio.NewReader(serverInR)
		for {
			// Read framing
			contentLength := -1
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					break
				}
				if strings.HasPrefix(strings.ToLower(line), "content-length:") {
					val := strings.TrimSpace(line[len("content-length:"):])
					contentLength, _ = strconv.Atoi(val)
				}
			}
			if contentLength <= 0 {
				continue
			}

			body := make([]byte, contentLength)
			if _, err := io.ReadFull(reader, body); err != nil {
				return
			}

			var req struct {
				ID     int64           `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			_ = json.Unmarshal(body, &req)

			// Route responses
			var result interface{}
			switch req.Method {
			case "initialize":
				result = map[string]interface{}{
					"capabilities": map[string]interface{}{
						"definitionProvider": true,
						"hoverProvider":      true,
					},
				}
			case "textDocument/definition":
				result = map[string]interface{}{
					"uri": "file:///workspace/target.go",
					"range": map[string]interface{}{
						"start": map[string]interface{}{"line": 15, "character": 4},
						"end":   map[string]interface{}{"line": 15, "character": 12},
					},
				}
			case "textDocument/hover":
				result = map[string]interface{}{
					"contents": map[string]interface{}{
						"kind":  "markdown",
						"value": "```go\nfunc TargetFunction() bool\n```",
					},
				}
			case "shutdown":
				result = nil
			case "exit":
				return
			default:
				// Ignore notifications like initialized, didOpen, didChange
				continue
			}

			resJSON, _ := json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  result,
			})
			header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(resJSON))
			_, _ = serverOutW.Write([]byte(header))
			_, _ = serverOutW.Write(resJSON)
		}
	}()

	// Construct Client wired to the mock pipes
	client := &Client{
		serverName: "mock-gopls",
		binPath:    "/mock/gopls",
		rootDir:    "/workspace",
		stdin:      clientInW,
		stdout:     clientOutR,
		reader:     bufio.NewReader(clientOutR),
		pending:    make(map[int64]chan *rpcResponse),
	}

	client.closeWg.Add(1)
	go client.readLoop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Initialize
	if err := client.initialize(ctx); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	client.isReady.Store(true)

	if !client.IsAvailable() {
		t.Fatalf("expected client to be available")
	}

	// 2. DidOpen & DidChange
	if err := client.DidOpen("/workspace/main.go", "package main"); err != nil {
		t.Errorf("DidOpen failed: %v", err)
	}
	if err := client.DidChange("/workspace/main.go", "package main\nfunc main() {}"); err != nil {
		t.Errorf("DidChange failed: %v", err)
	}

	// 3. Definition query
	targetFile, line, col, found := client.Definition(ctx, "/workspace/main.go", 5, 2)
	if !found {
		t.Fatalf("expected definition to be found")
	}
	if targetFile != "/workspace/target.go" || line != 16 || col != 5 {
		t.Errorf("expected (/workspace/target.go, 16, 5), got (%s, %d, %d)", targetFile, line, col)
	}

	// 4. Hover query
	hoverSnippet, found := client.Hover(ctx, "/workspace/main.go", 5, 2)
	if !found {
		t.Fatalf("expected hover to be found")
	}
	if hoverSnippet != "func TargetFunction() bool" {
		t.Errorf("expected 'func TargetFunction() bool', got %q", hoverSnippet)
	}

	// 5. Clean teardown
	client.isClosed.Store(true)
	_ = clientInW.Close()
	_ = clientOutR.Close()
	wg.Wait()
}

func TestDecodeSemanticTokens(t *testing.T) {
	legend := []string{
		"type", "class", "enum", "interface", "struct", "typeParameter",
		"parameter", "variable", "property", "enumMember", "function",
	}

	// Encodes:
	// Token 1: line 2, col 5, len 8, type "function" (idx 10), mod 0
	// Token 2: line 2, col 14 (deltaCol 9), len 4, type "parameter" (idx 6), mod 0
	// Token 3: line 5 (deltaLine 3), col 1, len 6, type "type" (idx 0), mod 1
	data := []uint32{
		2, 5, 8, 10, 0,
		0, 9, 4, 6, 0,
		3, 1, 6, 0, 1,
	}

	spans := DecodeSemanticTokens(data, legend)
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}

	// Span 1: line 2, col 5, len 8, type "function"
	if spans[0].Line != 2 || spans[0].StartCol != 5 || spans[0].Length != 8 || spans[0].TokenType != "function" {
		t.Errorf("span[0] mismatch: %+v", spans[0])
	}

	// Span 2: line 2, col 14, len 4, type "parameter"
	if spans[1].Line != 2 || spans[1].StartCol != 14 || spans[1].Length != 4 || spans[1].TokenType != "parameter" {
		t.Errorf("span[1] mismatch: %+v", spans[1])
	}

	// Span 3: line 5, col 1, len 6, type "type"
	if spans[2].Line != 5 || spans[2].StartCol != 1 || spans[2].Length != 6 || spans[2].TokenType != "type" || spans[2].TokenModifiers != 1 {
		t.Errorf("span[2] mismatch: %+v", spans[2])
	}
}

