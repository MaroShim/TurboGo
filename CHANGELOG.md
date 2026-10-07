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
- **LSP Code Completion Popup**: Real-time identifier autocompletion popup with authentic Turbo Vision double-line styling.
- **Semantic Tokens Highlighting**: Hybrid syntax overlay with debounced server synchronization using VS Code Go Dark+ colors.
- **Status Bar Visual Contrast**: Enhanced status bar message visibility with high-contrast background and padded layout.

### Fixed
- **Delve Output Streaming**: Stream live subprocess output to the `Alt+F5` User Screen buffer during active debugging sessions.

## [0.80] - 2026-09-11

### Added
- **Code Navigation & Search**: Project-wide text search and `F12` Go to Definition with dedicated search results modal dialog.
- **Navigation History Stack**: Multi-file jump history with bidirectional jump navigation (`Alt+Left` / `Alt+Right`).
- **Multi-Level Undo/Redo**: Full buffer undo and redo stack supporting multi-step rollbacks.
- **LSP Client Integration**: Native `gopls` client with server connection status badge and hover documentation snippets.
- **Modular Example Suite**: Bundled modular Go example packages (`fibonacci`, `stats`, `concurrency`, and multi-file math suite).

### Refactored
- **Atomic File Safety**: Temporary swap file staging, atomic rename, automatic UTF-8 BOM stripping, and non-text binary file rejection.
- **Native Terminal Cursor**: Centralized `DrawInputField` helper and native terminal cursor rendering.
- **macOS Alt Key Dispatch**: Streamlined Alt/Meta key handling.

### Fixed
- **Delve Debugger Stability**: Synchronized breakpoint state across goroutines, isolated watch variable inspection, and dynamic breakpoint creation/removal during active sessions.
- **Just My Code Debugging**: Automatically skip Go standard library and runtime internals during `F7` Trace Into.
- **Multi-File Stepping**: Resolved multi-file breakpoint tracking and automatic active file switching.

## [0.50] - 2026-09-08

### Added
- **Multi-File Package Compilation**: Automatic detection of `go.mod` module roots and multi-file Go package build support (`go build .`).
- **Cross-Platform Clipboard**: Unified clipboard abstraction supporting system clipboard (`Ctrl+C`, `Ctrl+X`, `Ctrl+V`, `Ctrl+A`) with native escape and internal buffer fallback.

## [0.10] - 2026-09-04

### Added
- **Classic Borland Turbo Vision UI**: Signature Turbo Blue canvas (`#0000A8`), double-line box frames (`╔═╗`), 3D text drop shadows, top pull-down menu bar (`F10`), and bottom hotkey bar.
- **Go Code Editor**: Syntax highlighting for Go keywords, types, literals (strings, runes, numbers), built-in functions, and comments.
- **Compiling Modal Dialog**: Authentic statistics modal showing target file, total lines, error counts, and elapsed build time with instant jump to error lines.
- **Alt+F5 User Screen**: Dedicated full-screen console viewer to inspect program stdout/stderr and exit codes.
- **Interactive Delve Debugger**: Breakpoint toggling (`F4`) with red highlight bar, active instruction pointer (`►`) with yellow highlight bar, Step Over (`F8`), Trace Into (`F7`), and bottom **Watches Window** for real-time variable inspection.
- **Retro PC Speaker Sound Effects**: Dual-tone compilation success chime, failure buzz, and debugger step pings with audio toggle (`Options ➔ Sound`).
- **Automated Test Suite**: Unit tests for editor operations, UI components, syntax highlighting, and about dialog.

## [0.01] - 2026-09-03

### Added
- **Initial Prototype**: Proof-of-concept Turbo Vision TUI shell and Go compiler runner.
