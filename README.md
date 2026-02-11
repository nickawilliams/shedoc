# Shedoc

Shedoc is a documentation standard for shell scripts, extending the familiar shebang pattern
with two additional sigils for structured documentation. It is descriptive, not prescriptive —
designed to document shell scripts however they happen to be written, without enforcing
opinions on structure or style.

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

File-level metadata for man pages and help output. All tags are optional, though tooling
(e.g., man page generation) may require specific tags.

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

Any shedoc tag can use the block form for multi-line content.

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

A file typically has one `#@/command` block.

### Subcommand Behavior

`#@/subcommand <name>` documents a subcommand. The `<name>` is what users type; the
function name can be anything. Common flags shared by all subcommands can be
documented in the `#@/command` block. When `#@/subcommand` blocks are present, the
available subcommands can be inferred — an explicit `@operand <command>` in the
`#@/command` block is optional.

## Block Tags (`@`)

Used within sheblocks to document inputs and outputs.

### Tag Continuation

A tag's description can span multiple lines. Any non-`@`, non-blank line following a
`@tag` continues that tag's description. A blank comment line (` #`) or the next `@tag`
terminates the continuation. Leading whitespace on continuation lines is trimmed.

```bash
 # @option -f | --format <type>   Output format. Supports json, yaml,
 #                                and xml with optional pretty-printing.
 #
 # @flag -v | --verbose           Enable verbose output
```

### Value Notation

| Syntax           | Meaning                 |
| ---------------- | ----------------------- |
| `<name>`         | Required                |
| `[name]`         | Optional                |
| `[name=default]` | Optional with default   |
| `<name...>`      | One or more (required)  |
| `[name...]`      | Zero or more (optional) |

### Input Tags

| Tag        | Syntax                                         | Description                         |
| ---------- | ---------------------------------------------- | ----------------------------------- |
| `@flag`    | `@flag -s \| --long` _description_             | Boolean flag (short, long, or both) |
| `@option`  | `@option -f \| --format <value>` _description_ | Option with required value          |
| `@option`  | `@option --format [value=json]` _description_  | Option with optional/default value  |
| `@operand` | `@operand <name>` _description_                | Required positional argument        |
| `@operand` | `@operand [name]` _description_                | Optional positional argument        |
| `@operand` | `@operand [name=default]` _description_        | Optional with default               |
| `@env`     | `@env VAR_NAME` _description_                  | Environment variable read           |
| `@reads`   | `@reads <path>` _description_                  | Implicit file read                  |
| `@stdin`   | `@stdin` _description_                         | Reads from standard input           |

The order of `@operand` tags reflects their positional order.

### Output Tags

| Tag       | Syntax                         | Description               |
| --------- | ------------------------------ | ------------------------- |
| `@exit`   | `@exit <code>` _description_   | Exit status code          |
| `@stdout` | `@stdout` _description_        | Writes to standard output |
| `@stderr` | `@stderr` _description_        | Writes to standard error  |
| `@sets`   | `@sets VAR_NAME` _description_ | Environment variable set  |
| `@writes` | `@writes <path>` _description_ | Implicit file write       |

### Metadata Tags

| Tag           | Syntax                  | Description         |
| ------------- | ----------------------- | ------------------- |
| `@deprecated` | `@deprecated [message]` | Marks as deprecated |

## Examples

### Comprehensive Example

A CLI tool with subcommands, demonstrating most Shedoc features:

```bash
#!/usr/bin/env bash

#?/name     deploy
#?/version  2.1.0
#?/synopsis deploy [-v] [-c config] <command> [args...]
#?/section  1
#?/author   Jane Developer
#?/license  MIT
#?/description
 # A deployment tool for managing application releases. Supports
 # multiple environments and rollback capabilities.
 ##
#?/examples
 # deploy status production
 # deploy push --force staging
 # echo "v1.2.3" | deploy push production
 ##

#@/command
 # Manages application deployments across environments.
 #
 # @flag    -v | --verbose          Enable verbose output
 # @option  -c | --config <path>    Path to configuration file
 # @operand <command>               Subcommand to run
 #
 # @env     DEPLOY_TOKEN            Authentication token for the deployment
 #                                  service. Can also be provided via the
 #                                  .deployrc configuration file.
 # @reads   ~/.deployrc             User configuration
 #
 # @exit    0                       Success
 # @exit    1                       General error
 # @exit    2                       Authentication failure
 # @stderr                          Error and diagnostic messages
 ##
main() {
    # top-level flag parsing ...

    case "$1" in
        push)     shift; cmd_push "$@" ;;
        status)   shift; cmd_status "$@" ;;
        rollback) shift; cmd_rollback "$@" ;;
        migrate)  shift; cmd_migrate "$@" ;;
        *)        echo "Unknown command: $1" >&2; exit 1 ;;
    esac
}

#@/subcommand push
 # Deploys the application to the specified environment.
 #
 # @flag    -f | --force             Skip confirmation prompt
 # @flag    --dry-run                Preview changes without deploying
 # @option  --tag [version]          Version tag (default: latest git tag)
 # @operand <environment>            Target environment (production, staging)
 # @operand [services...]            Specific services to deploy
 #
 # @stdin                            Reads version from STDIN if provided
 #
 # @exit    0                        Success
 # @exit    1                        Deploy failed
 # @stdout                           Deployment progress
 # @writes  /var/log/deploy.log      Deployment log
 ##
cmd_push() {
    # implementation
}

#@/subcommand status
 # Shows the current deployment status for an environment.
 #
 # @option  --format [fmt=text]      Output format (text, json, yaml)
 # @operand <environment>            Target environment
 #
 # @exit    0                        Success
 # @stdout                           Status information
 ##
cmd_status() {
    # implementation
}

#@/subcommand rollback
 # Rolls back to the previous deployment.
 #
 # @flag    -f | --force             Skip confirmation prompt
 # @operand <environment>            Target environment
 # @operand [version]                Specific version to roll back to
 #
 # @sets    DEPLOY_LAST_ROLLBACK     Timestamp of last rollback
 # @writes  /var/log/deploy.log      Rollback log entry
 #
 # @exit    0                        Success
 # @exit    1                        Rollback failed
 # @stdout                           Rollback progress
 ##
cmd_rollback() {
    # implementation
}

#@/subcommand migrate
 # @deprecated Use 'deploy push --migrate' instead.
 ##
cmd_migrate() {
    # implementation
}

main "$@"
```

### Standalone Command

When `#@/command` has no function following it, it documents the script's inline logic:

```bash
#!/usr/bin/env bash

#?/name    greet
#?/version 1.0.0

#@/command
 # Prints a greeting message.
 #
 # @operand [name=World]              Name to greet
 # @exit    0                         Success
 # @stdout                            Greeting message
 ##

echo "Hello, ${1:-World}!"
```

### Sourced Library

When the script is meant to be sourced, use `#@/public` and `#@/private`:

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

#@/private
 # Internal helper for validation.
 ##
_validate_input() {
    [[ -n "$1" ]]
}
```

## Notes

- Flags and options support short-only (`-v`), long-only (`--verbose`), or both (`-v | --verbose`).

- A single conceptual input may be provided via multiple forms (e.g., `-v`, `--verbose`, `VERBOSE=1`). The `@flag` syntax supports pipe-separated forms to express this.

- See [ROADMAP.md](ROADMAP.md) for planned features.
