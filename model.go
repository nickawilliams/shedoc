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
	Name         string      `json:"name,omitempty"`
	Description  string      `json:"description,omitempty"`
	FunctionName string      `json:"functionName,omitempty"`
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

// Env represents an environment variable read: @env VAR_NAME description
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
