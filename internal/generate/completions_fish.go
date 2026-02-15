package generate

import (
	"fmt"
	"io"

	"github.com/nickawilliams/shedoc"
)

func init() {
	shedoc.RegisterFormatter("completion:fish", &FishCompletionFormatter{})
}

// FishCompletionFormatter generates a fish completion script.
type FishCompletionFormatter struct{}

func (f *FishCompletionFormatter) Format(w io.Writer, doc *shedoc.Document) error {
	name := doc.Meta.Name
	if name == "" {
		return fmt.Errorf("completion generation requires #?/name")
	}

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

	fmt.Fprintf(w, "# fish completion for %s\n\n", name)

	hasSubcommands := len(subcommands) > 0

	// Global flags/options
	if cmdBlock != nil {
		writeFishFlags(w, name, cmdBlock.Flags, hasSubcommands, "")
		writeFishOptions(w, name, cmdBlock.Options, hasSubcommands, "")
	}

	// Subcommands
	if hasSubcommands {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "# Subcommands\n")
		for _, sub := range subcommands {
			desc := firstLine(sub.Description)
			if sub.Deprecated != nil {
				desc = "[deprecated] " + sub.Deprecated.Message
			}
			fmt.Fprintf(w, "complete -c %s -n '__fish_use_subcommand' -a %s", name, sub.Name)
			if desc != "" {
				fmt.Fprintf(w, " -d '%s'", fishEscape(desc))
			}
			fmt.Fprintln(w)
		}

		// Per-subcommand flags
		for _, sub := range subcommands {
			if len(sub.Flags) == 0 && len(sub.Options) == 0 {
				continue
			}
			fmt.Fprintln(w)
			fmt.Fprintf(w, "# %s subcommand\n", sub.Name)
			writeFishFlags(w, name, sub.Flags, false, sub.Name)
			writeFishOptions(w, name, sub.Options, false, sub.Name)
		}
	}

	fmt.Fprintln(w)
	return nil
}

func writeFishFlags(w io.Writer, cmd string, flags []shedoc.Flag, noSubcmd bool, subName string) {
	for _, f := range flags {
		fmt.Fprintf(w, "complete -c %s", cmd)
		if subName != "" {
			fmt.Fprintf(w, " -n '__fish_seen_subcommand_from %s'", subName)
		} else if noSubcmd {
			fmt.Fprintf(w, " -n '__fish_use_subcommand'")
		}
		if f.Short != "" {
			fmt.Fprintf(w, " -s %s", f.Short[1:]) // strip leading -
		}
		if f.Long != "" {
			fmt.Fprintf(w, " -l %s", f.Long[2:]) // strip leading --
		}
		if f.Description != "" {
			fmt.Fprintf(w, " -d '%s'", fishEscape(f.Description))
		}
		fmt.Fprintln(w)
	}
}

func writeFishOptions(w io.Writer, cmd string, options []shedoc.Option, noSubcmd bool, subName string) {
	for _, o := range options {
		fmt.Fprintf(w, "complete -c %s", cmd)
		if subName != "" {
			fmt.Fprintf(w, " -n '__fish_seen_subcommand_from %s'", subName)
		} else if noSubcmd {
			fmt.Fprintf(w, " -n '__fish_use_subcommand'")
		}
		if o.Short != "" {
			fmt.Fprintf(w, " -s %s", o.Short[1:])
		}
		if o.Long != "" {
			fmt.Fprintf(w, " -l %s", o.Long[2:])
		}
		fmt.Fprintf(w, " -r") // requires argument
		if o.Description != "" {
			fmt.Fprintf(w, " -d '%s'", fishEscape(o.Description))
		}
		fmt.Fprintln(w)
	}
}

func fishEscape(s string) string {
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		if s[i] == '\'' {
			result = append(result, '\\', '\'')
		} else {
			result = append(result, s[i])
		}
	}
	return string(result)
}
