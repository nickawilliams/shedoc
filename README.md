# Shedoc

Shedoc is a documentation standard and supporting tooling for shell scripts.

## Documenting Scripts (shedoc)

The _shedoc_ is used to document the script as a whole. Typically one or more
are placed at the top of the file below the shebang.

`#?/<tag> <description>`

Example:

```bash
#?/name
#?/synopsis
#?/summary
#?/index

#?/description
```

### `#?/synopsis`

### `#?/summary`

### `#?/index`

### `#?/description`

## Documenting Functions (sheblock)

The _shedoc block_ (colloquially the _sheblock_) is used for doucementation
scoped to a specific entity in the script. Primarily intended for functions, it
can also be used for variables and other entities.

Functions can take advantage of shedoc _sheblocks_. Following a similar
docblock pattern used in other languages, it provides a formal syntax for
documenting a function's ins, outs, and behavior.

```
#@/<main | public | private>
 # <description>
 #
 # @<tag> <description>
 # ...
 #
<function | variable | local | <etc>>
```

Example:

```bash
#@/public
 # Links all specified packages that are found in the target package's
 # dependency list.
 #
 # @env      ENV_VARIABLE an environment variable used by the function
 #
 # @arg      the target package containing dependencies to be linked
 # @arg      array   the array of dependencies to be linked, if present
 # @switch   [-a, --all]
 #
 # @stdin
 # @stdout
 # @stderr
 #
 # @exit     0   on success
 # @exit     1   if the target package cannot be found
 # @exit     2   if one or more of the matching dependencies could not be linked
 #
```

### `@env`

### `@arg`

### `@param`

### `@switch`

### `@stdin`

### `@stdout`

### `@stderr`

### `@exit`

## Tools

### Doc
