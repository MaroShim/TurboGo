package syntax

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var (
	ColorSyntaxKeyword = tcell.ColorYellow
	ColorSyntaxType    = tcell.ColorLightCyan
	ColorSyntaxString  = tcell.ColorLightCyan
	ColorSyntaxNumber  = tcell.ColorLightGreen
	ColorSyntaxComment = tcell.NewHexColor(0x808080) // Gray
	ColorSyntaxBuiltin = tcell.ColorGreen
	ColorSyntaxNormal  = tcell.ColorWhite
)

var goKeywords = map[string]bool{
	"break": true, "default": true, "func": true, "interface": true, "select": true,
	"case": true, "defer": true, "go": true, "map": true, "struct": true,
	"chan": true, "else": true, "goto": true, "package": true, "switch": true,
	"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
	"continue": true, "for": true, "import": true, "return": true, "var": true,
}

var goTypes = map[string]bool{
	"bool": true, "byte": true, "complex64": true, "complex128": true,
	"error": true, "float32": true, "float64": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"rune": true, "string": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
	"any": true,
}

var goConstants = map[string]bool{
	"true": true, "false": true, "iota": true, "nil": true,
}

var goBuiltins = map[string]bool{
	"append": true, "cap": true, "close": true, "complex": true, "copy": true,
	"delete": true, "imag": true, "len": true, "make": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true, "recover": true,
}

// Token represents a styled span of runes on a line
type Token struct {
	Style tcell.Style
	Char  rune
}

// HighlightLine parses a line of Go code and returns a slice of styled runes
func HighlightLine(line string, baseStyle tcell.Style, inBlockComment *bool) []Token {
	runes := []rune(line)
	n := len(runes)
	tokens := make([]Token, n)

	keywordStyle := baseStyle.Foreground(ColorSyntaxKeyword).Bold(true)
	typeStyle := baseStyle.Foreground(ColorSyntaxType)
	stringStyle := baseStyle.Foreground(ColorSyntaxString)
	numberStyle := baseStyle.Foreground(ColorSyntaxNumber)
	commentStyle := baseStyle.Foreground(ColorSyntaxComment)
	builtinStyle := baseStyle.Foreground(ColorSyntaxBuiltin)
	constStyle := baseStyle.Foreground(ColorSyntaxNumber).Bold(true)

	i := 0
	for i < n {
		// Handle block comments across lines
		if inBlockComment != nil && *inBlockComment {
			if i+1 < n && runes[i] == '*' && runes[i+1] == '/' {
				tokens[i] = Token{Style: commentStyle, Char: runes[i]}
				tokens[i+1] = Token{Style: commentStyle, Char: runes[i+1]}
				i += 2
				*inBlockComment = false
				continue
			}
			tokens[i] = Token{Style: commentStyle, Char: runes[i]}
			i++
			continue
		}

		// Single line comment
		if i+1 < n && runes[i] == '/' && runes[i+1] == '/' {
			for i < n {
				tokens[i] = Token{Style: commentStyle, Char: runes[i]}
				i++
			}
			break
		}

		// Block comment start
		if i+1 < n && runes[i] == '/' && runes[i+1] == '*' {
			tokens[i] = Token{Style: commentStyle, Char: runes[i]}
			tokens[i+1] = Token{Style: commentStyle, Char: runes[i+1]}
			i += 2
			if inBlockComment != nil {
				*inBlockComment = true
			}
			continue
		}

		// String literal: double quote
		if runes[i] == '"' {
			tokens[i] = Token{Style: stringStyle, Char: runes[i]}
			i++
			escaped := false
			for i < n {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				if !escaped && runes[i] == '"' {
					i++
					break
				}
				if runes[i] == '\\' && !escaped {
					escaped = true
				} else {
					escaped = false
				}
				i++
			}
			continue
		}

		// Raw string literal: backtick
		if runes[i] == '`' {
			tokens[i] = Token{Style: stringStyle, Char: runes[i]}
			i++
			for i < n {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				if runes[i] == '`' {
					i++
					break
				}
				i++
			}
			continue
		}

		// Rune literal: single quote
		if runes[i] == '\'' {
			tokens[i] = Token{Style: stringStyle, Char: runes[i]}
			i++
			escaped := false
			for i < n {
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				if !escaped && runes[i] == '\'' {
					i++
					break
				}
				if runes[i] == '\\' && !escaped {
					escaped = true
				} else {
					escaped = false
				}
				i++
			}
			continue
		}

		// Numbers
		if unicode.IsDigit(runes[i]) || (runes[i] == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '.' || runes[i] == '_') {
				tokens[i] = Token{Style: numberStyle, Char: runes[i]}
				i++
			}
			continue
		}

		// Identifiers and Keywords
		if unicode.IsLetter(runes[i]) || runes[i] == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			word := string(runes[start:i])
			var style tcell.Style
			if goKeywords[word] {
				style = keywordStyle
			} else if goTypes[word] {
				style = typeStyle
			} else if goConstants[word] {
				style = constStyle
			} else if goBuiltins[word] {
				style = builtinStyle
			} else {
				// Check if followed by '(' -> function call
				nextNonSpace := i
				for nextNonSpace < n && unicode.IsSpace(runes[nextNonSpace]) {
					nextNonSpace++
				}
				if nextNonSpace < n && runes[nextNonSpace] == '(' {
					style = baseStyle.Foreground(tcell.ColorWhite).Bold(true)
				} else {
					style = baseStyle.Foreground(ColorSyntaxNormal)
				}
			}

			for k := start; k < i; k++ {
				tokens[k] = Token{Style: style, Char: runes[k]}
			}
			continue
		}

		// Default punctuation / symbol / whitespace
		tokens[i] = Token{Style: baseStyle.Foreground(ColorSyntaxNormal), Char: runes[i]}
		i++
	}

	return tokens
}
