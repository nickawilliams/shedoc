package generate

import (
	"fmt"
	"io"
	"strings"

	"github.com/nickawilliams/shedoc"
)

func init() {
	shedoc.RegisterFormatter("help", &HelpTextFormatter{})
}

// HelpTextFormatter outputs a Document as --help style text.
type HelpTextFormatter struct{}

func (f *HelpTextFormatter) Format(w io.Writer, doc *shedoc.Document) error {
	// Header: name - description
	if doc.Meta.Name != "" {
		if doc.Meta.Description != "" {
			// Use first line of description as the brief.
			brief := firstLine(doc.Meta.Description)
			fmt.Fprintf(w, "%s - %s\n", doc.Meta.Name, brief)
		} else {
			fmt.Fprintln(w, doc.Meta.Name)
		}
		fmt.Fprintln(w)
	}

	// Usage
	if doc.Meta.Synopsis != "" {
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintf(w, "  %s\n", doc.Meta.Synopsis)
		fmt.Fprintln(w)
	}

	// Find the command block and subcommand blocks.
	var cmdBlock *shedoc.Block
	var subcommands []shedoc.Block
	for i := range doc.Blocks {
		switch doc.Blocks[i].Visibility {
		case shedoc.VisibilityCommand:
			cmdBlock = &doc.Blocks[i]
		case shedoc.VisibilitySubcommand:
			subcommands = append(subcommands, doc.Blocks[i])
		}
	}

	// Commands section
	if len(subcommands) > 0 {
		fmt.Fprintln(w, "Commands:")
		nameWidth := maxSubcommandNameWidth(subcommands)
		for _, sub := range subcommands {
			desc := firstLine(sub.Description)
			if sub.Deprecated != nil {
				if desc != "" {
					desc = "[deprecated] " + desc
				} else {
					desc = "[deprecated] " + sub.Deprecated.Message
				}
			}
			if desc != "" {
				fmt.Fprintf(w, "  %-*s  %s\n", nameWidth, sub.Name, desc)
			} else {
				fmt.Fprintf(w, "  %s\n", sub.Name)
			}
		}
		fmt.Fprintln(w)
	}

	// Options section (flags and options from the command block)
	if cmdBlock != nil && (len(cmdBlock.Flags) > 0 || len(cmdBlock.Options) > 0) {
		fmt.Fprintln(w, "Options:")
		printFlags(w, cmdBlock.Flags)
		printOptions(w, cmdBlock.Options)
		fmt.Fprintln(w)
	}

	// Environment section
	if cmdBlock != nil && len(cmdBlock.Env) > 0 {
		fmt.Fprintln(w, "Environment:")
		nameWidth := maxEnvNameWidth(cmdBlock.Env)
		for _, env := range cmdBlock.Env {
			desc := firstLine(env.Description)
			if desc != "" {
				fmt.Fprintf(w, "  %-*s  %s\n", nameWidth, env.Name, desc)
			} else {
				fmt.Fprintf(w, "  %s\n", env.Name)
			}
		}
		fmt.Fprintln(w)
	}

	// Exit Codes section
	if cmdBlock != nil && len(cmdBlock.Exit) > 0 {
		fmt.Fprintln(w, "Exit Codes:")
		codeWidth := maxExitCodeWidth(cmdBlock.Exit)
		for _, exit := range cmdBlock.Exit {
			if exit.Description != "" {
				fmt.Fprintf(w, "  %-*s  %s\n", codeWidth, exit.Code, exit.Description)
			} else {
				fmt.Fprintf(w, "  %s\n", exit.Code)
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}

func printFlags(w io.Writer, flags []shedoc.Flag) {
	for _, f := range flags {
		label := formatFlagLabel(f.Short, f.Long)
		if f.Description != "" {
			fmt.Fprintf(w, "  %-24s%s\n", label, f.Description)
		} else {
			fmt.Fprintf(w, "  %s\n", label)
		}
	}
}

func printOptions(w io.Writer, options []shedoc.Option) {
	for _, o := range options {
		label := formatOptionLabel(o.Short, o.Long, o.Value)
		if o.Description != "" {
			fmt.Fprintf(w, "  %-24s%s\n", label, o.Description)
		} else {
			fmt.Fprintf(w, "  %s\n", label)
		}
	}
}

func formatFlagLabel(short, long string) string {
	switch {
	case short != "" && long != "":
		return short + ", " + long
	case short != "":
		return short
	default:
		return "    " + long
	}
}

func formatOptionLabel(short, long string, val shedoc.Value) string {
	valStr := formatValue(val)
	switch {
	case short != "" && long != "":
		return short + ", " + long + " " + valStr
	case short != "":
		return short + " " + valStr
	default:
		return "    " + long + " " + valStr
	}
}

func formatValue(v shedoc.Value) string {
	name := v.Name
	if v.Variadic {
		name += "..."
	}
	if v.Required {
		return "<" + name + ">"
	}
	if v.Default != "" {
		return "[" + name + "=" + v.Default + "]"
	}
	return "[" + name + "]"
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func maxSubcommandNameWidth(subs []shedoc.Block) int {
	max := 0
	for _, s := range subs {
		if len(s.Name) > max {
			max = len(s.Name)
		}
	}
	return max
}

func maxEnvNameWidth(envs []shedoc.Env) int {
	max := 0
	for _, e := range envs {
		if len(e.Name) > max {
			max = len(e.Name)
		}
	}
	return max
}

func maxExitCodeWidth(exits []shedoc.Exit) int {
	max := 0
	for _, e := range exits {
		if len(e.Code) > max {
			max = len(e.Code)
		}
	}
	return max
}
