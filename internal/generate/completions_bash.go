package generate

import (
	"fmt"
	"io"
	"strings"

	"github.com/nickawilliams/shedoc"
)

func init() {
	shedoc.RegisterFormatter("completion:bash", &BashCompletionFormatter{})
}

// BashCompletionFormatter generates a bash completion script.
type BashCompletionFormatter struct{}

func (f *BashCompletionFormatter) Format(w io.Writer, doc *shedoc.Document) error {
	name := doc.Meta.Name
	if name == "" {
		return fmt.Errorf("completion generation requires #?/name")
	}

	funcName := strings.ReplaceAll(name, "-", "_")

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

	fmt.Fprintf(w, "# bash completion for %s\n", name)
	fmt.Fprintf(w, "_%s() {\n", funcName)
	fmt.Fprintf(w, "  local cur prev words cword\n")
	fmt.Fprintf(w, "  _init_completion || return\n")
	fmt.Fprintln(w)

	// Collect global flags/options
	var globalFlags []string
	if cmdBlock != nil {
		for _, flag := range cmdBlock.Flags {
			if flag.Short != "" {
				globalFlags = append(globalFlags, flag.Short)
			}
			if flag.Long != "" {
				globalFlags = append(globalFlags, flag.Long)
			}
		}
		for _, opt := range cmdBlock.Options {
			if opt.Short != "" {
				globalFlags = append(globalFlags, opt.Short)
			}
			if opt.Long != "" {
				globalFlags = append(globalFlags, opt.Long)
			}
		}
	}

	if len(subcommands) > 0 {
		// Subcommand names
		var subNames []string
		for _, sub := range subcommands {
			subNames = append(subNames, sub.Name)
		}

		fmt.Fprintf(w, "  local commands=\"%s\"\n", strings.Join(subNames, " "))
		fmt.Fprintln(w)

		// Subcommand-specific completions
		fmt.Fprintf(w, "  # Complete subcommand-specific flags\n")
		fmt.Fprintf(w, "  local i cmd\n")
		fmt.Fprintf(w, "  for ((i=1; i < cword; i++)); do\n")
		fmt.Fprintf(w, "    case \"${words[i]}\" in\n")
		for _, sub := range subcommands {
			subFlags := collectFlags(sub)
			if len(subFlags) > 0 {
				fmt.Fprintf(w, "      %s)\n", sub.Name)
				fmt.Fprintf(w, "        COMPREPLY=($(compgen -W \"%s\" -- \"$cur\"))\n", strings.Join(subFlags, " "))
				fmt.Fprintf(w, "        return\n")
				fmt.Fprintf(w, "        ;;\n")
			}
		}
		fmt.Fprintf(w, "    esac\n")
		fmt.Fprintf(w, "  done\n")
		fmt.Fprintln(w)

		// Top-level: complete subcommands and global flags
		allCompletions := append(subNames, globalFlags...)
		fmt.Fprintf(w, "  COMPREPLY=($(compgen -W \"%s\" -- \"$cur\"))\n", strings.Join(allCompletions, " "))
	} else if len(globalFlags) > 0 {
		fmt.Fprintf(w, "  COMPREPLY=($(compgen -W \"%s\" -- \"$cur\"))\n", strings.Join(globalFlags, " "))
	}

	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "complete -F _%s %s\n", funcName, name)
	return nil
}

func collectFlags(block shedoc.Block) []string {
	var flags []string
	for _, f := range block.Flags {
		if f.Short != "" {
			flags = append(flags, f.Short)
		}
		if f.Long != "" {
			flags = append(flags, f.Long)
		}
	}
	for _, o := range block.Options {
		if o.Short != "" {
			flags = append(flags, o.Short)
		}
		if o.Long != "" {
			flags = append(flags, o.Long)
		}
	}
	return flags
}
