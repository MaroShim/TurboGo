# I missed the 90s Borland Turbo Pascal & C++ IDEs, so I built native Turbo TUI IDEs for Go, Rust, and FORTRAN 77

Hey r/programming!

If you learned programming in the late 80s or 90s, you probably remember the magic of **Borland Turbo Pascal** or **Turbo C/C++**:
- That iconic dark blue background with crisp yellow text and menus
- Instant sub-second compilation at the press of `Ctrl+F9`
- Pull-down menus, dialog boxes, and mouse support that worked seamlessly in a text window
- A development environment that popped up in 0.1 seconds the moment you launched it

Modern IDEs like VS Code are great, but whenever I work on remote cloud VMs, Docker containers, or headless supercomputer clusters over SSH, I keep running into the same friction:
- **VS Code Remote-SSH** runs a memory-hungry Node.js server in the background, often crashing low-spec instances (like 1GB RAM VMs) with OOM errors.
- **Neovim** is fantastic, but managing dozens of plugins and complex Lua configs across many remote machines gets exhausting.
- **Nano / Micro** are too barebones when you need to navigate multiple files, jump to definitions, or inspect compiler errors.

So I wanted to bring back the lightning-fast, nostalgic experience of the old Borland Turbo series, and built **the Turbo TUI Trilogy** — standalone, zero-dependency TUI IDEs for both modern and legacy languages:

---

### The Trilogy:

#### 1. `tg` — Turbo Go
* **Background**: Go directly shares architectural DNA and heritage with Pascal/Modula.
* **Key Highlight**: Go's compiler is famously fast. In `tg`, hitting compile brings back that authentic "0.1-second instant build" joy from Turbo Pascal.
* **Best for**: Writing CLI tools, backend microservices, and lag-free remote development over SSH.

#### 2. `tr` — Turbo Rust
* **Background**: Turbo C++ vibes meet modern systems programming.
* **Key Highlight**: Integrates directly with `cargo check` and `cargo build`. When the compiler emits errors, `tr` parses them into a classic Borland modal error list dialog. Just press `Enter` to jump straight to the file, line, and column.
* **Best for**: Systems programming, learning Rust, and algorithmic problem solving in a clean, distraction-free environment.

#### 3. `tf77` — Turbo FORTRAN 77
* **Background**: A dedicated project for researchers in HPC, physics, and scientific simulation.
* **Key Highlight**: Millions of lines of mission-critical scientific code (weather forecasting, aerospace CFD, molecular dynamics, nuclear physics) still run in fixed-form FORTRAN 77 on headless supercomputers. Modern editors are notoriously awkward with F77 fixed-form.
* **F77-Specific Features**:
  - **Column 72 guide line** (statements beyond column 72 are ignored or cause errors).
  - **Column 6 continuation line indicator**.
  - Project-wide case-insensitive `SUBROUTINE`, `FUNCTION`, and `PROGRAM` definition jumping.
* **Best for**: Grad students, physicists, and engineers maintaining legacy simulation codes directly on remote cluster login nodes where VS Code is banned or too heavy.

---

### Common Core Features Across All Three:

* **Authentic Borland Turbo Vision UI**:
  - Classic Turbo Blue screen (`#0000A8`), double-line box frames (`╔═╗`), top pull-down menus with hotkey mnemonics, drop shadows, and PC speaker sound effects on build success/failure.
* **Interactive Debugger & Real-time Watches Window**:
  - `F4` to toggle breakpoints (`●`), `F5` start/continue, `F8` step over, `F7` trace into.
  - Active execution line highlighted with a solid yellow bar.
  - Built-in bottom **Watches Window** for live inspection of local variables, types, and values (`delve` for Go, `gdb/lldb` for Rust, and native internal/lldb engine for Fortran).
* **Go to Definition (`F12`)**:
  - Put the cursor on any function, struct, type, or subroutine and press `F12` — it scans the entire multi-file project and jumps directly to the definition (auto-loading the target file if needed).
* **Alt+F5 User Screen**:
  - The iconic Turbo C feature: switch to a full-screen DOS-style console view to inspect raw execution output, and return to the IDE with any keypress.
* **Modern Navigation & Editing**:
  - Mnemonic single-letter hotkeys in dropdown menus (e.g. `File` ➔ `N` New, `O` Open, `S` Save, `A` Save As).
  - Word-by-word cursor movement and block selection via `Ctrl+Left/Right` and macOS `Option+Left/Right`.
* **Zero External Dependencies (Ultra-lightweight static binary)**:
  - Built as a single static Go binary (~10-15MB). No Node.js, Python, or Electron runtime needed. Just drop the binary on any remote server or VM and it launches in 0.01 seconds.

---

### Quick Start & Installation

#### 1. Download Pre-built Binaries (GitHub Releases)
You can directly download pre-compiled standalone executables for **macOS (Apple Silicon)**, **Linux (x86_64)**, and **Windows (x64)** from the Releases page:
- [Turbo Go Releases (v0.90)](https://github.com/MaroShim/TurboGo/releases/latest)
- [Turbo Rust Releases (v0.90)](https://github.com/MaroShim/TurboRust/releases/latest)
- [Turbo Fortran Releases (v0.90)](https://github.com/MaroShim/TurboF77/releases/latest)

#### 2. Install directly via `go install`:

```bash
# Turbo Go
go install github.com/MaroShim/tg/cmd/tg@latest

# Turbo Rust
go install github.com/MaroShim/TurboRust/cmd/tr@latest

# Turbo Fortran (Classic F77 & Modern F90+)
go install github.com/MaroShim/tf77/cmd/tf77@latest
go install github.com/MaroShim/tf77/cmd/tf@latest
```

#### 3. Or build from source:

```bash
git clone https://github.com/MaroShim/TurboGo.git && cd TurboGo && go build -o bin/tg ./cmd/tg
```

---

### Repositories

- **Turbo Go**: [github.com/MaroShim/TurboGo](https://github.com/MaroShim/TurboGo)
- **Turbo Rust**: [github.com/MaroShim/TurboRust](https://github.com/MaroShim/TurboRust)
- **Turbo Fortran**: [github.com/MaroShim/TurboF77](https://github.com/MaroShim/TurboF77)

---

I’d love to hear feedback and memories from anyone who coded in Borland IDEs back in the day, or researchers still wrangling Fortran on supercomputers! What features or languages would you like to see next?
