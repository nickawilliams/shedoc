package shedoc

import (
	"fmt"
	"strings"
)

// parseTag dispatches to the appropriate tag parser based on the tag name.
// text is everything after "@tagname " on the line.
func parseTag(name, text string, line int) (tagName string, result any, err error) {
	switch name {
	case "flag":
		r, e := parseFlag(text, line)
		return name, r, e
	case "option":
		r, e := parseOption(text, line)
		return name, r, e
	case "operand":
		r, e := parseOperand(text, line)
		return name, r, e
	case "env":
		r, e := parseEnv(text, line)
		return name, r, e
	case "reads":
		r, e := parseReads(text, line)
		return name, r, e
	case "stdin":
		return name, &Stdin{Description: text, Line: line}, nil
	case "exit":
		r, e := parseExit(text, line)
		return name, r, e
	case "stdout":
		return name, &Stdout{Description: text, Line: line}, nil
	case "stderr":
		return name, &Stderr{Description: text, Line: line}, nil
	case "sets":
		r, e := parseSets(text, line)
		return name, r, e
	case "writes":
		r, e := parseWrites(text, line)
		return name, r, e
	case "deprecated":
		return name, &Deprecated{Message: text, Line: line}, nil
	default:
		return name, nil, fmt.Errorf("unknown tag @%s", name)
	}
}

// parseFlag parses: -s | --long description
// Supports short-only, long-only, or both with pipe separator.
func parseFlag(text string, line int) (*Flag, error) {
	f := &Flag{Line: line}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@flag requires at least one flag name")
	}

	rest := consumeFlags(text, &f.Short, &f.Long)
	f.Description = strings.TrimSpace(rest)
	return f, nil
}

// parseOption parses: -f | --format <value> description
func parseOption(text string, line int) (*Option, error) {
	o := &Option{Line: line}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@option requires at least one flag name and a value")
	}

	rest := consumeFlags(text, &o.Short, &o.Long)
	rest = strings.TrimSpace(rest)

	// Next token should be a value notation
	valStr, desc := splitFirstToken(rest)
	if valStr == "" {
		return nil, fmt.Errorf("@option requires a value notation (e.g., <value> or [value])")
	}

	v, err := ParseValue(valStr)
	if err != nil {
		return nil, fmt.Errorf("@option value: %w", err)
	}
	o.Value = v
	o.Description = strings.TrimSpace(desc)
	return o, nil
}

// parseOperand parses: <name> description or [name] description
func parseOperand(text string, line int) (*Operand, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@operand requires a value notation")
	}

	valStr, desc := splitFirstToken(text)
	v, err := ParseValue(valStr)
	if err != nil {
		return nil, fmt.Errorf("@operand value: %w", err)
	}

	return &Operand{
		Value:       v,
		Description: strings.TrimSpace(desc),
		Line:        line,
	}, nil
}

// parseEnv parses: VAR_NAME description
func parseEnv(text string, line int) (*Env, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@env requires a variable name")
	}

	name, desc := splitFirstToken(text)
	return &Env{
		Name:        name,
		Description: strings.TrimSpace(desc),
		Line:        line,
	}, nil
}

// parseReads parses: <path> description or a bare path
func parseReads(text string, line int) (*Reads, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@reads requires a path")
	}

	path, desc := splitFirstToken(text)
	return &Reads{
		Path:        path,
		Description: strings.TrimSpace(desc),
		Line:        line,
	}, nil
}

// parseExit parses: <code> description
func parseExit(text string, line int) (*Exit, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@exit requires an exit code")
	}

	code, desc := splitFirstToken(text)
	return &Exit{
		Code:        code,
		Description: strings.TrimSpace(desc),
		Line:        line,
	}, nil
}

// parseSets parses: VAR_NAME description
func parseSets(text string, line int) (*Sets, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@sets requires a variable name")
	}

	name, desc := splitFirstToken(text)
	return &Sets{
		Name:        name,
		Description: strings.TrimSpace(desc),
		Line:        line,
	}, nil
}

// parseWrites parses: <path> description
func parseWrites(text string, line int) (*Writes, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("@writes requires a path")
	}

	path, desc := splitFirstToken(text)
	return &Writes{
		Path:        path,
		Description: strings.TrimSpace(desc),
		Line:        line,
	}, nil
}

// consumeFlags parses flag names from the beginning of text, setting short
// and/or long as found. Returns the remaining text after flags.
// Handles: -s, --long, -s | --long
func consumeFlags(text string, short, long *string) string {
	text = strings.TrimSpace(text)

	for text != "" {
		if strings.HasPrefix(text, "--") {
			name, rest := splitFirstToken(text)
			*long = name
			text = strings.TrimSpace(rest)
		} else if strings.HasPrefix(text, "-") {
			name, rest := splitFirstToken(text)
			*short = name
			text = strings.TrimSpace(rest)
		} else {
			break
		}

		// Consume pipe separator if present
		if strings.HasPrefix(text, "|") {
			text = strings.TrimSpace(text[1:])
		} else {
			break
		}
	}

	return text
}

// splitFirstToken splits text into the first whitespace-delimited token and
// the rest.
func splitFirstToken(s string) (token, rest string) {
	s = strings.TrimSpace(s)
	idx := strings.IndexAny(s, " \t")
	if idx < 0 {
		return s, ""
	}
	return s[:idx], s[idx+1:]
}
