package shedoc

import (
	"testing"
)

func TestParseFlag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Flag
		wantErr bool
	}{
		{
			name:  "short only",
			input: "-v",
			want:  Flag{Short: "-v", Line: 1},
		},
		{
			name:  "long only",
			input: "--verbose",
			want:  Flag{Long: "--verbose", Line: 1},
		},
		{
			name:  "short and long",
			input: "-v | --verbose",
			want:  Flag{Short: "-v", Long: "--verbose", Line: 1},
		},
		{
			name:  "with description",
			input: "-v | --verbose Enable verbose output",
			want:  Flag{Short: "-v", Long: "--verbose", Description: "Enable verbose output", Line: 1},
		},
		{
			name:  "short with description",
			input: "-v Enable verbose output",
			want:  Flag{Short: "-v", Description: "Enable verbose output", Line: 1},
		},
		{
			name:  "long with description",
			input: "--verbose Enable verbose output",
			want:  Flag{Long: "--verbose", Description: "Enable verbose output", Line: 1},
		},
		{
			name:  "long only with hyphenated name",
			input: "--dry-run Preview changes without deploying",
			want:  Flag{Long: "--dry-run", Description: "Preview changes without deploying", Line: 1},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlag(tt.input, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseFlag(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFlag(%q) unexpected error: %v", tt.input, err)
			}
			if *got != tt.want {
				t.Errorf("parseFlag(%q) = %+v, want %+v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseOption(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Option
		wantErr bool
	}{
		{
			name:  "short and long with required value",
			input: "-f | --format <type> Output format",
			want: Option{
				Short:       "-f",
				Long:        "--format",
				Value:       Value{Name: "type", Required: true},
				Description: "Output format",
				Line:        1,
			},
		},
		{
			name:  "long with optional default value",
			input: "--format [type=json] Output format",
			want: Option{
				Long:        "--format",
				Value:       Value{Name: "type", Required: false, Default: "json"},
				Description: "Output format",
				Line:        1,
			},
		},
		{
			name:  "long with optional value no default",
			input: "--tag [version] Version tag",
			want: Option{
				Long:        "--tag",
				Value:       Value{Name: "version", Required: false},
				Description: "Version tag",
				Line:        1,
			},
		},
		{
			name:  "short and long no description",
			input: "-c | --config <path>",
			want: Option{
				Short: "-c",
				Long:  "--config",
				Value: Value{Name: "path", Required: true},
				Line:  1,
			},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no value notation",
			input:   "--format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOption(tt.input, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseOption(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOption(%q) unexpected error: %v", tt.input, err)
			}
			if *got != tt.want {
				t.Errorf("parseOption(%q) = %+v, want %+v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseOperand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Operand
		wantErr bool
	}{
		{
			name:  "required",
			input: "<environment> Target environment",
			want: Operand{
				Value:       Value{Name: "environment", Required: true},
				Description: "Target environment",
				Line:        1,
			},
		},
		{
			name:  "optional",
			input: "[name] Name to greet",
			want: Operand{
				Value:       Value{Name: "name", Required: false},
				Description: "Name to greet",
				Line:        1,
			},
		},
		{
			name:  "optional with default",
			input: "[name=World] Name to greet",
			want: Operand{
				Value:       Value{Name: "name", Required: false, Default: "World"},
				Description: "Name to greet",
				Line:        1,
			},
		},
		{
			name:  "variadic",
			input: "[services...] Specific services to deploy",
			want: Operand{
				Value:       Value{Name: "services", Required: false, Variadic: true},
				Description: "Specific services to deploy",
				Line:        1,
			},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOperand(tt.input, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseOperand(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOperand(%q) unexpected error: %v", tt.input, err)
			}
			if *got != tt.want {
				t.Errorf("parseOperand(%q) = %+v, want %+v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Env
		wantErr bool
	}{
		{
			name:  "with description",
			input: "DEPLOY_TOKEN Authentication token",
			want:  Env{Name: "DEPLOY_TOKEN", Description: "Authentication token", Line: 1},
		},
		{
			name:  "no description",
			input: "HOME",
			want:  Env{Name: "HOME", Line: 1},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnv(tt.input, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseEnv(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseEnv(%q) unexpected error: %v", tt.input, err)
			}
			if *got != tt.want {
				t.Errorf("parseEnv(%q) = %+v, want %+v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseExit(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Exit
		wantErr bool
	}{
		{
			name:  "with description",
			input: "0 Success",
			want:  Exit{Code: "0", Description: "Success", Line: 1},
		},
		{
			name:  "no description",
			input: "1",
			want:  Exit{Code: "1", Line: 1},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExit(tt.input, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseExit(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseExit(%q) unexpected error: %v", tt.input, err)
			}
			if *got != tt.want {
				t.Errorf("parseExit(%q) = %+v, want %+v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestParseReads(t *testing.T) {
	got, err := parseReads("~/.deployrc User configuration", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := &Reads{Path: "~/.deployrc", Description: "User configuration", Line: 5}
	if *got != *want {
		t.Errorf("got %+v, want %+v", *got, *want)
	}
}

func TestParseReadsNoDescription(t *testing.T) {
	got, err := parseReads("~/.config", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "~/.config" || got.Description != "" {
		t.Errorf("got %+v", *got)
	}
}

func TestParseReadsEmpty(t *testing.T) {
	_, err := parseReads("", 1)
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseWrites(t *testing.T) {
	got, err := parseWrites("/var/log/deploy.log Deployment log", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := &Writes{Path: "/var/log/deploy.log", Description: "Deployment log", Line: 10}
	if *got != *want {
		t.Errorf("got %+v, want %+v", *got, *want)
	}
}

func TestParseSets(t *testing.T) {
	got, err := parseSets("DEPLOY_LAST_ROLLBACK Timestamp of last rollback", 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := &Sets{Name: "DEPLOY_LAST_ROLLBACK", Description: "Timestamp of last rollback", Line: 15}
	if *got != *want {
		t.Errorf("got %+v, want %+v", *got, *want)
	}
}

func TestParseSetsNoDescription(t *testing.T) {
	got, err := parseSets("MY_VAR", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "MY_VAR" || got.Description != "" {
		t.Errorf("got %+v", *got)
	}
}

func TestParseSetsEmpty(t *testing.T) {
	_, err := parseSets("", 1)
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseWritesNoDescription(t *testing.T) {
	got, err := parseWrites("/tmp/out", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "/tmp/out" || got.Description != "" {
		t.Errorf("got %+v", *got)
	}
}

func TestParseWritesEmpty(t *testing.T) {
	_, err := parseWrites("", 1)
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseOperandNoDescription(t *testing.T) {
	got, err := parseOperand("<file>", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value.Name != "file" || got.Description != "" {
		t.Errorf("got %+v", *got)
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		text     string
		wantName string
		wantErr  bool
	}{
		{"stdin", "stdin", "Reads version from STDIN", "stdin", false},
		{"stdout", "stdout", "Deployment progress", "stdout", false},
		{"stderr", "stderr", "Error messages", "stderr", false},
		{"deprecated", "deprecated", "Use 'deploy push --migrate' instead.", "deprecated", false},
		{"deprecated empty", "deprecated", "", "deprecated", false},
		{"unknown", "foobar", "something", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _, err := parseTag(tt.tagName, tt.text, 1)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTag(%q, %q) expected error", tt.tagName, tt.text)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTag(%q, %q) unexpected error: %v", tt.tagName, tt.text, err)
			}
			if name != tt.wantName {
				t.Errorf("parseTag(%q, %q) name = %q, want %q", tt.tagName, tt.text, name, tt.wantName)
			}
		})
	}
}
