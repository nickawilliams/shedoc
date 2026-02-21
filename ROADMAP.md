# Shedoc Roadmap

## Specification

Features under consideration for future versions of the [spec](SPEC.md).

### v2 Candidates

#### Flag Shorthand Syntax

Compact bracket notation for flags where the short form character appears in the long name:

```bash
# @flag [v]erbose              → -v, --verbose
# @flag midd[l]e              → -l, --middle
# @flag las[t]                → -t, --last
# @flag long-only             → --long-only
# @flag [s]                   → -s
```

Rules:

- `[x]` within a name marks the short form character
- No brackets = long form only
- Brackets with single char only = short form only

This would complement the universal explicit syntax (`@flag -v | --verbose`) as
syntactic sugar for the common case where the short flag letter appears in the
long flag name.

#### Dependencies

Document external command requirements:

```bash
#?/requires jq curl
```

Or at function level:

```bash
#@/public
 # @requires jq
 ##
```

**Considerations:** Platform differences (GNU vs BSD), version requirements, POSIX variations make this complex. May be informational only.

#### Cross-references

Link to other functions or man pages:

```bash
#@/public
 # @see other_function
 # @see grep(1)
 ##
```

#### Inline Examples

Function-level `@example` tag:

```bash
#@/public
 # @example to_upper "hello" → "HELLO"
 ##
```

#### Mutually Exclusive Options

Indicate options that can't be used together:

```bash
# Option A: Groups
# @flag [v]erbose group=output
# @flag [q]uiet group=output

# Option B: Constraint
# @exclusive verbose quiet
```

#### Multi-value / Repeatable Options

Distinguish between:

- `--include a b c` (one flag, multiple values)
- `--include a --include b` (repeated flag)

Possible syntax:

```bash
# @option [i]nclude <pattern...>    Multi-value
# @option [i]nclude <pattern>...    Repeatable
```

#### Variable Documentation

Support for documenting global variables and constants:

```bash
#@/public
 # Default timeout in seconds.
 # @readonly
 # @default 30
 ##
declare -gr TIMEOUT=30
```

#### Nested Subcommands

Support for `git remote add` style nested commands:

```bash
#@/command remote-add
```

Or hierarchical syntax TBD.

#### Aliases

Document command/option aliases:

```bash
#@/command checkout
 # @alias co
 ##
```

#### Interactive Prompts

Document interactive user prompts:

```bash
#@/command
 # @prompt "Continue deploy?"    Confirmation unless --force is set
 ##
```

Useful for automation contexts where knowing a script will block for input is important.

#### Shorthand Patterns

Reduce boilerplate for common input patterns:

```bash
# Pattern definition
#?/pattern standard  -{short} | --{name} | {PREFIX}_{NAME}=

# Usage
# @arg:standard verbose  Enable verbose output
```

Expands to `-v | --verbose | SCRIPT_VERBOSE=`.

## Tooling

Tools under consideration to support the Shedoc ecosystem.

### Parser

The foundation for all other tools. Reads shell scripts and extracts Shedoc
comments into structured data (JSON/AST).

### Shell Completions Generator

Generate bash, zsh, and fish completion scripts from Shedoc. Flags, options,
subcommands, and operands provide everything needed for completions.

### Help Text Generator

Auto-generate `--help` output from Shedoc comments. Could be a function sourced
into the script, keeping help text in sync with documentation.

### Man Page Generator

Generate troff/groff man pages from Shedoc. The `#?/` tags map almost directly
to man page sections.

### Formatter

Auto-format Shedoc comments: column-align tag values and descriptions, normalize
spacing between groups, and apply consistent continuation line style.

### Linter

Validate Shedoc syntax is well-formed. Could also flag undocumented public
functions, missing exit codes, or other documentation gaps.

### Drift Detection

Compare Shedoc documentation against the actual implementation. Detect
discrepancies like documented flags that don't exist, undocumented subcommands,
or mismatched exit codes.

### Editor Integration

Syntax highlighting for Shedoc comments, tag autocomplete, and snippets for
VS Code, vim, and other editors.
