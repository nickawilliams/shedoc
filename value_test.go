package shedoc

import (
	"testing"
)

func TestParseValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Value
		wantErr bool
	}{
		{
			name:  "required",
			input: "<name>",
			want:  Value{Name: "name", Required: true},
		},
		{
			name:  "optional",
			input: "[name]",
			want:  Value{Name: "name", Required: false},
		},
		{
			name:  "optional with default",
			input: "[name=default]",
			want:  Value{Name: "name", Required: false, Default: "default"},
		},
		{
			name:  "required variadic",
			input: "<name...>",
			want:  Value{Name: "name", Required: true, Variadic: true},
		},
		{
			name:  "optional variadic",
			input: "[name...]",
			want:  Value{Name: "name", Required: false, Variadic: true},
		},
		{
			name:  "hyphenated name",
			input: "<file-path>",
			want:  Value{Name: "file-path", Required: true},
		},
		{
			name:  "default with dots",
			input: "[version=1.0.0]",
			want:  Value{Name: "version", Required: false, Default: "1.0.0"},
		},
		{
			name:  "whitespace trimmed",
			input: "  <name>  ",
			want:  Value{Name: "name", Required: true},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no brackets",
			input:   "name",
			wantErr: true,
		},
		{
			name:    "mismatched brackets open angle close square",
			input:   "<name]",
			wantErr: true,
		},
		{
			name:    "mismatched brackets open square close angle",
			input:   "[name>",
			wantErr: true,
		},
		{
			name:    "empty brackets",
			input:   "<>",
			wantErr: true,
		},
		{
			name:    "default on required",
			input:   "<name=foo>",
			wantErr: true,
		},
		{
			name:    "only ellipsis",
			input:   "<...>",
			wantErr: true,
		},
		{
			name:    "empty name before equals",
			input:   "[=default]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseValue(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseValue(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseValue(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseValue(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
