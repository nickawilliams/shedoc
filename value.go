package shedoc

import (
	"fmt"
	"strings"
)

// ParseValue parses value notation like <name>, [name], [name=default],
// <name...>, or [name...] into a Value struct.
func ParseValue(s string) (Value, error) {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return Value{}, fmt.Errorf("invalid value notation: %q", s)
	}

	open := s[0]
	close := s[len(s)-1]

	var required bool
	switch {
	case open == '<' && close == '>':
		required = true
	case open == '[' && close == ']':
		required = false
	default:
		return Value{}, fmt.Errorf("invalid value notation: %q (must be <...> or [...])", s)
	}

	inner := s[1 : len(s)-1]
	if inner == "" {
		return Value{}, fmt.Errorf("invalid value notation: %q (empty name)", s)
	}

	var variadic bool
	if strings.HasSuffix(inner, "...") {
		variadic = true
		inner = strings.TrimSuffix(inner, "...")
		if inner == "" {
			return Value{}, fmt.Errorf("invalid value notation: %q (empty name before ...)", s)
		}
	}

	var def string
	if idx := strings.Index(inner, "="); idx >= 0 {
		if required {
			return Value{}, fmt.Errorf("invalid value notation: %q (defaults not allowed in required values)", s)
		}
		def = inner[idx+1:]
		inner = inner[:idx]
		if inner == "" {
			return Value{}, fmt.Errorf("invalid value notation: %q (empty name before =)", s)
		}
	}

	return Value{
		Name:     inner,
		Required: required,
		Default:  def,
		Variadic: variadic,
	}, nil
}
