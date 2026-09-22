# Changelog

All notable changes to **Turbo Go (`tg`)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.90] - 2026-09-21

### Added
- **Submenu Mnemonic Hotkeys**: Direct single-letter execution in drop-down menus (e.g. `File` ➔ `N` New, `O` Open, `S` Save, `A` Save As; `Edit` ➔ `U` Undo, `R` Redo, `C` Copy).
- **Word-by-Word Navigation**: Move cursor word-by-word with `Ctrl + Left/Right` (and macOS `Option + Left/Right`).
- **macOS Option Meta Key Guide**: Added documentation and tips for configuring Option as Meta key in macOS Terminal.app and iTerm2.
- **MIT License**: Added official open-source MIT License.
- **Language-Separated Documentation**: Split `README.md` into dedicated English (`README.md`) and Korean (`README.ko.md`) with reciprocal navigation links.
- **Authentic Terminal Screenshots**: Embedded high-resolution native terminal screenshots in documentation.
- **Windows Defender Notice**: Added guidance for Windows SmartScreen false-positive warnings on unsigned binaries.

## [0.89] - 2026-09-18

### Added
- **Automated Multi-Platform Release CI/CD**: GitHub Actions workflow and `scripts/build_release.sh` building pure static binaries for macOS (Apple Silicon), Linux (amd64), and Windows (amd64) with SHA256 checksums.
- **Mouse Support**: Mouse click, drag block selection, and mouse wheel scrolling across editor, menu bar, and dialogs.
- **Unsaved Changes Confirmation**: Borland-style modal alert dialog prompting to save or discard changes when opening files or exiting.
- **Scratch Buffer Support**: Untitled scratch buffer enabling autocomplete, hover, and syntax styling before saving to disk.
- **Go Dark+ Palette**: Enhanced syntax and semantic token color mapping inspired by VS Code Go Dark+.
- **User Screen Live Streaming**: Stream live subprocess output to the `Alt+F5` User Screen buffer during Delve debug sessions.

## [0.88] - 2026-09-14

### Added
- **Classic Borland Turbo Vision UI**: Signature Turbo Blue canvas (`#0000A8`), double-line box borders (`╔═╗`), 3D text drop shadows, top pull-down menu bar, and bottom hotkey bar.
- **Compiling Modal Dialog**: Authentic statistics modal showing target, total lines, error counts, and elapsed build time with instant jump to error lines.
- **Multi-File & Go Module Support**: Automatic detection of `go.mod` roots and multi-file package aggregation.
- **Interactive Delve Debugger**: Breakpoint toggling (`F4`) with full-width red highlight bars, active instruction pointer (`►`) with full-width yellow bars, Step Over (`F8`), Trace Into (`F7`), and bottom **Watches Window** for real-time variable inspection.
- **Retro PC Speaker Sound Effects**: Dual-tone compilation success chime, failure buzz, and debugger step pings with audio toggle (`Options ➔ Sound`).
- **Code Intelligence (LSP)**: `gopls` integration with real-time autocompletion popup, hover documentation snippets, and semantic tokens.
- **Navigation Stack & Undo**: `F12` Go to Definition with multi-file jump history (`Alt+Left/Right` back/forward) and multi-step Undo/Redo stack.
- **Atomic File Operations**: Safe atomic saves via swap files, UTF-8 BOM auto-stripping, and non-text binary file protection.
- **Cross-Platform Clipboard**: Seamless integration with system clipboard (`Ctrl+C`, `Ctrl+X`, `Ctrl+V`, `Ctrl+A`).
