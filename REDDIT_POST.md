# I missed the 90s Borland Turbo Pascal & C++ IDEs, so I built native Turbo TUI IDEs for Go, Rust, and FORTRAN 77

Hey r/programming!

If you learned programming in the late 80s or 90s, you probably remember the magic of **Borland Turbo Pascal 7.0** or **Turbo C++ 3.0**:
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

* **Authentic Borland UI**:
  - Classic blue screen, double-line boxes, pull-down menus (`Alt+F`, `Alt+S`, `Alt+R`, etc.), and PC speaker beeps on build/error.
* **Go to Definition (`F12`)**:
  - Put the cursor on any function, struct, type, or subroutine and press `F12` — it scans the entire multi-file project and jumps directly to the definition (auto-loading the target file if needed).
* **Project-Wide Search (`Alt+F3`)**:
  - Recursively searches all project source code with an interactive Borland-style modal dialog to browse and jump to results.
* **Zero External Dependencies (Ultra-lightweight single binary)**:
  - Built as a single static Go binary under 15MB. No Node.js, Python, or Electron runtime needed. Just drop the binary on any remote server and it launches in 0.01 seconds.

---

### Quick Start & Installation

Install directly via `go install`:

```bash
# Turbo Go
go install github.com/MaroShim/tg/cmd/tg@latest

# Turbo Rust
go install github.com/MaroShim/tr/cmd/tr@latest

# Turbo FORTRAN 77
go install github.com/MaroShim/tf77/cmd/tf77@latest
```

Or build from source:

```bash
git clone https://github.com/MaroShim/tg.git && cd tg && go build -o bin/tg ./cmd/tg
```

---

### Repositories

- **Turbo Go**: [github.com/MaroShim/tg](https://github.com/MaroShim/tg)
- **Turbo Rust**: [github.com/MaroShim/tr](https://github.com/MaroShim/tr)
- **Turbo FORTRAN 77**: [github.com/MaroShim/tf77](https://github.com/MaroShim/tf77)

*(Attaching 1–2 screenshots of the blue editor in action when posting gets great responses!)*

---

I’d love to hear feedback and memories from anyone who coded in Borland IDEs back in the day, or researchers still wrangling Fortran on supercomputers! What features or languages would you like to see next?
