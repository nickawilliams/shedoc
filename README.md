# Shedoc

Shedoc is a documentation standard and supporting tooling for shell scripts.

## Documenting Scripts

```bash
#?/synopsis
#?/summary
#?/index

#?/description
```

### `#?/synopsis`

### `#?/summary`

### `#?/index`

### `#?/description`

## Documenting Functions

Functions can take advantage of shedoc docblocks. Following a similar pattern
used in other languages, it provides a formal syntax for documenting a
function's ins, outs, and behavior.

```bash
##
 # [description]
 #
 # @env
 #
 # @arg
 # @switch
 #
 # @stdin
 # @stdout
 # @stderr
 #
 # @return
```

```bash
##
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
