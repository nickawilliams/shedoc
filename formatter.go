package shedoc

import "io"

// Formatter transforms a parsed Document into a specific output format.
type Formatter interface {
	Format(w io.Writer, doc *Document) error
}

var formatters = map[string]Formatter{}

// RegisterFormatter adds a formatter under the given name.
func RegisterFormatter(name string, f Formatter) {
	formatters[name] = f
}

// GetFormatter returns the formatter registered under the given name, or nil.
func GetFormatter(name string) Formatter {
	return formatters[name]
}

// RegisteredFormats returns the names of all registered formatters.
func RegisteredFormats() []string {
	names := make([]string, 0, len(formatters))
	for name := range formatters {
		names = append(names, name)
	}
	return names
}
