package generate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nickawilliams/shedoc"
)

func TestManPageFormatter_Comprehensive(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{
			Name:        "deploy",
			Version:     "2.1.0",
			Synopsis:    "deploy [-v] [-c config] <command> [args...]",
			Description: "A deployment tool for managing application releases.",
			Section:     "1",
			Author:      "Jane Developer",
			License:     "MIT",
			Examples:    "deploy status production\ndeploy push --force staging",
		},
		Blocks: []shedoc.Block{
			{
				Visibility:  shedoc.VisibilityCommand,
				Description: "Manages application deployments.",
				Flags: []shedoc.Flag{
					{Short: "-v", Long: "--verbose", Description: "Enable verbose output"},
				},
				Options: []shedoc.Option{
					{Short: "-c", Long: "--config", Value: shedoc.Value{Name: "path", Required: true}, Description: "Config file"},
				},
				Env: []shedoc.Env{
					{Name: "DEPLOY_TOKEN", Description: "Authentication token"},
				},
				Reads: []shedoc.Reads{
					{Path: "~/.deployrc", Description: "User configuration"},
				},
				Exit: []shedoc.Exit{
					{Code: "0", Description: "Success"},
					{Code: "1", Description: "General error"},
				},
			},
			{
				Visibility:  shedoc.VisibilitySubcommand,
				Name:        "push",
				Description: "Deploys the application.",
				Flags: []shedoc.Flag{
					{Short: "-f", Long: "--force", Description: "Skip confirmation"},
				},
			},
			{
				Visibility: shedoc.VisibilitySubcommand,
				Name:       "migrate",
				Deprecated: &shedoc.Deprecated{Message: "Use 'deploy push --migrate' instead."},
			},
		},
	}

	var buf bytes.Buffer
	f := &ManPageFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	checks := []struct {
		label string
		text  string
	}{
		{"TH header", ".TH DEPLOY 1"},
		{"NAME section", ".SH NAME"},
		{"name with brief", "deploy \\- A deployment tool"},
		{"SYNOPSIS section", ".SH SYNOPSIS"},
		{"DESCRIPTION section", ".SH DESCRIPTION"},
		{"OPTIONS section", ".SH OPTIONS"},
		{"verbose flag", "\\-v, \\-\\-verbose"},
		{"config option", "\\-c, \\-\\-config"},
		{"COMMANDS section", ".SH COMMANDS"},
		{"push subcommand", ".B push"},
		{"migrate deprecated", "[deprecated]"},
		{"ENVIRONMENT section", ".SH ENVIRONMENT"},
		{"DEPLOY_TOKEN", "DEPLOY_TOKEN"},
		{"FILES section", ".SH FILES"},
		{"deployrc", ".deployrc"},
		{"EXIT STATUS section", ".SH EXIT STATUS"},
		{"EXAMPLES section", ".SH EXAMPLES"},
		{"AUTHOR section", ".SH AUTHOR"},
		{"author name", "Jane Developer"},
	}

	for _, check := range checks {
		if !strings.Contains(got, check.text) {
			t.Errorf("[%s] output missing %q\n\nfull output:\n%s", check.label, check.text, got)
		}
	}
}

func TestManPageFormatter_Minimal(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{
			Name: "greet",
		},
	}

	var buf bytes.Buffer
	f := &ManPageFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, ".TH GREET 1") {
		t.Errorf("output missing .TH header\n%s", got)
	}
}

func TestManPageFormatter_DeprecatedEmptyMessage(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{Name: "tool"},
		Blocks: []shedoc.Block{
			{Visibility: shedoc.VisibilityCommand},
			{
				Visibility: shedoc.VisibilitySubcommand,
				Name:       "old",
				Deprecated: &shedoc.Deprecated{Message: ""},
			},
		},
	}

	var buf bytes.Buffer
	f := &ManPageFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "This command is deprecated.") {
		t.Errorf("missing default deprecated message\n%s", got)
	}
}

func TestTroffEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"--verbose", "\\-\\-verbose"},
		{"-v", "\\-v"},
		{"plain text", "plain text"},
		{"back\\slash", "back\\\\slash"},
	}

	for _, tt := range tests {
		got := troffEscape(tt.input)
		if got != tt.want {
			t.Errorf("troffEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
