# Shedoc

Shedoc is a documentation standard for shell scripts, extending the familiar shebang pattern
with two additional sigils for structured documentation.

## Sigils

| Sigil | Name     | Purpose                                      |
| ----- | -------- | -------------------------------------------- |
| `#!/` | Shebang  | Interpreter path (standard)                  |
| `#?/` | Shedoc   | File metadata (name, version, synopsis)      |
| `#@/` | Sheblock | Code documentation (functions, entry points) |

The `/` begins a reference — filesystem path for shebangs, documentation tag for shedocs.

## Block Syntax

Multi-line documentation uses indented continuation lines:

```bash
#@/visibility
 # Description line
 # @tag value
 ##
```

- **Open:** `#?/tag` or `#@/visibility`
- **Continue:** `␣#␣` (space, hash, space)
- **Close:** ` ##` (space, double hash)

Single-line shedocs need no closing:

```bash
#?/name my-script
#?/version 1.0.0
```

## Shedoc Tags (`#?/`)

File-level metadata for man pages and help output.

| Tag              | Description                       |
| ---------------- | --------------------------------- |
| `#?/name`        | Script name and brief description |
| `#?/version`     | Version string                    |
| `#?/synopsis`    | Usage pattern                     |
| `#?/description` | Full description (multi-line)     |
| `#?/examples`    | Usage examples (multi-line)       |
| `#?/section`     | Man page section (default: 1)     |
| `#?/author`      | Author name                       |
| `#?/license`     | License identifier                |

## Sheblock Visibility (`#@/`)

| Visibility             | Meaning                                        |
| ---------------------- | ---------------------------------------------- |
| `#@/command`           | CLI command (function or script)               |
| `#@/subcommand <name>` | Subcommand (function invoked via command name) |
| `#@/public`            | Public function, available when sourced        |
| `#@/private`           | Internal function, not part of public API      |
| `#@/`                  | Bare (no visibility) defaults to public        |

### Command Behavior

`#@/command` can document:

1. **A function** — when immediately followed by a function declaration
2. **The script itself** — when standalone (no function follows)

### Subcommand Behavior

`#@/subcommand <name>` documents a subcommand. The `<name>` is what users type; the
function name can be anything. Common flags shared by all subcommands should be
documented in the `#@/command` block.

## Input/Output Types

|  I/O   | Name                 | Example                   | Description         |
| :----: | -------------------- | ------------------------- | ------------------- |
| input  | flag (short)         | `cmd -f`                  | boolean argument    |
| input  | flag (long)          | `cmd --flag-arg`          | boolean argument    |
| input  | option               | `cmd --named-arg "value"` | named argument      |
| input  | operand              | `cmd value`               | positional argument |
| input  | prompt               | `Enter password:`         |                     |
| input  | STDIN                | `cmd < data.txt`          |                     |
| input  | environment variable | `ENV_VAR=value cmd`       |                     |
| input  | file                 | `~/.cmdrc`                |                     |
| output | exit code            | `exit 1`                  |                     |
| output | STDOUT               | `echo "output"`           |                     |
| output | STDERR               | `echo "error" >&2`        |                     |
| output | environment variable | `export FOO=bar`          |                     |
| output | file                 | `/var/log/cmd.log`        |                     |

## Block Tags (`@`)

Used within sheblocks to document inputs and outputs.

### Value Notation

| Syntax           | Meaning                 |
| ---------------- | ----------------------- |
| `<name>`         | Required                |
| `[name]`         | Optional                |
| `[name=default]` | Optional with default   |
| `<name...>`      | One or more (required)  |
| `[name...]`      | Zero or more (optional) |

### Input Tags

| Tag        | Syntax                           | Description                         |
| ---------- | -------------------------------- | ----------------------------------- |
| `@flag`    | `@flag -s \| --long`             | Boolean flag (short, long, or both) |
| `@option`  | `@option -f \| --format <value>` | Option with required value          |
| `@option`  | `@option --format [value=json]`  | Option with optional/default value  |
| `@operand` | `@operand <name>`                | Required positional argument        |
| `@operand` | `@operand [name]`                | Optional positional argument        |
| `@operand` | `@operand [name=default]`        | Optional with default               |
| `@env`     | `@env VAR_NAME`                  | Environment variable read           |
| `@reads`   | `@reads <path>`                  | Implicit file read                  |
| `@stdin`   | `@stdin`                         | Reads from standard input           |
| `@prompt`  | `@prompt "message"`              | Interactive user prompt             |

### Output Tags

| Tag       | Syntax                         | Description                   |
| --------- | ------------------------------ | ----------------------------- |
| `@exit`   | `@exit <code> <description>`   | Exit status code              |
| `@return` | `@return <code> <description>` | Return status (for functions) |
| `@stdout` | `@stdout`                      | Writes to standard output     |
| `@stderr` | `@stderr`                      | Writes to standard error      |
| `@sets`   | `@sets VAR_NAME`               | Environment variable set      |
| `@writes` | `@writes <path>`               | Implicit file write           |

### Metadata Tags

| Tag           | Syntax                  | Description         |
| ------------- | ----------------------- | ------------------- |
| `@deprecated` | `@deprecated [message]` | Marks as deprecated |

## Examples

### Example 1: Function as Entry Point

When a single function is the CLI interface:

```bash
#!/usr/bin/env bash

#?/name     process-data
#?/version  1.0.0
#?/synopsis process-data [-v] [-f format] <file>

#@/command
 # Processes data from a file or STDIN and outputs the result.
 #
 # @flag    -v | --verbose            Enable verbose output
 # @option  -f | --format <type>      Output format (json, yaml, xml)
 # @operand <file>                    Input file to process
 #
 # @env     PROCESS_DATA_API_KEY      API key for processing service
 # @reads   ~/.process_datarc         User configuration
 # @stdin                             Reads input if no file provided
 #
 # @exit    0                         Success
 # @exit    1                         API key not set
 # @exit    2                         User cancelled
 #
 # @stdout                            Processed output
 # @stderr                            Error messages
 ##
process_data() {
    # implementation
}

process_data "$@"
```

### Example 2: Standalone Entry Point

When the script has inline logic (no wrapper function):

```bash
#!/usr/bin/env bash

#?/name    greet
#?/version 1.0.0

#@/command
 # Prints a greeting message.
 #
 # @flag    -v | --verbose            Include extra details
 # @operand <name>                    Name to greet
 #
 # @exit    0                         Success
 # @stdout                            Greeting message
 ##

verbose=false
while getopts "v" opt; do
    case "$opt" in
        v) verbose=true ;;
    esac
done
shift $((OPTIND - 1))

echo "Hello, ${1:-World}!"
```

### Example 3: Library with Public/Private Functions

When the script is meant to be sourced:

```bash
#!/usr/bin/env bash

#?/name        string-utils
#?/version     1.0.0
#?/description
 # A library of string manipulation functions.
 ##

#@/public
 # Converts a string to uppercase.
 #
 # @operand <string>    The string to convert
 # @stdout              Uppercase result
 ##
to_upper() {
    echo "${1^^}"
}

#@/public
 # Converts a string to lowercase.
 #
 # @operand <string>    The string to convert
 # @stdout              Lowercase result
 ##
to_lower() {
    echo "${1,,}"
}

#@/private
 # Internal helper for validation.
 ##
_validate_input() {
    [[ -n "$1" ]]
}
```

### Example 4: Script with Subcommands

When a script has multiple subcommands:

```bash
#!/usr/bin/env bash

#?/name     pkg
#?/version  1.0.0
#?/synopsis pkg <command> [options]

#@/command
 # A simple package manager.
 #
 # @flag    -v | --verbose  Enable verbose output (applies to all commands)
 # @operand <command>       Subcommand to run
 ##

#@/subcommand install
 # Installs a package.
 #
 # @operand <package>       Package name to install
 # @flag    -f | --force    Overwrite existing installation
 # @exit    0               Success
 # @exit    1               Package not found
 ##
cmd_install() {
    # implementation
}

#@/subcommand remove
 # Removes an installed package.
 #
 # @operand <package>       Package name to remove
 # @flag    -f | --force    Remove without confirmation
 # @exit    0               Success
 # @exit    1               Package not installed
 ##
cmd_remove() {
    # implementation
}

#@/subcommand list
 # Lists installed packages.
 #
 # @flag    -a | --all      Include system packages
 # @stdout                  List of packages
 ##
cmd_list() {
    # implementation
}

case "$1" in
    install) shift; cmd_install "$@" ;;
    remove)  shift; cmd_remove "$@" ;;
    list)    shift; cmd_list "$@" ;;
    *)       echo "Unknown command: $1" >&2; exit 1 ;;
esac
```

## Notes

- A single conceptual input may be provided via multiple forms (e.g., `-v`, `--verbose`, `VERBOSE=1`). The `@flag` syntax supports pipe-separated forms to express this.

- See [ROADMAP.md](ROADMAP.md) for planned features.
