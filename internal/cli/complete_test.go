package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nickawilliams/shedoc"
)

func parseTestDoc(t *testing.T) *shedoc.Document {
	t.Helper()
	doc, err := shedoc.Parse(filepath.Join("..", "..", "testdata", "comprehensive.sh"))
	if err != nil {
		t.Fatalf("failed to parse comprehensive.sh: %v", err)
	}
	return doc
}

func TestCompletionCandidates_TopLevel(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy " — cursor after space, should get subcommands + global flags
	candidates := completionCandidates(doc, "deploy ", 7)

	// Should contain subcommand names
	names := candidateWords(candidates)
	for _, want := range []string{"push", "status", "rollback", "migrate"} {
		if !contains(names, want) {
			t.Errorf("expected subcommand %q in candidates, got %v", want, names)
		}
	}
	// Should contain global flags
	for _, want := range []string{"-v", "--verbose", "-c", "--config"} {
		if !contains(names, want) {
			t.Errorf("expected global flag %q in candidates, got %v", want, names)
		}
	}
}

func TestCompletionCandidates_TopLevelPrefix(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy p" — partial word "p", should match "push"
	candidates := completionCandidates(doc, "deploy p", 8)
	names := candidateWords(candidates)
	if !contains(names, "push") {
		t.Errorf("expected 'push' in candidates, got %v", names)
	}
	if contains(names, "status") {
		t.Errorf("should not contain 'status' when filtering by 'p', got %v", names)
	}
}

func TestCompletionCandidates_FlagPrefix(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy --" — partial word "--", should match --verbose and --config
	candidates := completionCandidates(doc, "deploy --", 9)
	names := candidateWords(candidates)
	for _, want := range []string{"--verbose", "--config"} {
		if !contains(names, want) {
			t.Errorf("expected %q in candidates, got %v", want, names)
		}
	}
	// Should not contain short flags
	if contains(names, "-v") {
		t.Errorf("should not contain '-v' when filtering by '--', got %v", names)
	}
}

func TestCompletionCandidates_Subcommand(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy push " — inside push subcommand, should get push flags + global flags
	candidates := completionCandidates(doc, "deploy push ", 12)
	names := candidateWords(candidates)
	// push-specific flags
	for _, want := range []string{"-f", "--force", "--dry-run", "--tag"} {
		if !contains(names, want) {
			t.Errorf("expected push flag %q in candidates, got %v", want, names)
		}
	}
	// global flags should also be present
	for _, want := range []string{"-v", "--verbose"} {
		if !contains(names, want) {
			t.Errorf("expected global flag %q in candidates, got %v", want, names)
		}
	}
	// Should NOT contain other subcommand names
	if contains(names, "status") {
		t.Errorf("should not contain subcommand names inside push, got %v", names)
	}
}

func TestCompletionCandidates_SubcommandFlagPrefix(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy push --d" — filtering push flags by --d
	candidates := completionCandidates(doc, "deploy push --d", 15)
	names := candidateWords(candidates)
	if !contains(names, "--dry-run") {
		t.Errorf("expected '--dry-run' in candidates, got %v", names)
	}
	if contains(names, "--force") {
		t.Errorf("should not contain '--force' when filtering by '--d', got %v", names)
	}
}

func TestCompletionCandidates_AfterValueOption(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy --config " — --config takes a value, should suppress completions
	candidates := completionCandidates(doc, "deploy --config ", 16)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates after value option, got %v", candidateWords(candidates))
	}
}

func TestCompletionCandidates_AfterValueOptionShort(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy -c " — -c takes a value, should suppress completions
	candidates := completionCandidates(doc, "deploy -c ", 10)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates after short value option, got %v", candidateWords(candidates))
	}
}

func TestCompletionCandidates_AfterValueOptionInSubcommand(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy push --tag " — --tag takes a value, should suppress
	candidates := completionCandidates(doc, "deploy push --tag ", 18)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates after subcommand value option, got %v", candidateWords(candidates))
	}
}

func TestCompletionCandidates_NoBlocks(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{Name: "empty"},
	}
	candidates := completionCandidates(doc, "empty ", 6)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates for script with no blocks, got %v", candidateWords(candidates))
	}
}

func TestCompletionCandidates_OnlyCommandName(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy" — just the command name, no space, nothing to complete
	candidates := completionCandidates(doc, "deploy", 6)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates for bare command name, got %v", candidateWords(candidates))
	}
}

func TestRunCompleteHandler_BashOutput(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")

	t.Setenv("COMP_LINE", "deploy ")
	t.Setenv("COMP_POINT", "7")

	var buf bytes.Buffer
	err := runCompleteHandler(&buf, scriptPath, "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, want := range []string{"push", "status", "rollback"} {
		if !contains(lines, want) {
			t.Errorf("expected %q in bash output, got: %s", want, output)
		}
	}
}

func TestRunCompleteHandler_FishOutput(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")

	t.Setenv("COMP_LINE", "deploy ")
	t.Setenv("COMP_POINT", "7")

	var buf bytes.Buffer
	err := runCompleteHandler(&buf, scriptPath, "fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Fish format should have tab-separated word\tdescription
	if !strings.Contains(output, "push\t") {
		t.Errorf("expected fish format with tab separator, got: %s", output)
	}
}

func TestRunCompleteHandler_NoCompLine(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")

	// Ensure COMP_LINE is not set
	os.Unsetenv("COMP_LINE")

	var buf bytes.Buffer
	err := runCompleteHandler(&buf, scriptPath, "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output without COMP_LINE, got: %s", buf.String())
	}
}

func TestRunCompleteSetup_Bash(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")
	absPath, _ := filepath.Abs(scriptPath)

	var buf bytes.Buffer
	err := runCompleteSetup(&buf, scriptPath, "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should contain complete -C with absolute path
	if !strings.Contains(output, "complete -C") {
		t.Errorf("expected 'complete -C' in bash setup, got: %s", output)
	}
	if !strings.Contains(output, absPath) {
		t.Errorf("expected absolute path %q in bash setup, got: %s", absPath, output)
	}
	if !strings.Contains(output, "deploy") {
		t.Errorf("expected command name 'deploy' in bash setup, got: %s", output)
	}
}

func TestRunCompleteSetup_Zsh(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")

	var buf bytes.Buffer
	err := runCompleteSetup(&buf, scriptPath, "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "COMP_LINE") {
		t.Errorf("expected COMP_LINE in zsh setup, got: %s", output)
	}
	if !strings.Contains(output, "COMP_POINT") {
		t.Errorf("expected COMP_POINT in zsh setup, got: %s", output)
	}
	if !strings.Contains(output, "compadd") {
		t.Errorf("expected compadd in zsh setup, got: %s", output)
	}
	if !strings.Contains(output, "compdef") {
		t.Errorf("expected compdef in zsh setup, got: %s", output)
	}
}

func TestRunCompleteSetup_Fish(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")
	absPath, _ := filepath.Abs(scriptPath)

	var buf bytes.Buffer
	err := runCompleteSetup(&buf, scriptPath, "fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "complete -c deploy") {
		t.Errorf("expected 'complete -c deploy' in fish setup, got: %s", output)
	}
	if !strings.Contains(output, "--shell fish") {
		t.Errorf("expected '--shell fish' in fish setup, got: %s", output)
	}
	if !strings.Contains(output, absPath) {
		t.Errorf("expected absolute path in fish setup, got: %s", output)
	}
}

func TestRunCompleteSetup_InvalidShell(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "testdata", "comprehensive.sh")

	var buf bytes.Buffer
	err := runCompleteSetup(&buf, scriptPath, "powershell")
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected 'unsupported shell' in error, got: %v", err)
	}
}

func TestRunCompleteSetup_FallbackName(t *testing.T) {
	// Create a temp script with no #?/name
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "my-tool.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\n#@/command\n # @flag --help Show help\n ##\n"), 0o644)

	var buf bytes.Buffer
	err := runCompleteSetup(&buf, scriptPath, "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should fall back to filename without extension
	if !strings.Contains(output, "my-tool") {
		t.Errorf("expected fallback name 'my-tool' in setup, got: %s", output)
	}
}

func TestCompletionCandidates_FishDescriptions(t *testing.T) {
	doc := parseTestDoc(t)

	candidates := completionCandidates(doc, "deploy ", 7)

	// Subcommands should have descriptions
	for _, c := range candidates {
		if c.word == "push" && c.description == "" {
			t.Error("expected push subcommand to have a description")
		}
		if c.word == "migrate" && !strings.Contains(c.description, "deprecated") {
			t.Errorf("expected migrate to have deprecated description, got: %q", c.description)
		}
	}
}

func TestCompletionCandidates_StatusSubcommand(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy status " — inside status subcommand
	candidates := completionCandidates(doc, "deploy status ", 14)
	names := candidateWords(candidates)
	if !contains(names, "--format") {
		t.Errorf("expected '--format' in status candidates, got %v", names)
	}
}

func TestCompletionCandidates_AfterFormatOption(t *testing.T) {
	doc := parseTestDoc(t)

	// "deploy status --format " — --format takes value, suppress
	candidates := completionCandidates(doc, "deploy status --format ", 23)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates after --format (value option), got %v", candidateWords(candidates))
	}
}

// helpers

func candidateWords(cs []candidate) []string {
	words := make([]string, len(cs))
	for i, c := range cs {
		words[i] = c.word
	}
	return words
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
