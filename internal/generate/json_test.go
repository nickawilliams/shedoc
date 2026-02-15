package generate

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nickawilliams/shedoc"
)

func TestJSONFormatter(t *testing.T) {
	doc := &shedoc.Document{
		Meta: shedoc.Meta{
			Name:    "test-script",
			Version: "1.0.0",
		},
		Blocks: []shedoc.Block{
			{
				Visibility:  shedoc.VisibilityCommand,
				Description: "A test command.",
				Flags: []shedoc.Flag{
					{Short: "-v", Long: "--verbose", Description: "Verbose"},
				},
			},
		},
	}

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, doc); err != nil {
		t.Fatal(err)
	}

	// Verify output is valid JSON.
	var roundtrip shedoc.Document
	if err := json.Unmarshal(buf.Bytes(), &roundtrip); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	if roundtrip.Meta.Name != "test-script" {
		t.Errorf("Meta.Name = %q, want %q", roundtrip.Meta.Name, "test-script")
	}
	if len(roundtrip.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(roundtrip.Blocks))
	}
}
