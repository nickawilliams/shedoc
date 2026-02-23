# shedoc

A documentation standard and CLI tool for shell scripts. Shedoc extends the
familiar shebang (`#!/`) with two additional sigils — `#?/` for file metadata
and `#@/` for code documentation — to embed structured, machine-readable
documentation directly in shell scripts.

```bash
#!/usr/bin/env bash

#?/name    greet
#?/version 1.0.0

#@/command
 # Prints a greeting message.
 #
 # @operand [name=World]   Name to greet
 #
 # @stdout                  Greeting message
 #
 # @exit    0               Success
 ##

echo "Hello, ${1:-World}!"
```

The `shedoc` CLI parses these annotations and outputs structured data in a
variety of formats: JSON, help text, man pages, and shell completion scripts.

## Installation

```bash
go install github.com/nickawilliams/shedoc/cmd/shedoc@latest
```

## Usage

```bash
shedoc script.sh                        # JSON (default)
shedoc script.sh -t help                # --help style text
shedoc script.sh -t man                 # troff man page
shedoc script.sh -t completion:bash     # bash completion script
shedoc script.sh -t completion:zsh      # zsh completion script
shedoc script.sh -t completion:fish     # fish completion script
shedoc script.sh -g version             # extract a single metadata value
cat script.sh | shedoc -                # read from stdin
shedoc a.sh b.sh                        # multiple files → NDJSON
```

### Flags

| Flag | Description |
| --- | --- |
| `-t, --to <format>` | Output format (`json`, `help`, `man`, `completion:bash`, `completion:zsh`, `completion:fish`) |
| `-g, --get <path>` | Extract a single `#?/` path value as plain text |
| `-o, --output <path>` | Write output to file instead of stdout |
| `-w, --warnings` | Include warnings in JSON output |
| `-q, --quiet` | Suppress warnings on stderr |
| `--version` | Print version |

### Library Usage

The parser is also available as a Go library:

```go
import "github.com/nickawilliams/shedoc"

doc, err := shedoc.Parse("script.sh")
fmt.Println(doc.Meta.Name)    // "greet"
fmt.Println(doc.Meta.Version) // "1.0.0"
```

## Specification

The full shedoc documentation standard is defined in [SPEC.md](SPEC.md).
See [ROADMAP.md](ROADMAP.md) for planned specification and tooling features.

## License

[BSD-3-Clause](LICENSE.md)
