package syntax

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHighlightLine(t *testing.T) {
	baseStyle := tcell.StyleDefault
	line := `func hello() string { return "world" } // comment`

	inBlockComment := false
	tokens := HighlightLine(line, baseStyle, &inBlockComment)

	if len(tokens) != len([]rune(line)) {
		t.Fatalf("expected token count %d, got %d", len([]rune(line)), len(tokens))
	}

	// First 4 runes should be "func" with keyword color
	funcTok := tokens[0]
	fg, _, _ := funcTok.Style.Decompose()
	if fg != ColorSyntaxKeyword {
		t.Errorf("expected func to have ColorSyntaxKeyword, got %v", fg)
	}

	// Check string literal "world"
	strIdx := -1
	for i, r := range []rune(line) {
		if r == '"' {
			strIdx = i
			break
		}
	}
	if strIdx != -1 {
		fgStr, _, _ := tokens[strIdx].Style.Decompose()
		if fgStr != ColorSyntaxString {
			t.Errorf("expected string to have ColorSyntaxString, got %v", fgStr)
		}
	}

	// Check comment at the end
	commIdx := len(tokens) - 1
	fgComm, _, _ := tokens[commIdx].Style.Decompose()
	if fgComm != ColorSyntaxComment {
		t.Errorf("expected comment to have ColorSyntaxComment, got %v", fgComm)
	}
}
