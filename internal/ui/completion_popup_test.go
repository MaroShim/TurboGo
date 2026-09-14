package ui

import (
	"testing"

	"tg/internal/lsp"
)

func TestCompletionPopupNavigationAndFiltering(t *testing.T) {
	popup := NewCompletionPopup()
	if popup.IsVisible() {
		t.Fatalf("expected popup to start hidden")
	}

	items := []lsp.CompletionItem{
		{Label: "Println", Kind: lsp.CompletionKindFunction, Detail: "func(a ...any)"},
		{Label: "Printf", Kind: lsp.CompletionKindFunction, Detail: "func(format string, a ...any)"},
		{Label: "Sprint", Kind: lsp.CompletionKindFunction, Detail: "func(a ...any) string"},
		{Label: "Sprintf", Kind: lsp.CompletionKindFunction, Detail: "func(format string, a ...any) string"},
	}

	popup.Show(items, 4, 10, 5, 80, 24)
	if !popup.IsVisible() {
		t.Fatalf("expected popup to be visible")
	}
	if len(popup.Filtered) != 4 {
		t.Fatalf("expected 4 filtered items, got %d", len(popup.Filtered))
	}

	// 1. Initial selection
	sel := popup.GetSelected()
	if sel == nil || sel.Label != "Println" {
		t.Fatalf("expected Println selected, got %+v", sel)
	}

	// 2. Down navigation
	popup.MoveDown()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "Printf" {
		t.Fatalf("expected Printf selected, got %+v", sel)
	}

	// 3. Up navigation back to 0
	popup.MoveUp()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "Println" {
		t.Fatalf("expected Println selected, got %+v", sel)
	}

	// 4. Wrap-around Up navigation (Rule 52)
	popup.MoveUp()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "Sprintf" {
		t.Fatalf("expected Sprintf selected on wrap-around, got %+v", sel)
	}

	// 5. Wrap-around Down navigation
	popup.MoveDown()
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "Println" {
		t.Fatalf("expected Println selected on wrap-around, got %+v", sel)
	}

	// 6. Filtering with prefix "sp"
	popup.SetFilter("sp")
	if len(popup.Filtered) != 2 {
		t.Fatalf("expected 2 items matching 'sp', got %d", len(popup.Filtered))
	}
	sel = popup.GetSelected()
	if sel == nil || sel.Label != "Sprint" {
		t.Fatalf("expected Sprint selected, got %+v", sel)
	}

	// 7. Filtering with non-matching string hides popup
	popup.SetFilter("zzzz")
	if popup.IsVisible() {
		t.Fatalf("expected popup to hide on zero matches")
	}
}

func TestEditorCompletionHelpers(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{"fmt.Prin"}
	ed.CursorY = 0
	ed.CursorX = 8 // at end of "Prin"

	pref, startCol := ed.GetWordPrefixAtCursor()
	if pref != "Prin" || startCol != 4 {
		t.Fatalf("GetWordPrefixAtCursor: got pref=%q, startCol=%d; want pref='Prin', startCol=4", pref, startCol)
	}

	ed.ApplyCompletion(startCol, "Println")
	if ed.Lines[0] != "fmt.Println" {
		t.Fatalf("ApplyCompletion: expected 'fmt.Println', got %q", ed.Lines[0])
	}
	if ed.CursorX != 11 {
		t.Fatalf("ApplyCompletion: expected CursorX=11, got %d", ed.CursorX)
	}

	// Undo rollback verification
	if !ed.Undo() {
		t.Fatalf("Undo failed")
	}
	if ed.Lines[0] != "fmt.Prin" {
		t.Fatalf("expected rollback to 'fmt.Prin', got %q", ed.Lines[0])
	}
}
