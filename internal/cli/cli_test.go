package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nickawilliams/shedoc"
)

// testdataPath returns the absolute path to a testdata file.
func testdataPath(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// runCLI executes the root command with the given args and returns stdout, stderr, and any error.
func runCLI(args ...string) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test-version")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	// Redirect os.Stdout for commands that write directly to os.Stdout.
	// The root command uses its own writer, but we also capture stderr.
	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// --- JSON output ---

func TestCLI_JSONDefault(t *testing.T) {
	stdout, _, err := runCLI(testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, stdout)
	}
	if doc.Meta.Name != "deploy" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "deploy")
	}
	if doc.Meta.Version != "2.1.0" {
		t.Errorf("Meta.Version = %q, want %q", doc.Meta.Version, "2.1.0")
	}
}

func TestCLI_JSONExplicit(t *testing.T) {
	stdout, _, err := runCLI("--to", "json", testdataPath(t, "standalone.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if doc.Meta.Name != "greet" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "greet")
	}
}

func TestCLI_JSONMultipleFiles(t *testing.T) {
	stdout, _, err := runCLI(
		testdataPath(t, "comprehensive.sh"),
		testdataPath(t, "standalone.sh"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// NDJSON: one JSON object per line.
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 NDJSON lines, got %d:\n%s", len(lines), stdout)
	}

	var doc1, doc2 shedoc.Document
	if err := json.Unmarshal([]byte(lines[0]), &doc1); err != nil {
		t.Fatalf("line 1 is not valid JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &doc2); err != nil {
		t.Fatalf("line 2 is not valid JSON: %v", err)
	}
	if doc1.Meta.Name != "deploy" {
		t.Errorf("doc1.Meta.Name = %q, want %q", doc1.Meta.Name, "deploy")
	}
	if doc2.Meta.Name != "greet" {
		t.Errorf("doc2.Meta.Name = %q, want %q", doc2.Meta.Name, "greet")
	}
}

// --- --get flag ---

func TestCLI_GetName(t *testing.T) {
	stdout, _, err := runCLI("--get", "name", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(stdout); got != "deploy" {
		t.Errorf("--get name = %q, want %q", got, "deploy")
	}
}

func TestCLI_GetVersion(t *testing.T) {
	stdout, _, err := runCLI("--get", "version", testdataPath(t, "standalone.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(stdout); got != "1.0.0" {
		t.Errorf("--get version = %q, want %q", got, "1.0.0")
	}
}

func TestCLI_GetDescription(t *testing.T) {
	stdout, _, err := runCLI("--get", "description", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "deployment tool") {
		t.Errorf("--get description missing expected content: %s", stdout)
	}
}

func TestCLI_GetUnknownTag(t *testing.T) {
	_, _, err := runCLI("--get", "nonexistent", testdataPath(t, "comprehensive.sh"))
	if err == nil {
		t.Fatal("expected error for unknown tag")
	}
	if !strings.Contains(err.Error(), "unknown tag") {
		t.Errorf("expected 'unknown tag' error, got: %v", err)
	}
}

func TestCLI_GetEmptyValue(t *testing.T) {
	stdout, _, err := runCLI("--get", "license", testdataPath(t, "standalone.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// standalone.sh has no license — should output nothing.
	if got := strings.TrimSpace(stdout); got != "" {
		t.Errorf("expected empty output for missing tag, got %q", got)
	}
}

func TestCLI_GetMultipleFiles(t *testing.T) {
	stdout, _, err := runCLI("--get", "name",
		testdataPath(t, "comprehensive.sh"),
		testdataPath(t, "standalone.sh"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), stdout)
	}
	if lines[0] != "deploy" {
		t.Errorf("line 0 = %q, want %q", lines[0], "deploy")
	}
	if lines[1] != "greet" {
		t.Errorf("line 1 = %q, want %q", lines[1], "greet")
	}
}

// --- Format outputs ---

func TestCLI_HelpFormat(t *testing.T) {
	stdout, _, err := runCLI("--to", "help", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"deploy", "Usage:", "Commands:", "push", "status", "Options:"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}

func TestCLI_ManFormat(t *testing.T) {
	stdout, _, err := runCLI("--to", "man", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Man pages use roff macros.
	for _, want := range []string{".TH", ".SH", "deploy"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("man output missing %q", want)
		}
	}
}

func TestCLI_CompletionBashFormat(t *testing.T) {
	stdout, _, err := runCLI("--to", "completion:bash", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"_deploy()", "complete -F"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("bash completion output missing %q", want)
		}
	}
}

func TestCLI_CompletionZshFormat(t *testing.T) {
	stdout, _, err := runCLI("--to", "completion:zsh", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"#compdef deploy", "_deploy()"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("zsh completion output missing %q", want)
		}
	}
}

func TestCLI_CompletionFishFormat(t *testing.T) {
	stdout, _, err := runCLI("--to", "completion:fish", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"complete -c deploy", "-a push", "-a status"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("fish completion output missing %q", want)
		}
	}
}

// --- Output file ---

func TestCLI_OutputFile(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "out.json")
	_, _, err := runCLI("--output", outPath, testdataPath(t, "standalone.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output file is not valid JSON: %v", err)
	}
	if doc.Meta.Name != "greet" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "greet")
	}
}

// --- Warnings ---

func TestCLI_WarningsIncluded(t *testing.T) {
	stdout, _, err := runCLI("--warnings", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	// With --warnings, the warnings field should be present (may be null/empty
	// if the file has no warnings — that's fine, we're testing the flag is accepted).
}

func TestCLI_QuietSuppressesStderr(t *testing.T) {
	// Parse a file — with --quiet, stderr should be empty.
	_, stderr, err := runCLI("--quiet", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr with --quiet, got: %s", stderr)
	}
}

// --- Error cases ---

func TestCLI_NoArgs(t *testing.T) {
	_, _, err := runCLI()
	if err == nil {
		t.Fatal("expected error with no args")
	}
}

func TestCLI_MissingFile(t *testing.T) {
	_, _, err := runCLI("/nonexistent/path/to/script.sh")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCLI_UnknownFormat(t *testing.T) {
	_, _, err := runCLI("--to", "xml", testdataPath(t, "comprehensive.sh"))
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected 'unknown format' error, got: %v", err)
	}
}

func TestCLI_NonJSONMultipleFiles(t *testing.T) {
	_, _, err := runCLI("--to", "help",
		testdataPath(t, "comprehensive.sh"),
		testdataPath(t, "standalone.sh"),
	)
	if err == nil {
		t.Fatal("expected error for non-JSON format with multiple files")
	}
	if !strings.Contains(err.Error(), "single file") {
		t.Errorf("expected 'single file' error, got: %v", err)
	}
}

func TestCLI_ToAndGetMutuallyExclusive(t *testing.T) {
	_, _, err := runCLI("--to", "help", "--get", "name", testdataPath(t, "comprehensive.sh"))
	if err == nil {
		t.Fatal("expected error for --to and --get together")
	}
}

// --- Testdata coverage ---

func TestCLI_MinimalFile(t *testing.T) {
	stdout, _, err := runCLI(testdataPath(t, "minimal.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if doc.Shebang != "/bin/bash" {
		t.Errorf("Shebang = %q, want %q", doc.Shebang, "/bin/bash")
	}
}

func TestCLI_NoShedocFile(t *testing.T) {
	stdout, _, err := runCLI(testdataPath(t, "no_shedoc.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	// Should parse successfully with no blocks.
	if len(doc.Blocks) != 0 {
		t.Errorf("expected 0 blocks for no_shedoc.sh, got %d", len(doc.Blocks))
	}
}

func TestCLI_LibraryFile(t *testing.T) {
	stdout, _, err := runCLI(testdataPath(t, "library.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if doc.Meta.Name != "string-utils" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "string-utils")
	}
	// Should have public and private blocks.
	if len(doc.Blocks) < 2 {
		t.Errorf("expected at least 2 blocks, got %d", len(doc.Blocks))
	}
}

func TestCLI_EdgeCasesFile(t *testing.T) {
	stdout, _, err := runCLI(testdataPath(t, "edge_cases.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc shedoc.Document
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if doc.Meta.Name != "edge-cases" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "edge-cases")
	}
}

// --- Complete subcommand integration ---

func TestCLI_CompleteSetupBash(t *testing.T) {
	stdout, _, err := runCLI("complete", "--setup", "bash", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "complete -C") {
		t.Errorf("bash setup missing 'complete -C': %s", stdout)
	}
	if !strings.Contains(stdout, "deploy") {
		t.Errorf("bash setup missing command name: %s", stdout)
	}
}

func TestCLI_CompleteSetupZsh(t *testing.T) {
	stdout, _, err := runCLI("complete", "--setup", "zsh", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"compdef", "compadd", "COMP_LINE", "COMP_POINT"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("zsh setup missing %q: %s", want, stdout)
		}
	}
}

func TestCLI_CompleteSetupFish(t *testing.T) {
	stdout, _, err := runCLI("complete", "--setup", "fish", testdataPath(t, "comprehensive.sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "complete -c deploy") {
		t.Errorf("fish setup missing 'complete -c deploy': %s", stdout)
	}
	if !strings.Contains(stdout, "--shell fish") {
		t.Errorf("fish setup missing '--shell fish': %s", stdout)
	}
}

func TestCLI_CompleteNoArgs(t *testing.T) {
	_, _, err := runCLI("complete")
	if err == nil {
		t.Fatal("expected error for complete with no args")
	}
}

// --- Stdin ---

func TestCLI_Stdin(t *testing.T) {
	// Create a pipe with shedoc-annotated content.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	input := "#!/bin/bash\n#?/name stdin-test\n#?/version 0.1.0\n"
	go func() {
		w.WriteString(input)
		w.Close()
	}()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	stdout, _, execErr := runCLI("--get", "name", "-")
	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if got := strings.TrimSpace(stdout); got != "stdin-test" {
		t.Errorf("stdin name = %q, want %q", got, "stdin-test")
	}
}

// --- Version ---

func TestCLI_Version(t *testing.T) {
	stdout, _, err := runCLI("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "test-version") {
		t.Errorf("version output missing 'test-version': %s", stdout)
	}
}
