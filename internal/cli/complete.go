package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nickawilliams/shedoc"
	"github.com/spf13/cobra"
)

var (
	flagCompleteShell string
	flagCompleteSetup string
)

func newCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete [flags] <file>",
		Short: "Dynamic shell completion for shedoc-annotated scripts",
		Long: `Two modes of operation:

  Handler mode (invoked at tab-press time by the shell):
    shedoc complete deploy.sh
    shedoc complete --shell fish deploy.sh

  Setup mode (run once to configure your shell):
    shedoc complete --setup bash deploy.sh
    shedoc complete --setup zsh deploy.sh
    shedoc complete --setup fish deploy.sh`,
		Args:          cobra.MinimumNArgs(1),
		RunE:          runComplete,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().StringVar(&flagCompleteShell, "shell", "bash", "output format for handler mode (bash, fish)")
	cmd.Flags().StringVar(&flagCompleteSetup, "setup", "", "output shell registration code (bash, zsh, fish)")

	cmd.MarkFlagsMutuallyExclusive("shell", "setup")

	return cmd
}

func runComplete(cmd *cobra.Command, args []string) error {
	scriptPath := args[0]

	w := cmd.OutOrStdout()

	if flagCompleteSetup != "" {
		return runCompleteSetup(w, scriptPath, flagCompleteSetup)
	}

	return runCompleteHandler(w, scriptPath, flagCompleteShell)
}

// runCompleteSetup outputs shell-specific registration code.
func runCompleteSetup(w io.Writer, scriptPath, shell string) error {
	doc, err := shedoc.Parse(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", scriptPath, err)
	}

	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	cmdName := doc.Meta.Name
	if cmdName == "" {
		cmdName = strings.TrimSuffix(filepath.Base(scriptPath), filepath.Ext(scriptPath))
	}

	switch shell {
	case "bash":
		fmt.Fprintf(w, "complete -C \"shedoc complete %s\" %s\n", absPath, cmdName)
	case "zsh":
		funcName := "_" + strings.ReplaceAll(cmdName, "-", "_") + "_shedoc"
		fmt.Fprintf(w, "%s() {\n", funcName)
		fmt.Fprintf(w, "  local COMP_LINE COMP_POINT\n")
		fmt.Fprintf(w, "  COMP_LINE=\"${words[*]}\"\n")
		fmt.Fprintf(w, "  COMP_POINT=${#COMP_LINE}\n")
		fmt.Fprintf(w, "  local completions\n")
		fmt.Fprintf(w, "  completions=($(COMP_LINE=\"$COMP_LINE\" COMP_POINT=\"$COMP_POINT\" shedoc complete %s))\n", absPath)
		fmt.Fprintf(w, "  compadd -a completions\n")
		fmt.Fprintf(w, "}\n")
		fmt.Fprintf(w, "compdef %s %s\n", funcName, cmdName)
	case "fish":
		fmt.Fprintf(w, "complete -c %s -a '(COMP_LINE=(commandline) COMP_POINT=(commandline -C) shedoc complete --shell fish %s)'\n", cmdName, absPath)
	default:
		return fmt.Errorf("unsupported shell: %q (supported: bash, zsh, fish)", shell)
	}

	return nil
}

// runCompleteHandler reads COMP_LINE/COMP_POINT, parses the script, and outputs
// matching completions.
func runCompleteHandler(w io.Writer, scriptPath, shell string) error {
	compLine := os.Getenv("COMP_LINE")
	if compLine == "" {
		return nil // no completion context, nothing to output
	}

	compPoint := len(compLine)
	if cp := os.Getenv("COMP_POINT"); cp != "" {
		_, _ = fmt.Sscanf(cp, "%d", &compPoint)
	}

	doc, err := shedoc.Parse(scriptPath)
	if err != nil {
		return nil // silently fail during completion
	}

	candidates := completionCandidates(doc, compLine, compPoint)
	for _, c := range candidates {
		if shell == "fish" {
			desc := strings.ReplaceAll(c.description, "\t", " ")
			fmt.Fprintf(w, "%s\t%s\n", c.word, desc)
		} else {
			fmt.Fprintln(w, c.word)
		}
	}
	return nil
}

type candidate struct {
	word        string
	description string
}

// completionCandidates determines the available completions given the document
// and current input state.
func completionCandidates(doc *shedoc.Document, compLine string, compPoint int) []candidate {
	// Truncate at cursor position.
	if compPoint < len(compLine) {
		compLine = compLine[:compPoint]
	}

	words := strings.Fields(compLine)

	// Determine if we're completing a new (empty) word or a partial word.
	// If the line ends with whitespace, cursor is on a new empty word.
	endsWithSpace := len(compLine) > 0 && compLine[len(compLine)-1] == ' '

	var curWord string
	if !endsWithSpace && len(words) > 1 {
		curWord = words[len(words)-1]
		words = words[:len(words)-1]
	} else if !endsWithSpace && len(words) == 1 {
		// Only the command name, partially typed — nothing to complete
		return nil
	}

	// Skip words[0] — it's the command name itself.
	if len(words) > 0 {
		words = words[1:]
	}

	// Extract command and subcommand blocks.
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

	// No command block and no subcommands — nothing to complete.
	if cmdBlock == nil && len(subcommands) == 0 {
		return nil
	}

	// Find if a subcommand has been specified.
	var matchedSub *shedoc.Block
	for _, w := range words {
		for i := range subcommands {
			if subcommands[i].Name == w {
				matchedSub = &subcommands[i]
				break
			}
		}
		if matchedSub != nil {
			break
		}
	}

	// Check if prevWord is an option that takes a value — suppress completions.
	prevWord := ""
	if len(words) > 0 {
		prevWord = words[len(words)-1]
	}
	// When !endsWithSpace && curWord != "", curWord is part of words
	// and prevWord stays empty — no special handling needed.

	if prevWord != "" && isValueOption(prevWord, cmdBlock, matchedSub) {
		return nil
	}

	// Build candidate list.
	var candidates []candidate

	if matchedSub != nil {
		// Inside a subcommand: subcommand-specific flags + global flags.
		candidates = append(candidates, flagCandidates(matchedSub)...)
		if cmdBlock != nil {
			candidates = append(candidates, flagCandidates(cmdBlock)...)
		}
	} else {
		// Top-level: subcommand names + global flags.
		for _, sub := range subcommands {
			desc := firstLineCli(sub.Description)
			if sub.Deprecated != nil {
				desc = "[deprecated] " + sub.Deprecated.Message
			}
			candidates = append(candidates, candidate{word: sub.Name, description: desc})
		}
		if cmdBlock != nil {
			candidates = append(candidates, flagCandidates(cmdBlock)...)
		}
	}

	// Filter by prefix.
	if curWord != "" {
		var filtered []candidate
		for _, c := range candidates {
			if strings.HasPrefix(c.word, curWord) {
				filtered = append(filtered, c)
			}
		}
		return filtered
	}

	return candidates
}

// flagCandidates returns completion candidates for all flags and options in a block.
func flagCandidates(block *shedoc.Block) []candidate {
	var cs []candidate
	for _, f := range block.Flags {
		if f.Short != "" {
			cs = append(cs, candidate{word: f.Short, description: f.Description})
		}
		if f.Long != "" {
			cs = append(cs, candidate{word: f.Long, description: f.Description})
		}
	}
	for _, o := range block.Options {
		if o.Short != "" {
			cs = append(cs, candidate{word: o.Short, description: o.Description})
		}
		if o.Long != "" {
			cs = append(cs, candidate{word: o.Long, description: o.Description})
		}
	}
	return cs
}

// isValueOption checks if the given word is an option (not flag) that expects a value.
func isValueOption(word string, blocks ...*shedoc.Block) bool {
	for _, b := range blocks {
		if b == nil {
			continue
		}
		for _, o := range b.Options {
			if o.Short == word || o.Long == word {
				return true
			}
		}
	}
	return false
}

// firstLineCli returns the first line of a potentially multi-line string.
func firstLineCli(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
