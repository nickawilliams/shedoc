package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nickawilliams/shedoc"
	_ "github.com/nickawilliams/shedoc/internal/generate" // register formatters
	"github.com/spf13/cobra"
)

var (
	flagTo       string
	flagGet      string
	flagOutput   string
	flagWarnings bool
	flagQuiet    bool
)

// NewRootCmd creates the root shedoc command.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shedoc [flags] <file...>",
		Short:   "Parse and output shell script documentation",
		Version: version,
		Args:    cobra.MinimumNArgs(1),
		RunE:    runRoot,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().StringVarP(&flagTo, "to", "t", "json", "output format (json, help, man, completion:bash, completion:zsh, completion:fish)")
	cmd.Flags().StringVarP(&flagGet, "get", "g", "", "extract a single #?/ tag value")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "write output to file instead of stdout")
	cmd.Flags().BoolVarP(&flagWarnings, "warnings", "w", false, "include warnings in output")
	cmd.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "suppress warnings on stderr")

	cmd.MarkFlagsMutuallyExclusive("to", "get")

	cmd.AddCommand(newCompleteCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Determine output writer.
	var w io.Writer = cmd.OutOrStdout()
	if flagOutput != "" {
		f, err := os.Create(flagOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	// Parse input files.
	docs, err := parseFiles(args)
	if err != nil {
		return err
	}

	// Emit warnings to stderr if not suppressed.
	if !flagQuiet {
		for _, doc := range docs {
			for _, warn := range doc.Warnings {
				source := doc.Path
				if source == "" {
					source = "<stdin>"
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d: warning: %s\n", source, warn.Line, warn.Message)
			}
		}
	}

	// Strip warnings from output unless explicitly requested.
	if !flagWarnings {
		for i := range docs {
			docs[i].Warnings = nil
		}
	}

	// Handle --get: extract a single meta tag value.
	if flagGet != "" {
		return runGet(w, docs)
	}

	// Non-JSON formats accept a single file only.
	if flagTo != "json" && len(docs) > 1 {
		return fmt.Errorf("format %q supports a single file; got %d", flagTo, len(docs))
	}

	// Look up formatter.
	formatter := shedoc.GetFormatter(flagTo)
	if formatter == nil {
		return fmt.Errorf("unknown format: %q\navailable formats: %s", flagTo, strings.Join(shedoc.RegisteredFormats(), ", "))
	}

	// Output.
	if len(docs) == 1 {
		return formatter.Format(w, docs[0])
	}

	// Multiple files: NDJSON (one JSON object per line).
	for _, doc := range docs {
		if err := formatter.Format(w, doc); err != nil {
			return err
		}
	}
	return nil
}

func runGet(w io.Writer, docs []*shedoc.Document) error {
	for _, doc := range docs {
		val, ok := getMetaField(&doc.Meta, flagGet)
		if !ok {
			return fmt.Errorf("unknown tag: %q", flagGet)
		}
		if val != "" {
			fmt.Fprintln(w, val)
		}
	}
	return nil
}

func getMetaField(m *shedoc.Meta, tag string) (string, bool) {
	switch tag {
	case "name":
		return m.Name, true
	case "version":
		return m.Version, true
	case "synopsis":
		return m.Synopsis, true
	case "description":
		return m.Description, true
	case "examples":
		return m.Examples, true
	case "section":
		return m.Section, true
	case "author":
		return m.Author, true
	case "license":
		return m.License, true
	default:
		return "", false
	}
}

func parseFiles(args []string) ([]*shedoc.Document, error) {
	var docs []*shedoc.Document
	for _, arg := range args {
		if arg == "-" {
			doc, err := shedoc.ParseReader(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("failed to parse stdin: %w", err)
			}
			docs = append(docs, doc)
			continue
		}

		doc, err := shedoc.Parse(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", arg, err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
