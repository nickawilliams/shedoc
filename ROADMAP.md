# Shedoc Roadmap

Features under consideration for future versions.

## v2 Candidates

### Flag Shorthand Syntax

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

### Dependencies

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

### Cross-references

Link to other functions or man pages:

```bash
#@/public
 # @see other_function
 # @see grep(1)
 ##
```

### Inline Examples

Function-level `@example` tag:

```bash
#@/public
 # @example to_upper "hello" → "HELLO"
 ##
```

### Mutually Exclusive Options

Indicate options that can't be used together:

```bash
# Option A: Groups
# @flag [v]erbose group=output
# @flag [q]uiet group=output

# Option B: Constraint
# @exclusive verbose quiet
```

### Multi-value / Repeatable Options

Distinguish between:

- `--include a b c` (one flag, multiple values)
- `--include a --include b` (repeated flag)

Possible syntax:

```bash
# @option [i]nclude <pattern...>    Multi-value
# @option [i]nclude <pattern>...    Repeatable
```

### Variable Documentation

Support for documenting global variables and constants:

```bash
#@/public
 # Default timeout in seconds.
 # @readonly
 # @default 30
 ##
declare -gr TIMEOUT=30
```

### Nested Subcommands

Support for `git remote add` style nested commands:

```bash
#@/command remote-add
```

Or hierarchical syntax TBD.

### Aliases

Document command/option aliases:

```bash
#@/command checkout
 # @alias co
 ##
```

### Shorthand Patterns

Reduce boilerplate for common input patterns:

```bash
# Pattern definition
#?/pattern standard  -{short} | --{name} | {PREFIX}_{NAME}=

# Usage
# @arg:standard verbose  Enable verbose output
```

Expands to `-v | --verbose | SCRIPT_VERBOSE=`.

## Ideas

- Multi-language documentation support
- Tooling for generating man pages, markdown, HTML
- IDE/editor integration (syntax highlighting, linting)
- Validation tooling (check docs match implementation)
