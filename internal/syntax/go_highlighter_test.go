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

func TestHighlightLineEdgeCases(t *testing.T) {
	baseStyle := tcell.StyleDefault

	// 1. Empty line
	inBlock := false
	tokens := HighlightLine("", baseStyle, &inBlock)
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty line, got %d", len(tokens))
	}

	// 2. Block comment spanning lines
	inBlock = false
	line1 := "/* start comment"
	tokens1 := HighlightLine(line1, baseStyle, &inBlock)
	if !inBlock {
		t.Errorf("expected inBlock true after unclosed block comment")
	}
	fg1, _, _ := tokens1[0].Style.Decompose()
	if fg1 != ColorSyntaxComment {
		t.Errorf("expected ColorSyntaxComment for line1, got %v", fg1)
	}

	line2 := "still comment */"
	tokens2 := HighlightLine(line2, baseStyle, &inBlock)
	if inBlock {
		t.Errorf("expected inBlock false after closing */")
	}
	fg2, _, _ := tokens2[len(tokens2)-1].Style.Decompose()
	if fg2 != ColorSyntaxComment {
		t.Errorf("expected ColorSyntaxComment for line2, got %v", fg2)
	}

	// 3. Numbers: decimal, hex, float
	lineNum := "x := 42 + 0xFF + 3.14"
	tokensNum := HighlightLine(lineNum, baseStyle, &inBlock)
	// Find index of '4'
	numIdx := 5
	fgNum, _, _ := tokensNum[numIdx].Style.Decompose()
	if fgNum != ColorSyntaxNumber {
		t.Errorf("expected ColorSyntaxNumber for '42', got %v", fgNum)
	}

	// 4. Keyword embedded in identifier: 'deferment' should not be highlighted as 'defer'
	lineIdent := "deferment := true"
	tokensIdent := HighlightLine(lineIdent, baseStyle, &inBlock)
	fgIdent, _, _ := tokensIdent[0].Style.Decompose()
	if fgIdent == ColorSyntaxKeyword {
		t.Errorf("identifier 'deferment' should not be highlighted as keyword")
	}

	// 5. Raw string literal with backticks
	lineRaw := "msg := `hello\nworld`"
	tokensRaw := HighlightLine(lineRaw, baseStyle, &inBlock)
	// Find index of '`'
	backtickIdx := 7
	fgRaw, _, _ := tokensRaw[backtickIdx].Style.Decompose()
	if fgRaw != ColorSyntaxString {
		t.Errorf("expected ColorSyntaxString for raw string, got %v", fgRaw)
	}
}
