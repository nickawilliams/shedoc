package generate

import (
	"fmt"
	"io"
	"strings"

	"github.com/nickawilliams/shedoc"
)

func init() {
	shedoc.RegisterFormatter("completion:zsh", &ZshCompletionFormatter{})
}

// ZshCompletionFormatter generates a zsh completion script.
type ZshCompletionFormatter struct{}

func (f *ZshCompletionFormatter) Format(w io.Writer, doc *shedoc.Document) error {
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

	fmt.Fprintf(w, "#compdef %s\n\n", name)
	fmt.Fprintf(w, "_%s() {\n", name)

	if len(subcommands) > 0 {
		// Global arguments
		fmt.Fprintf(w, "  local -a global_args\n")
		fmt.Fprintf(w, "  global_args=(\n")
		if cmdBlock != nil {
			writeZshFlags(w, cmdBlock.Flags)
			writeZshOptions(w, cmdBlock.Options)
		}
		fmt.Fprintf(w, "    '1:command:->commands'\n")
		fmt.Fprintf(w, "    '*::arg:->args'\n")
		fmt.Fprintf(w, "  )\n\n")

		fmt.Fprintf(w, "  _arguments -s $global_args\n\n")
		fmt.Fprintf(w, "  case $state in\n")
		fmt.Fprintf(w, "    commands)\n")
		fmt.Fprintf(w, "      local -a commands\n")
		fmt.Fprintf(w, "      commands=(\n")
		for _, sub := range subcommands {
			desc := firstLine(sub.Description)
			if sub.Deprecated != nil {
				desc = "[deprecated] " + sub.Deprecated.Message
			}
			desc = strings.ReplaceAll(desc, "'", "'\\''")
			fmt.Fprintf(w, "        '%s:%s'\n", sub.Name, desc)
		}
		fmt.Fprintf(w, "      )\n")
		fmt.Fprintf(w, "      _describe 'command' commands\n")
		fmt.Fprintf(w, "      ;;\n")

		fmt.Fprintf(w, "    args)\n")
		fmt.Fprintf(w, "      case $words[1] in\n")
		for _, sub := range subcommands {
			subFlags := collectZshArgs(sub)
			if len(subFlags) > 0 {
				fmt.Fprintf(w, "        %s)\n", sub.Name)
				fmt.Fprintf(w, "          _arguments -s \\\n")
				for i, arg := range subFlags {
					if i < len(subFlags)-1 {
						fmt.Fprintf(w, "            %s \\\n", arg)
					} else {
						fmt.Fprintf(w, "            %s\n", arg)
					}
				}
				fmt.Fprintf(w, "          ;;\n")
			}
		}
		fmt.Fprintf(w, "      esac\n")
		fmt.Fprintf(w, "      ;;\n")
		fmt.Fprintf(w, "  esac\n")
	} else {
		// No subcommands — just flags/options
		fmt.Fprintf(w, "  _arguments -s \\\n")
		var args []string
		if cmdBlock != nil {
			args = collectZshArgs(*cmdBlock)
		}
		for i, arg := range args {
			if i < len(args)-1 {
				fmt.Fprintf(w, "    %s \\\n", arg)
			} else {
				fmt.Fprintf(w, "    %s\n", arg)
			}
		}
	}

	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "_%s\n", name)
	return nil
}

func writeZshFlags(w io.Writer, flags []shedoc.Flag) {
	for _, f := range flags {
		desc := strings.ReplaceAll(f.Description, "'", "'\\''")
		if f.Short != "" && f.Long != "" {
			fmt.Fprintf(w, "    '(%s %s)'{%s,%s}'[%s]'\n", f.Short, f.Long, f.Short, f.Long, desc)
		} else if f.Long != "" {
			fmt.Fprintf(w, "    '%s[%s]'\n", f.Long, desc)
		} else if f.Short != "" {
			fmt.Fprintf(w, "    '%s[%s]'\n", f.Short, desc)
		}
	}
}

func writeZshOptions(w io.Writer, options []shedoc.Option) {
	for _, o := range options {
		desc := strings.ReplaceAll(o.Description, "'", "'\\''")
		valDesc := o.Value.Name
		if o.Short != "" && o.Long != "" {
			fmt.Fprintf(w, "    '(%s %s)'{%s,%s}'[%s]:%s:'\n", o.Short, o.Long, o.Short, o.Long, desc, valDesc)
		} else if o.Long != "" {
			fmt.Fprintf(w, "    '%s[%s]:%s:'\n", o.Long, desc, valDesc)
		} else if o.Short != "" {
			fmt.Fprintf(w, "    '%s[%s]:%s:'\n", o.Short, desc, valDesc)
		}
	}
}

func collectZshArgs(block shedoc.Block) []string {
	var args []string
	for _, f := range block.Flags {
		desc := strings.ReplaceAll(f.Description, "'", "'\\''")
		if f.Short != "" && f.Long != "" {
			args = append(args, fmt.Sprintf("'(%s %s)'{%s,%s}'[%s]'", f.Short, f.Long, f.Short, f.Long, desc))
		} else if f.Long != "" {
			args = append(args, fmt.Sprintf("'%s[%s]'", f.Long, desc))
		} else if f.Short != "" {
			args = append(args, fmt.Sprintf("'%s[%s]'", f.Short, desc))
		}
	}
	for _, o := range block.Options {
		desc := strings.ReplaceAll(o.Description, "'", "'\\''")
		valDesc := o.Value.Name
		if o.Short != "" && o.Long != "" {
			args = append(args, fmt.Sprintf("'(%s %s)'{%s,%s}'[%s]:%s:'", o.Short, o.Long, o.Short, o.Long, desc, valDesc))
		} else if o.Long != "" {
			args = append(args, fmt.Sprintf("'%s[%s]:%s:'", o.Long, desc, valDesc))
		} else if o.Short != "" {
			args = append(args, fmt.Sprintf("'%s[%s]:%s:'", o.Short, desc, valDesc))
		}
	}
	return args
}
