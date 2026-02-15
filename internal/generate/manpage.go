package generate

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nickawilliams/shedoc"
)

func init() {
	shedoc.RegisterFormatter("man", &ManPageFormatter{})
}

// ManPageFormatter outputs a Document as a troff/groff man page.
type ManPageFormatter struct{}

func (f *ManPageFormatter) Format(w io.Writer, doc *shedoc.Document) error {
	section := doc.Meta.Section
	if section == "" {
		section = "1"
	}

	name := doc.Meta.Name
	if name == "" {
		name = "UNKNOWN"
	}

	date := time.Now().Format("2006-01-02")
	version := doc.Meta.Version

	// .TH header
	fmt.Fprintf(w, ".TH %s %s %q %q\n",
		troffEscape(strings.ToUpper(name)),
		section,
		date,
		version,
	)

	// NAME section
	fmt.Fprintln(w, ".SH NAME")
	if doc.Meta.Description != "" {
		brief := firstLine(doc.Meta.Description)
		fmt.Fprintf(w, "%s \\- %s\n", troffEscape(name), troffEscape(brief))
	} else {
		fmt.Fprintln(w, troffEscape(name))
	}

	// SYNOPSIS section
	if doc.Meta.Synopsis != "" {
		fmt.Fprintln(w, ".SH SYNOPSIS")
		fmt.Fprintf(w, ".B %s\n", troffEscape(doc.Meta.Synopsis))
	}

	// DESCRIPTION section
	if doc.Meta.Description != "" {
		fmt.Fprintln(w, ".SH DESCRIPTION")
		writeManText(w, doc.Meta.Description)
	}

	// Find command block and subcommands.
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

	// OPTIONS section
	if cmdBlock != nil && (len(cmdBlock.Flags) > 0 || len(cmdBlock.Options) > 0) {
		fmt.Fprintln(w, ".SH OPTIONS")
		for _, flag := range cmdBlock.Flags {
			label := formatFlagLabel(flag.Short, flag.Long)
			fmt.Fprintf(w, ".TP\n.B %s\n", troffEscape(label))
			if flag.Description != "" {
				writeManText(w, flag.Description)
			}
		}
		for _, opt := range cmdBlock.Options {
			label := formatOptionLabel(opt.Short, opt.Long, opt.Value)
			fmt.Fprintf(w, ".TP\n.B %s\n", troffEscape(label))
			if opt.Description != "" {
				writeManText(w, opt.Description)
			}
		}
	}

	// COMMANDS section
	if len(subcommands) > 0 {
		fmt.Fprintln(w, ".SH COMMANDS")
		for _, sub := range subcommands {
			fmt.Fprintf(w, ".TP\n.B %s\n", troffEscape(sub.Name))
			if sub.Deprecated != nil {
				msg := sub.Deprecated.Message
				if msg == "" {
					msg = "This command is deprecated."
				}
				fmt.Fprintf(w, "[deprecated] %s\n", troffEscape(msg))
			} else if sub.Description != "" {
				writeManText(w, sub.Description)
			}

			// Subcommand flags and options
			for _, flag := range sub.Flags {
				label := formatFlagLabel(flag.Short, flag.Long)
				fmt.Fprintf(w, ".RS\n.TP\n.B %s\n", troffEscape(label))
				if flag.Description != "" {
					writeManText(w, flag.Description)
				}
				fmt.Fprintln(w, ".RE")
			}
			for _, opt := range sub.Options {
				label := formatOptionLabel(opt.Short, opt.Long, opt.Value)
				fmt.Fprintf(w, ".RS\n.TP\n.B %s\n", troffEscape(label))
				if opt.Description != "" {
					writeManText(w, opt.Description)
				}
				fmt.Fprintln(w, ".RE")
			}
		}
	}

	// ENVIRONMENT section
	var envVars []shedoc.Env
	if cmdBlock != nil {
		envVars = cmdBlock.Env
	}
	if len(envVars) > 0 {
		fmt.Fprintln(w, ".SH ENVIRONMENT")
		for _, env := range envVars {
			fmt.Fprintf(w, ".TP\n.B %s\n", troffEscape(env.Name))
			if env.Description != "" {
				writeManText(w, env.Description)
			}
		}
	}

	// FILES section
	var files []struct{ path, desc string }
	if cmdBlock != nil {
		for _, r := range cmdBlock.Reads {
			files = append(files, struct{ path, desc string }{r.Path, r.Description})
		}
		for _, wr := range cmdBlock.Writes {
			files = append(files, struct{ path, desc string }{wr.Path, wr.Description})
		}
	}
	if len(files) > 0 {
		fmt.Fprintln(w, ".SH FILES")
		for _, f := range files {
			fmt.Fprintf(w, ".TP\n.B %s\n", troffEscape(f.path))
			if f.desc != "" {
				writeManText(w, f.desc)
			}
		}
	}

	// EXIT STATUS section
	if cmdBlock != nil && len(cmdBlock.Exit) > 0 {
		fmt.Fprintln(w, ".SH EXIT STATUS")
		for _, exit := range cmdBlock.Exit {
			fmt.Fprintf(w, ".TP\n.B %s\n", troffEscape(exit.Code))
			if exit.Description != "" {
				writeManText(w, exit.Description)
			}
		}
	}

	// EXAMPLES section
	if doc.Meta.Examples != "" {
		fmt.Fprintln(w, ".SH EXAMPLES")
		for _, line := range strings.Split(doc.Meta.Examples, "\n") {
			fmt.Fprintln(w, ".PP")
			fmt.Fprintf(w, ".B %s\n", troffEscape(line))
		}
	}

	// AUTHOR section
	if doc.Meta.Author != "" {
		fmt.Fprintln(w, ".SH AUTHOR")
		writeManText(w, doc.Meta.Author)
	}

	return nil
}

// troffEscape escapes special troff characters.
func troffEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "-", "\\-")
	return s
}

// writeManText writes a block of text as troff paragraphs.
func writeManText(w io.Writer, text string) {
	fmt.Fprintln(w, troffEscape(text))
}
