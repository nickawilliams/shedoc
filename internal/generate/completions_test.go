package generate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nickawilliams/shedoc"
)

var completionTestDoc = &shedoc.Document{
	Meta: shedoc.Meta{
		Name: "deploy",
	},
	Blocks: []shedoc.Block{
		{
			Visibility: shedoc.VisibilityCommand,
			Flags: []shedoc.Flag{
				{Short: "-v", Long: "--verbose", Description: "Enable verbose output"},
			},
			Options: []shedoc.Option{
				{Short: "-c", Long: "--config", Value: shedoc.Value{Name: "path", Required: true}, Description: "Config file"},
			},
		},
		{
			Visibility:  shedoc.VisibilitySubcommand,
			Name:        "push",
			Description: "Deploy the application.",
			Flags: []shedoc.Flag{
				{Short: "-f", Long: "--force", Description: "Skip confirmation"},
			},
		},
		{
			Visibility:  shedoc.VisibilitySubcommand,
			Name:        "status",
			Description: "Show deployment status.",
		},
	},
}

func TestBashCompletionFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &BashCompletionFormatter{}
	if err := f.Format(&buf, completionTestDoc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	checks := []string{
		"_deploy()",
		"complete -F _deploy deploy",
		"commands=\"push status\"",
		"-v",
		"--verbose",
		"-c",
		"--config",
		"-f",
		"--force",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("bash output missing %q\n\n%s", check, got)
		}
	}
}

func TestZshCompletionFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &ZshCompletionFormatter{}
	if err := f.Format(&buf, completionTestDoc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	checks := []string{
		"#compdef deploy",
		"_deploy()",
		"'push:Deploy the application.'",
		"'status:Show deployment status.'",
		"--verbose",
		"--config",
		"--force",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("zsh output missing %q\n\n%s", check, got)
		}
	}
}

func TestFishCompletionFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &FishCompletionFormatter{}
	if err := f.Format(&buf, completionTestDoc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	checks := []string{
		"complete -c deploy",
		"-a push",
		"-a status",
		"-s v",
		"-l verbose",
		"-s c",
		"-l config",
		"-s f",
		"-l force",
		"__fish_use_subcommand",
		"__fish_seen_subcommand_from push",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("fish output missing %q\n\n%s", check, got)
		}
	}
}

// Test with short-only and long-only flags to cover branch variants.
var completionTestDocMixedFlags = &shedoc.Document{
	Meta: shedoc.Meta{
		Name: "tool",
	},
	Blocks: []shedoc.Block{
		{
			Visibility: shedoc.VisibilityCommand,
			Flags: []shedoc.Flag{
				{Short: "-v", Description: "Verbose"},
				{Long: "--dry-run", Description: "Dry run"},
			},
			Options: []shedoc.Option{
				{Short: "-o", Value: shedoc.Value{Name: "file", Required: true}, Description: "Output file"},
				{Long: "--format", Value: shedoc.Value{Name: "fmt", Required: true}, Description: "Format"},
			},
		},
	},
}

func TestBashCompletionFormatter_MixedFlags(t *testing.T) {
	var buf bytes.Buffer
	f := &BashCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocMixedFlags); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, check := range []string{"-v", "--dry-run", "-o", "--format"} {
		if !strings.Contains(got, check) {
			t.Errorf("bash output missing %q\n\n%s", check, got)
		}
	}
}

func TestZshCompletionFormatter_MixedFlags(t *testing.T) {
	var buf bytes.Buffer
	f := &ZshCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocMixedFlags); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, check := range []string{"-v", "--dry-run", "-o", "--format"} {
		if !strings.Contains(got, check) {
			t.Errorf("zsh output missing %q\n\n%s", check, got)
		}
	}
}

func TestFishCompletionFormatter_MixedFlags(t *testing.T) {
	var buf bytes.Buffer
	f := &FishCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocMixedFlags); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, check := range []string{"-s v", "-l dry-run", "-s o", "-l format"} {
		if !strings.Contains(got, check) {
			t.Errorf("fish output missing %q\n\n%s", check, got)
		}
	}
}

func TestZshCompletionFormatter_NoSubcommands(t *testing.T) {
	var buf bytes.Buffer
	f := &ZshCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocMixedFlags); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "_arguments -s") {
		t.Errorf("zsh output missing _arguments for no-subcommand case\n\n%s", got)
	}
}

func TestFishCompletionFormatter_NoSubcommands(t *testing.T) {
	var buf bytes.Buffer
	f := &FishCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocMixedFlags); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Should not contain subcommand-related conditions.
	if strings.Contains(got, "__fish_use_subcommand") {
		t.Errorf("fish output should not contain __fish_use_subcommand for no-subcommand case\n\n%s", got)
	}
}

func TestBashCompletionFormatter_NoSubcommands(t *testing.T) {
	var buf bytes.Buffer
	f := &BashCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocMixedFlags); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "COMPREPLY") {
		t.Errorf("bash output missing COMPREPLY\n\n%s", got)
	}
}

// Test with subcommands that have mixed flag forms to cover writeZshFlags/writeZshOptions
// short-only, long-only, and both-short-and-long branches, plus collectFlags branches
// and fishEscape with apostrophes.
var completionTestDocSubcmdMixed = &shedoc.Document{
	Meta: shedoc.Meta{
		Name: "app",
	},
	Blocks: []shedoc.Block{
		{
			Visibility: shedoc.VisibilityCommand,
			Flags: []shedoc.Flag{
				{Short: "-v", Description: "Verbose"},
				{Long: "--dry-run", Description: "Dry run"},
			},
			Options: []shedoc.Option{
				{Short: "-o", Value: shedoc.Value{Name: "file", Required: true}, Description: "Output"},
				{Long: "--format", Value: shedoc.Value{Name: "fmt", Required: true}, Description: "Format"},
			},
		},
		{
			Visibility:  shedoc.VisibilitySubcommand,
			Name:        "run",
			Description: "Run the app.",
			Flags: []shedoc.Flag{
				{Short: "-f", Long: "--force", Description: "Force it"},
				{Short: "-q", Description: "Quiet"},
				{Long: "--no-cache", Description: "It's uncached"},
			},
			Options: []shedoc.Option{
				{Short: "-t", Long: "--target", Value: shedoc.Value{Name: "host", Required: true}, Description: "Target"},
				{Short: "-p", Value: shedoc.Value{Name: "port", Required: true}, Description: "Port"},
				{Long: "--timeout", Value: shedoc.Value{Name: "ms", Required: true}, Description: "Timeout"},
			},
		},
		{
			Visibility:  shedoc.VisibilitySubcommand,
			Name:        "stop",
			Description: "Stop the app.",
		},
	},
}

func TestBashCompletionFormatter_SubcmdMixed(t *testing.T) {
	var buf bytes.Buffer
	f := &BashCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocSubcmdMixed); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// collectFlags should gather all flag variants from the "run" subcommand.
	for _, check := range []string{"-f", "--force", "-q", "--no-cache", "-t", "--target", "-p", "--timeout"} {
		if !strings.Contains(got, check) {
			t.Errorf("bash output missing %q\n\n%s", check, got)
		}
	}
}

func TestZshCompletionFormatter_SubcmdMixed(t *testing.T) {
	var buf bytes.Buffer
	f := &ZshCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocSubcmdMixed); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// writeZshFlags: short+long, short-only, long-only
	for _, check := range []string{
		"'-v[Verbose]'", "'--dry-run",      // global short-only and long-only flags
		"(-f --force)", "'-q[Quiet]'",      // subcommand both and short-only flag
		"--no-cache",                       // subcommand long-only flag
		"(-t --target)", "'-p[Port]",       // subcommand both and short-only option
		"'--timeout",                       // subcommand long-only option
	} {
		if !strings.Contains(got, check) {
			t.Errorf("zsh output missing %q\n\n%s", check, got)
		}
	}
}

func TestFishCompletionFormatter_SubcmdMixed(t *testing.T) {
	var buf bytes.Buffer
	f := &FishCompletionFormatter{}
	if err := f.Format(&buf, completionTestDocSubcmdMixed); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// writeFishFlags/writeFishOptions with subcommand-specific flags
	for _, check := range []string{
		"-s f", "-l force",
		"-s q",
		"-l no-cache",
		"-s t", "-l target",
		"-s p",
		"-l timeout",
		"__fish_seen_subcommand_from run",
		// fishEscape: apostrophe in "It's uncached" → "It\'s uncached"
		"It\\'s uncached",
	} {
		if !strings.Contains(got, check) {
			t.Errorf("fish output missing %q\n\n%s", check, got)
		}
	}
}

func TestCompletionFormatter_NoName(t *testing.T) {
	doc := &shedoc.Document{}

	formatters := []struct {
		name string
		f    shedoc.Formatter
	}{
		{"bash", &BashCompletionFormatter{}},
		{"zsh", &ZshCompletionFormatter{}},
		{"fish", &FishCompletionFormatter{}},
	}

	for _, ff := range formatters {
		t.Run(ff.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ff.f.Format(&buf, doc)
			if err == nil {
				t.Error("expected error for missing name")
			}
		})
	}
}
