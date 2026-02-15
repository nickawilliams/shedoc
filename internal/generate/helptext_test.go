package generate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nickawilliams/shedoc"
)

func TestHelpTextFormatter_Comprehensive(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{
			Name:        "deploy",
			Version:     "2.1.0",
			Synopsis:    "deploy [-v] [-c config] <command> [args...]",
			Description: "A deployment tool for managing application releases.\nSupports multiple environments.",
		},
		Blocks: []shedoc.Block{
			{
				Visibility:  shedoc.VisibilityCommand,
				Description: "Manages application deployments.",
				Flags: []shedoc.Flag{
					{Short: "-v", Long: "--verbose", Description: "Enable verbose output"},
				},
				Options: []shedoc.Option{
					{Short: "-c", Long: "--config", Value: shedoc.Value{Name: "path", Required: true}, Description: "Path to configuration file"},
				},
				Env: []shedoc.Env{
					{Name: "DEPLOY_TOKEN", Description: "Authentication token"},
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
			},
			{
				Visibility:  shedoc.VisibilitySubcommand,
				Name:        "status",
				Description: "Shows deployment status.",
			},
			{
				Visibility: shedoc.VisibilitySubcommand,
				Name:       "migrate",
				Deprecated: &shedoc.Deprecated{Message: "Use 'deploy push --migrate' instead."},
			},
		},
	}

	var buf bytes.Buffer
	f := &HelpTextFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()

	// Verify key sections are present.
	checks := []string{
		"deploy - A deployment tool for managing application releases.",
		"Usage:",
		"deploy [-v] [-c config] <command> [args...]",
		"Commands:",
		"push",
		"status",
		"migrate",
		"[deprecated]",
		"Options:",
		"-v, --verbose",
		"-c, --config <path>",
		"Environment:",
		"DEPLOY_TOKEN",
		"Exit Codes:",
		"0",
		"1",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("output missing %q\n\nfull output:\n%s", check, got)
		}
	}
}

func TestHelpTextFormatter_Minimal(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{
			Name: "greet",
		},
	}

	var buf bytes.Buffer
	f := &HelpTextFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "greet") {
		t.Errorf("output missing name\n%s", got)
	}
}

func TestHelpTextFormatter_LongOnlyFlag(t *testing.T) {
	doc := &shedoc.Document{
		Blocks: []shedoc.Block{
			{
				Visibility: shedoc.VisibilityCommand,
				Flags: []shedoc.Flag{
					{Long: "--dry-run", Description: "Preview changes"},
				},
			},
		},
	}

	var buf bytes.Buffer
	f := &HelpTextFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "    --dry-run") {
		t.Errorf("long-only flag not indented correctly\n%s", got)
	}
}

func TestHelpTextFormatter_ShortOnlyFlagAndOption(t *testing.T) {
	doc := &shedoc.Document{
		Blocks: []shedoc.Block{
			{
				Visibility: shedoc.VisibilityCommand,
				Flags: []shedoc.Flag{
					{Short: "-v", Description: "Verbose"},
				},
				Options: []shedoc.Option{
					{Short: "-o", Value: shedoc.Value{Name: "file", Required: true}, Description: "Output"},
					{Long: "--format", Value: shedoc.Value{Name: "fmt", Required: true}, Description: "Format"},
				},
			},
		},
	}

	var buf bytes.Buffer
	f := &HelpTextFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "-v") {
		t.Errorf("missing short-only flag\n%s", got)
	}
	if !strings.Contains(got, "-o <file>") {
		t.Errorf("missing short-only option\n%s", got)
	}
	if !strings.Contains(got, "    --format <fmt>") {
		t.Errorf("missing long-only option\n%s", got)
	}
}

func TestHelpTextFormatter_NoDescription(t *testing.T) {
	doc := &shedoc.Document{
		Blocks: []shedoc.Block{
			{
				Visibility: shedoc.VisibilityCommand,
				Flags: []shedoc.Flag{
					{Short: "-v"},
				},
				Options: []shedoc.Option{
					{Long: "--format", Value: shedoc.Value{Name: "fmt", Required: true}},
				},
				Env: []shedoc.Env{
					{Name: "MY_VAR"},
				},
				Exit: []shedoc.Exit{
					{Code: "0"},
				},
			},
			{
				Visibility: shedoc.VisibilitySubcommand,
				Name:       "sub",
			},
		},
	}

	var buf bytes.Buffer
	f := &HelpTextFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "-v\n") {
		t.Errorf("flag without description not rendered\n%s", got)
	}
	if !strings.Contains(got, "sub\n") {
		t.Errorf("subcommand without description not rendered\n%s", got)
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name string
		val  shedoc.Value
		want string
	}{
		{"required", shedoc.Value{Name: "path", Required: true}, "<path>"},
		{"optional", shedoc.Value{Name: "name", Required: false}, "[name]"},
		{"optional with default", shedoc.Value{Name: "fmt", Required: false, Default: "text"}, "[fmt=text]"},
		{"required variadic", shedoc.Value{Name: "files", Required: true, Variadic: true}, "<files...>"},
		{"optional variadic", shedoc.Value{Name: "args", Required: false, Variadic: true}, "[args...]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.val)
			if got != tt.want {
				t.Errorf("formatValue(%+v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}
