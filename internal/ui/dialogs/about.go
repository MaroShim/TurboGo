package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"tg/internal/ui"
)

// AboutDialog shows the retro Borland About information
type AboutDialog struct {
	Visible bool
}

func NewAboutDialog() *AboutDialog {
	return &AboutDialog{Visible: false}
}

func (a *AboutDialog) Show() {
	a.Visible = true
}

func (a *AboutDialog) Hide() {
	a.Visible = false
}

func (a *AboutDialog) Draw(screen tcell.Screen, screenW, screenH int) {
	if !a.Visible {
		return
	}

	dialogW := 48
	dialogH := 14
	x := (screenW - dialogW) / 2
	y := (screenH - dialogH) / 2

	ui.DrawDialogBox(screen, x, y, dialogW, dialogH, "About")

	textStyle := tcell.StyleDefault.Background(ui.ColorDialogBg).Foreground(ui.ColorDialogFg)
	titleStyle := textStyle.Foreground(tcell.ColorMaroon).Bold(true)
	accentStyle := textStyle.Foreground(tcell.ColorNavy).Bold(true)

	lines := []struct {
		text  string
		style tcell.Style
	}{
		{"╔══════════════════════════════════════════╗", accentStyle},
		{"║          TURBO GO  Version 1.0           ║", titleStyle},
		{"║     Retro IDE for The Go Language        ║", accentStyle},
		{"╚══════════════════════════════════════════╝", accentStyle},
		{"", textStyle},
		{"Inspired by Borland Turbo Pascal & Turbo C", textStyle},
		{"Integrated with Go Toolchain & Delve", textStyle},
		{"Retro TUI Powered by tcell/v2", textStyle},
	}

	for i, l := range lines {
		px := x + (dialogW-len([]rune(l.text)))/2
		for j, r := range l.text {
			screen.SetContent(px+j, y+2+i, r, nil, l.style)
		}
	}

	ui.DrawButton(screen, x+(dialogW-12)/2, y+dialogH-3, 12, "OK", true)
}
