# Shedoc v1 — Go Implementation Plan

## Overview

Build a Go CLI tool and library that parses shedoc-annotated shell scripts into structured data and generates useful outputs (JSON, help text, man pages, shell completions).

## Module Identity

```
module github.com/nickawilliams/shedoc
go 1.24
```

Binary name: `shedoc`

### Cleanup

Remove existing bash tooling (`libexec/`, `scripts/`, `deps/`, `package.sh`, and the
current `Makefile`) before starting. The Go implementation replaces all of it.

## Project Structure

```
shedoc/
├── cmd/
│   └── shedoc/
│       └── main.go                    # CLI entry point, version vars
├── internal/
│   ├── cli/
│   │   ├── root.go                    # Root cobra command (parse + format output)
│   │   └── complete.go               # `shedoc complete` subcommand
│   └── generate/
│       ├── helptext.go                # --help output renderer
│       ├── helptext_test.go
│       ├── manpage.go                 # troff/groff man page renderer
│       ├── manpage_test.go
│       ├── completions_bash.go        # bash completion script generator
│       ├── completions_zsh.go         # zsh completion script generator
│       ├── completions_fish.go        # fish completion script generator
│       └── completions_test.go
├── testdata/
│   ├── comprehensive.sh              # deploy example from README
│   ├── comprehensive.json            # golden parse output
│   ├── standalone.sh                  # greet (standalone command)
│   ├── standalone.json
│   ├── library.sh                     # string-utils (sourced library)
│   ├── library.json
│   ├── minimal.sh                     # just a shebang, no shedoc
│   ├── minimal.json
│   ├── multiline.sh                   # multi-line tag descriptions
│   ├── multiline.json
│   ├── subcommands.sh                 # subcommand-heavy test case
│   ├── subcommands.json
│   ├── edge_cases.sh                  # bare #@/, empty blocks, etc.
│   ├── edge_cases.json
│   └── no_shedoc.sh                   # plain script, zero shedoc
├── model.go                           # Public data model (Document, Block, etc.)
├── formatter.go                       # Formatter interface and registry
├── parser.go                          # Line-by-line state machine parser
├── parser_test.go
├── tag.go                             # Tag-specific parsers (@flag, @option, etc.)
├── tag_test.go
├── value.go                           # Value notation parser (<required>, [opt], etc.)
├── value_test.go
├── doc.go                             # Package-level godoc
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yml
├── README.md
├── ROADMAP.md
└── LICENSE.md
```

### Key Structural Decisions

- **Library at module root** — import path is `github.com/nickawilliams/shedoc`, package name is `shedoc`. Clean and idiomatic (`shedoc.Parse()`, `shedoc.Document{}`).
- **`internal/generate/`** — generators are internal. Third parties use the parser library; they don't import our renderers. Can be promoted to public later if demand warrants.
- **`internal/cli/`** — all CLI wiring is internal. No external dependency on our command structure.
- **`testdata/`** at project root — shared across parser and generator tests.
- **Zero external deps for the library** — only `bufio`, `strings`, `regexp`, `io`, `encoding/json`. The `cmd/` and `internal/cli/` layers depend on `cobra`.

---

## Data Model

All structs in `model.go`. This is the public API surface — what `Parse()` returns and what generators consume.

```go
package shedoc

// Document is the top-level parse result for a single shell script file.
type Document struct {
    Path     string    `json:"path,omitempty"`
    Shebang  string    `json:"shebang,omitempty"`
    Meta     Meta      `json:"meta"`
    Blocks   []Block   `json:"blocks,omitempty"`
    Warnings []Warning `json:"warnings,omitempty"`
}

// Meta holds file-level metadata from #?/ shedoc tags.
type Meta struct {
    Name        string `json:"name,omitempty"`
    Version     string `json:"version,omitempty"`
    Synopsis    string `json:"synopsis,omitempty"`
    Description string `json:"description,omitempty"`
    Examples    string `json:"examples,omitempty"`
    Section     string `json:"section,omitempty"`
    Author      string `json:"author,omitempty"`
    License     string `json:"license,omitempty"`
}

// Visibility represents the access level of a documented block.
type Visibility string

const (
    VisibilityCommand    Visibility = "command"
    VisibilitySubcommand Visibility = "subcommand"
    VisibilityPublic     Visibility = "public"
    VisibilityPrivate    Visibility = "private"
)

// Block represents a single sheblock (#@/) documentation entry.
type Block struct {
    Visibility   Visibility  `json:"visibility"`
    Name         string      `json:"name,omitempty"`         // subcommand name (visibility=subcommand) or inferred name
    Description  string      `json:"description,omitempty"`
    FunctionName string      `json:"functionName,omitempty"` // shell function following this block, if any
    Line         int         `json:"line"`

    // Inputs
    Flags    []Flag    `json:"flags,omitempty"`
    Options  []Option  `json:"options,omitempty"`
    Operands []Operand `json:"operands,omitempty"`
    Env      []Env     `json:"env,omitempty"`
    Reads    []Reads   `json:"reads,omitempty"`
    Stdin    *Stdin    `json:"stdin,omitempty"`

    // Outputs
    Exit   []Exit   `json:"exit,omitempty"`
    Stdout *Stdout  `json:"stdout,omitempty"`
    Stderr *Stderr  `json:"stderr,omitempty"`
    Sets   []Sets   `json:"sets,omitempty"`
    Writes []Writes `json:"writes,omitempty"`

    // Metadata
    Deprecated *Deprecated `json:"deprecated,omitempty"`
}

// Flag represents a boolean flag: @flag -s | --long description
type Flag struct {
    Short       string `json:"short,omitempty"`
    Long        string `json:"long,omitempty"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Option represents an option with a value: @option -f | --format <value> description
type Option struct {
    Short       string `json:"short,omitempty"`
    Long        string `json:"long,omitempty"`
    Value       Value  `json:"value"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Operand represents a positional argument: @operand <name> description
type Operand struct {
    Value       Value  `json:"value"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Value represents parsed value notation: <required>, [optional], [opt=default], <var...>
type Value struct {
    Name     string `json:"name"`
    Required bool   `json:"required"`
    Default  string `json:"default,omitempty"`
    Variadic bool   `json:"variadic,omitempty"`
}

// Env represents an environment variable: @env VAR_NAME description
type Env struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Reads represents an implicit file read: @reads <path> description
type Reads struct {
    Path        string `json:"path"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Stdin represents standard input: @stdin description
type Stdin struct {
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Exit represents an exit status: @exit <code> description
type Exit struct {
    Code        string `json:"code"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Stdout represents standard output: @stdout description
type Stdout struct {
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Stderr represents standard error: @stderr description
type Stderr struct {
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Sets represents an environment variable set: @sets VAR_NAME description
type Sets struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Writes represents an implicit file write: @writes <path> description
type Writes struct {
    Path        string `json:"path"`
    Description string `json:"description,omitempty"`
    Line        int    `json:"line"`
}

// Deprecated marks a block as deprecated: @deprecated [message]
type Deprecated struct {
    Message string `json:"message,omitempty"`
    Line    int    `json:"line"`
}

// Warning represents a non-fatal parse issue.
type Warning struct {
    Line    int    `json:"line"`
    Message string `json:"message"`
}
```

### Design Rationale

- **`Stdin`, `Stdout`, `Stderr` are pointers** — at most one per block; nil vs present-but-empty is meaningful.
- **`Flags`, `Options`, `Operands` are slices** — multiple allowed, order matters (operand position is significant).
- **`Line` on every element** — supports linting, error reporting, and editor integration.
- **`FunctionName`** — captured by scanning for a function declaration after a block closes. Critical for associating blocks with code.
- **`Value` is a struct, not a string** — generators don't re-parse `<name...>`; they check `.Required` and `.Variadic` directly.
- **`Warnings` on Document** — parser is lenient by default. Malformed tags produce warnings, not errors. Hard errors only for I/O failures.
- **JSON tags use camelCase** — standard JSON convention.

---

## Formatter Interface

All output formats implement a common interface, defined in the public `shedoc` package:

```go
// Formatter transforms a parsed Document into a specific output format.
type Formatter interface {
    Format(w io.Writer, doc *Document) error
}
```

Built-in formatters are registered by name in a registry:

```go
var formatters = map[string]Formatter{}

// RegisterFormatter adds a formatter under the given name.
func RegisterFormatter(name string, f Formatter)

// GetFormatter returns the formatter for the given name, or nil.
func GetFormatter(name string) Formatter
```

Built-in formats (`json`, `help`, `man`, `completion:bash`, etc.) register themselves
via `init()` in their respective files under `internal/generate/`. The CLI resolves
the `-t` flag value through this registry.

This keeps the architecture open for future extensibility (e.g., external formatters
via subprocess dispatch) without adding any complexity now.

---

## Parser Design

Single-pass, line-by-line state machine in `parser.go`. No separate lexer — shedoc's line-oriented syntax doesn't need one.

### Public API

```go
// Parse parses shedoc documentation from a shell script file.
func Parse(path string) (*Document, error)

// ParseReader parses shedoc documentation from a reader.
func ParseReader(r io.Reader) (*Document, error)
```

### State Machine

```
stateTop → (shebang, shedoc single-line, shedoc block open, sheblock open, function decl)
stateShedoc → (continuation lines, block close → stateTop)
stateSheblock → (description lines, @tag lines, tag continuations, block close → stateTop)
```

### Line Classification

| Pattern | Classification |
|---------|---------------|
| `^#!/` | Shebang |
| `^#\?/(\w+)\s+(.+)$` | Shedoc single-line |
| `^#\?/(\w+)\s*$` | Shedoc block open |
| `^#@/(\w*)\s*(.*)$` | Sheblock open |
| `^ # (.*)$` | Continuation line |
| `^ ##$` | Block close |
| `^\s*(\w+)\s*\(\)` or `^\s*function\s+(\w+)` | Function declaration |

### Tag Parsing (`tag.go`)

Each tag type gets a dedicated parse function:

- `parseFlag(text) → Flag` — handles `-s`, `--long`, `-s | --long`, plus trailing description
- `parseOption(text) → Option` — like flag but extracts a `Value` between flag names and description
- `parseOperand(text) → Operand` — extracts `Value` and trailing description
- `parseEnv(text) → Env` — `VAR_NAME` + description
- `parseExit(text) → Exit` — code + description
- Simple tags (`@stdin`, `@stdout`, `@stderr`, `@deprecated`) — just capture the description

### Value Notation Parsing (`value.go`)

```go
func ParseValue(s string) (Value, error)
```

Handles: `<name>`, `[name]`, `[name=default]`, `<name...>`, `[name...]`

### Error Philosophy

**Lenient by default.** The parser collects `Warning`s for malformed tags rather than returning hard errors. A documentation tool shouldn't refuse to parse a file because of one typo. Hard errors are reserved for I/O failures (file not found, permission denied, etc.).

---

## CLI Design

Uses `cobra`. Binary is `shedoc`. The root command is the primary action (parse and
output); `complete` is a subcommand for live shell integration.

### Root Command — `shedoc <file...>`

Parses shedoc-annotated shell scripts and outputs the result in the specified format.

```
shedoc script.sh                          # JSON output (default)
shedoc script.sh -t help                  # help text
shedoc script.sh -t man                   # troff man page
shedoc script.sh -t completion:bash       # bash completion script
shedoc script.sh -t completion:zsh        # zsh completion script
shedoc script.sh -t completion:fish       # fish completion script
shedoc script.sh -g description           # extract a single tag value
shedoc script.sh -g version               # "2.1.0"
shedoc a.sh b.sh                          # multiple files (JSON array)
cat script.sh | shedoc -                  # read from stdin
```

Flags:
- `-t, --to <format>` — output format (default: `json`)
  - `json` — structured JSON
  - `help` — `--help` style text
  - `man` — troff/groff man page
  - `completion:bash` — bash completion script
  - `completion:zsh` — zsh completion script
  - `completion:fish` — fish completion script
- `-g, --get <tag>` — extract a single `#?/` tag value as plain text (e.g. `name`, `version`, `description`). Mutually exclusive with `--to`.
- `-o, --output <path>` — write to file instead of stdout
- `-w, --warnings` — include warnings in output (JSON) or emit to stderr (other formats)
- `-q, --quiet` — suppress warnings on stderr
- `--version` — print version info

Single file → single JSON object. Multiple files (paths, globs, directories) → NDJSON (one JSON object per line). Non-JSON formats (`-t help`, `-t man`, `-t completion:*`) accept a single file only.

#### Help Text Output

When `-t help` is used, generates `--help` style output:

```
deploy - A deployment tool for managing application releases

Usage:
  deploy [-v] [-c config] <command> [args...]

Commands:
  push        Deploys the application to the specified environment
  status      Shows the current deployment status for an environment
  rollback    Rolls back to the previous deployment
  migrate     [deprecated] Use 'deploy push --migrate' instead

Options:
  -v, --verbose          Enable verbose output
  -c, --config <path>    Path to configuration file

Environment:
  DEPLOY_TOKEN           Authentication token for the deployment service

Exit Codes:
  0    Success
  1    General error
  2    Authentication failure
```

#### Man Page Output

When `-t man` is used, maps `#?/` tags to man page sections:
- `#?/name` → `.TH` + `NAME`
- `#?/synopsis` → `SYNOPSIS`
- `#?/description` → `DESCRIPTION`
- `#@/command` flags/options → `OPTIONS`
- `#@/subcommand` blocks → `COMMANDS`
- `#?/examples` → `EXAMPLES`
- `@env` tags → `ENVIRONMENT`
- `@reads`/`@writes` tags → `FILES`
- `@exit` tags → `EXIT STATUS`
- `#?/author` → `AUTHOR`

### Subcommand — `shedoc complete <file>`

Dynamic completion hook for live shell integration. Instead of generating and
installing static completion scripts, shells invoke `shedoc complete` at tab-press
time to get completions that always reflect the current documentation.

```bash
# bash — one line in .bashrc
complete -C "shedoc complete deploy.sh" deploy

# zsh — via _comdef wrapper
# fish — via complete -c ... -a "(shedoc complete ...)"
```

The parser runs on each invocation, inspects the cursor position / current word
from shell environment variables (`COMP_LINE`, `COMP_POINT`, etc.), and returns
matching completions.

### Exit Codes

- `0` — success
- `1` — parse error or general failure
- `2` — invalid usage (bad flags, missing arguments)

---

## Testing Strategy

### Unit Tests — Parser

**Table-driven tests** with inline shell script snippets and expected `Document` structs:

```go
func TestParse(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  *Document
    }{
        {"shebang only", "#!/bin/bash\n", &Document{Shebang: "/bin/bash"}},
        // ...
    }
}
```

### Golden File Tests

Each `testdata/*.sh` has a corresponding `testdata/*.json` with expected parse output. A `-update` flag regenerates goldens.

### Tag & Value Parsing

Focused unit tests for each tag parser and value notation parser. These are small, pure functions with many edge cases.

### Generator Tests

Construct `Document` structs programmatically → run generator → compare output against golden strings/files. Especially important for man page troff output.

### CLI Integration Tests

Build the binary, run against `testdata/` scripts, assert stdout/stderr/exit code. Consider `rogpeppe/go-internal/testscript` for this.

### Test Dependencies

- `github.com/stretchr/testify` for assertions
- `rogpeppe/go-internal/testscript` for CLI integration (optional)

---

## Build & Release

### Makefile

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build test lint clean install golden

build:
	go build $(LDFLAGS) -o dist/shedoc ./cmd/shedoc

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf dist/

install:
	go install $(LDFLAGS) ./cmd/shedoc

golden:
	go test ./... -update
```

### goreleaser

Cross-compile for linux/darwin/windows (amd64/arm64). Homebrew tap at `nickawilliams/homebrew-tap`.

---

## v1 Scope

### Ships in v1

1. **Parser library** — full spec support as documented in README.md
2. **JSON output** (`shedoc script.sh`) — foundational output format (default)
3. **Help text format** (`shedoc script.sh -t help`) — immediate practical value
4. **Man page format** (`shedoc script.sh -t man`) — natural fit for the `#?/` tag design
5. **Completion script formats** (`shedoc script.sh -t completion:{bash,zsh,fish}`) — static generation
6. **Live completion hook** (`shedoc complete script.sh`) — dynamic shell integration

### Deferred

- Formatter (`shedoc fmt`), linter (`shedoc lint`), drift detection (`shedoc diff`) — need real-world usage feedback
- Editor integration — separate project, depends on stable parser
- v2 spec features (flag shorthand, `@requires`, `@see`, `@example`, etc.)

---

## Implementation Sequence

Each step produces a testable, reviewable unit:

1. `go mod init` + directory skeleton + `model.go` — review the data model first
2. `value.go` + tests — small, self-contained, fully testable
3. `tag.go` + tests — depends on value.go, fully testable with string inputs
4. `parser.go` + tests — state machine, uses tag.go, inline snippet tests + golden files
5. `testdata/*.sh` + `testdata/*.json` — golden fixtures validate the full pipeline
6. `formatter.go` — `Formatter` interface and registry
7. `cmd/shedoc/main.go` + `internal/cli/root.go` — first usable binary (`shedoc script.sh` outputs JSON, `-g` extraction)
8. `internal/generate/helptext.go` + `-t help` wiring
9. `internal/generate/manpage.go` + `-t man` wiring
10. `internal/generate/completions_*.go` + `-t completion:*` wiring
11. `internal/cli/complete.go` — live completion hook subcommand
12. `.goreleaser.yml` + Makefile — build and release infrastructure
13. End-to-end CLI integration tests

---

## Edge Cases the Parser Must Handle

- **Bare `#@/`** (no visibility keyword) → defaults to `public`
- **`#@/subcommand push`** → visibility=subcommand, name="push"
- **Standalone `#@/command`** (no function follows) → `FunctionName` stays empty
- **Tag continuation whitespace** — leading alignment whitespace on continuation lines is trimmed
- **No shedoc at all** → valid Document with just Shebang (if present)
- **No shebang** → valid (sourced libraries may omit it)
- **`#!/usr/bin/env bash`** → store full string after `#!` in Shebang
- **Windows line endings** → `bufio.Scanner` with `ScanLines` handles `\r\n`
