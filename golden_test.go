package shedoc

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestGoldenFiles(t *testing.T) {
	files, err := filepath.Glob("testdata/*.sh")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no testdata/*.sh files found")
	}

	for _, shFile := range files {
		name := strings.TrimSuffix(filepath.Base(shFile), ".sh")
		goldenFile := strings.TrimSuffix(shFile, ".sh") + ".json"

		t.Run(name, func(t *testing.T) {
			doc, err := Parse(shFile)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", shFile, err)
			}

			// Clear the path for deterministic output.
			doc.Path = ""

			got, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				t.Fatalf("json.Marshal error: %v", err)
			}
			got = append(got, '\n')

			if *update {
				if err := os.WriteFile(goldenFile, got, 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				t.Logf("updated %s", goldenFile)
				return
			}

			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s (run with -update to create): %v", goldenFile, err)
			}

			if string(got) != string(want) {
				t.Errorf("output mismatch for %s\ngot:\n%s\nwant:\n%s", shFile, got, want)
			}
		})
	}
}
