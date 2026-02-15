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
