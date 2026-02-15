package shedoc

import (
	"strings"
	"testing"
)

func TestParseShebang(t *testing.T) {
	doc := mustParse(t, "#!/bin/bash\n")
	if doc.Shebang != "/bin/bash" {
		t.Errorf("Shebang = %q, want %q", doc.Shebang, "/bin/bash")
	}
}

func TestParseShebangEnv(t *testing.T) {
	doc := mustParse(t, "#!/usr/bin/env bash\n")
	if doc.Shebang != "/usr/bin/env bash" {
		t.Errorf("Shebang = %q, want %q", doc.Shebang, "/usr/bin/env bash")
	}
}

func TestParseShedocInline(t *testing.T) {
	input := `#!/bin/bash
#?/name my-script
#?/version 1.0.0
#?/section 1
#?/author Jane Developer
#?/license MIT
`
	doc := mustParse(t, input)
	if doc.Meta.Name != "my-script" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "my-script")
	}
	if doc.Meta.Version != "1.0.0" {
		t.Errorf("Meta.Version = %q, want %q", doc.Meta.Version, "1.0.0")
	}
	if doc.Meta.Section != "1" {
		t.Errorf("Meta.Section = %q, want %q", doc.Meta.Section, "1")
	}
	if doc.Meta.Author != "Jane Developer" {
		t.Errorf("Meta.Author = %q, want %q", doc.Meta.Author, "Jane Developer")
	}
	if doc.Meta.License != "MIT" {
		t.Errorf("Meta.License = %q, want %q", doc.Meta.License, "MIT")
	}
}

func TestParseShedocInlineSynopsis(t *testing.T) {
	input := `#!/bin/bash
#?/synopsis deploy [-v] [-c config] <command> [args...]
`
	doc := mustParse(t, input)
	if doc.Meta.Synopsis != "deploy [-v] [-c config] <command> [args...]" {
		t.Errorf("Meta.Synopsis = %q, want %q", doc.Meta.Synopsis, "deploy [-v] [-c config] <command> [args...]")
	}
}

func TestParseShedocBlock(t *testing.T) {
	input := `#!/bin/bash
#?/description
 # A deployment tool for managing
 # application releases.
 ##
`
	doc := mustParse(t, input)
	want := "A deployment tool for managing\napplication releases."
	if doc.Meta.Description != want {
		t.Errorf("Meta.Description = %q, want %q", doc.Meta.Description, want)
	}
}

func TestParseSheblockCommand(t *testing.T) {
	input := `#!/bin/bash
#@/command
 # Manages deployments.
 #
 # @flag -v | --verbose Enable verbose output
 # @option -c | --config <path> Config file
 # @operand <command> Subcommand to run
 # @env DEPLOY_TOKEN Auth token
 # @reads ~/.deployrc User config
 # @exit 0 Success
 # @exit 1 General error
 # @stderr Error messages
 ##
main() {
    echo "hello"
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}

	b := doc.Blocks[0]
	if b.Visibility != VisibilityCommand {
		t.Errorf("Visibility = %q, want %q", b.Visibility, VisibilityCommand)
	}
	if b.Description != "Manages deployments." {
		t.Errorf("Description = %q, want %q", b.Description, "Manages deployments.")
	}
	if b.FunctionName != "main" {
		t.Errorf("FunctionName = %q, want %q", b.FunctionName, "main")
	}
	if len(b.Flags) != 1 {
		t.Fatalf("got %d flags, want 1", len(b.Flags))
	}
	if b.Flags[0].Short != "-v" || b.Flags[0].Long != "--verbose" {
		t.Errorf("Flag = %+v", b.Flags[0])
	}
	if len(b.Options) != 1 {
		t.Fatalf("got %d options, want 1", len(b.Options))
	}
	if b.Options[0].Long != "--config" {
		t.Errorf("Option.Long = %q", b.Options[0].Long)
	}
	if len(b.Operands) != 1 {
		t.Fatalf("got %d operands, want 1", len(b.Operands))
	}
	if len(b.Env) != 1 {
		t.Fatalf("got %d env, want 1", len(b.Env))
	}
	if len(b.Reads) != 1 {
		t.Fatalf("got %d reads, want 1", len(b.Reads))
	}
	if len(b.Exit) != 2 {
		t.Fatalf("got %d exit, want 2", len(b.Exit))
	}
	if b.Stderr == nil {
		t.Fatal("Stderr is nil")
	}
}

func TestParseSheblockSubcommand(t *testing.T) {
	input := `#!/bin/bash
#@/subcommand push
 # Deploys the application.
 #
 # @flag -f | --force Skip confirmation
 # @operand <environment> Target env
 ##
cmd_push() {
    echo "pushing"
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	b := doc.Blocks[0]
	if b.Visibility != VisibilitySubcommand {
		t.Errorf("Visibility = %q, want %q", b.Visibility, VisibilitySubcommand)
	}
	if b.Name != "push" {
		t.Errorf("Name = %q, want %q", b.Name, "push")
	}
	if b.FunctionName != "cmd_push" {
		t.Errorf("FunctionName = %q, want %q", b.FunctionName, "cmd_push")
	}
}

func TestParseSheblockPublic(t *testing.T) {
	input := `#!/bin/bash
#@/public
 # Converts a string to uppercase.
 #
 # @operand <string> The string to convert
 # @stdout Uppercase result
 ##
to_upper() {
    echo "${1^^}"
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	b := doc.Blocks[0]
	if b.Visibility != VisibilityPublic {
		t.Errorf("Visibility = %q, want %q", b.Visibility, VisibilityPublic)
	}
	if b.FunctionName != "to_upper" {
		t.Errorf("FunctionName = %q, want %q", b.FunctionName, "to_upper")
	}
	if b.Stdout == nil {
		t.Fatal("Stdout is nil")
	}
}

func TestParseSheblockPrivate(t *testing.T) {
	input := `#!/bin/bash
#@/private
 # Internal helper.
 ##
_validate_input() {
    [[ -n "$1" ]]
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	b := doc.Blocks[0]
	if b.Visibility != VisibilityPrivate {
		t.Errorf("Visibility = %q, want %q", b.Visibility, VisibilityPrivate)
	}
	if b.FunctionName != "_validate_input" {
		t.Errorf("FunctionName = %q, want %q", b.FunctionName, "_validate_input")
	}
}

func TestParseSheblockBare(t *testing.T) {
	input := `#!/bin/bash
#@/
 # Bare block defaults to public.
 ##
my_func() {
    echo "hello"
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	if doc.Blocks[0].Visibility != VisibilityPublic {
		t.Errorf("Visibility = %q, want %q", doc.Blocks[0].Visibility, VisibilityPublic)
	}
}

func TestParseStandaloneCommand(t *testing.T) {
	input := `#!/usr/bin/env bash
#@/command
 # Prints a greeting.
 #
 # @operand [name=World] Name to greet
 # @exit 0 Success
 # @stdout Greeting message
 ##

echo "Hello, ${1:-World}!"
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	b := doc.Blocks[0]
	if b.FunctionName != "" {
		t.Errorf("FunctionName = %q, want empty (standalone command)", b.FunctionName)
	}
	if b.Operands[0].Value.Default != "World" {
		t.Errorf("Operand default = %q, want %q", b.Operands[0].Value.Default, "World")
	}
}

func TestParseTagContinuation(t *testing.T) {
	input := `#!/bin/bash
#@/command
 # Does things.
 #
 # @env DEPLOY_TOKEN Authentication token for the deployment
 #                    service. Can also be provided via the
 #                    .deployrc configuration file.
 ##
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	env := doc.Blocks[0].Env[0]
	want := "Authentication token for the deployment service. Can also be provided via the .deployrc configuration file."
	if env.Description != want {
		t.Errorf("Env.Description = %q, want %q", env.Description, want)
	}
}

func TestParseDeprecated(t *testing.T) {
	input := `#!/bin/bash
#@/subcommand migrate
 # @deprecated Use 'deploy push --migrate' instead.
 ##
cmd_migrate() {
    echo "migrating"
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	b := doc.Blocks[0]
	if b.Deprecated == nil {
		t.Fatal("Deprecated is nil")
	}
	if b.Deprecated.Message != "Use 'deploy push --migrate' instead." {
		t.Errorf("Deprecated.Message = %q", b.Deprecated.Message)
	}
}

func TestParseNoShedoc(t *testing.T) {
	input := `#!/bin/bash
echo "hello world"
`
	doc := mustParse(t, input)
	if doc.Shebang != "/bin/bash" {
		t.Errorf("Shebang = %q, want %q", doc.Shebang, "/bin/bash")
	}
	if len(doc.Blocks) != 0 {
		t.Errorf("got %d blocks, want 0", len(doc.Blocks))
	}
}

func TestParseNoShebang(t *testing.T) {
	input := `#?/name string-utils
#?/version 1.0.0
`
	doc := mustParse(t, input)
	if doc.Shebang != "" {
		t.Errorf("Shebang = %q, want empty", doc.Shebang)
	}
	if doc.Meta.Name != "string-utils" {
		t.Errorf("Meta.Name = %q, want %q", doc.Meta.Name, "string-utils")
	}
}

func TestParseMultipleBlocks(t *testing.T) {
	input := `#!/bin/bash
#@/command
 # Main entry.
 ##
main() { :; }

#@/subcommand push
 # Push it.
 ##
cmd_push() { :; }

#@/subcommand pull
 # Pull it.
 ##
cmd_pull() { :; }
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 3 {
		t.Fatalf("got %d blocks, want 3", len(doc.Blocks))
	}
	if doc.Blocks[0].Visibility != VisibilityCommand {
		t.Errorf("Block 0 Visibility = %q", doc.Blocks[0].Visibility)
	}
	if doc.Blocks[1].Name != "push" {
		t.Errorf("Block 1 Name = %q", doc.Blocks[1].Name)
	}
	if doc.Blocks[2].Name != "pull" {
		t.Errorf("Block 2 Name = %q", doc.Blocks[2].Name)
	}
}

func TestParseFunctionKeyword(t *testing.T) {
	input := `#!/bin/bash
#@/public
 # A function.
 ##
function my_func {
    echo "hello"
}
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	if doc.Blocks[0].FunctionName != "my_func" {
		t.Errorf("FunctionName = %q, want %q", doc.Blocks[0].FunctionName, "my_func")
	}
}

func TestParseWarningOnUnknownShedocTag(t *testing.T) {
	input := `#!/bin/bash
#?/foobar something
`
	doc := mustParse(t, input)
	if len(doc.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(doc.Warnings))
	}
	if !strings.Contains(doc.Warnings[0].Message, "foobar") {
		t.Errorf("Warning message = %q, want mention of foobar", doc.Warnings[0].Message)
	}
}

func TestParseSetsAndWrites(t *testing.T) {
	input := `#!/bin/bash
#@/subcommand rollback
 # Rolls back.
 #
 # @sets DEPLOY_LAST_ROLLBACK Timestamp of last rollback
 # @writes /var/log/deploy.log Rollback log entry
 ##
cmd_rollback() { :; }
`
	doc := mustParse(t, input)
	b := doc.Blocks[0]
	if len(b.Sets) != 1 {
		t.Fatalf("got %d sets, want 1", len(b.Sets))
	}
	if b.Sets[0].Name != "DEPLOY_LAST_ROLLBACK" {
		t.Errorf("Sets.Name = %q", b.Sets[0].Name)
	}
	if len(b.Writes) != 1 {
		t.Fatalf("got %d writes, want 1", len(b.Writes))
	}
	if b.Writes[0].Path != "/var/log/deploy.log" {
		t.Errorf("Writes.Path = %q", b.Writes[0].Path)
	}
}

func TestParseStdinTag(t *testing.T) {
	input := `#!/bin/bash
#@/command
 # Does things.
 #
 # @stdin Reads input data
 ##
`
	doc := mustParse(t, input)
	b := doc.Blocks[0]
	if b.Stdin == nil {
		t.Fatal("Stdin is nil")
	}
	if b.Stdin.Description != "Reads input data" {
		t.Errorf("Stdin.Description = %q", b.Stdin.Description)
	}
}

func TestParseTagContinuationAllTypes(t *testing.T) {
	input := `#!/bin/bash
#@/command
 # Does things.
 #
 # @flag -v | --verbose Enable verbose
 #                      output mode
 # @option -c | --config <path> Path to
 #                              config file
 # @operand <name> The name of
 #                 the thing
 # @reads ~/.config Read from
 #                  config dir
 # @stdin Reads data
 #        from pipe
 # @exit 0 Everything
 #        went fine
 # @stdout Outputs the
 #         result text
 # @stderr Error and
 #         diagnostic messages
 # @sets MY_VAR Sets this
 #              variable
 # @writes /tmp/out Writes output
 #                  to file
 ##
`
	doc := mustParse(t, input)
	b := doc.Blocks[0]

	if b.Flags[0].Description != "Enable verbose output mode" {
		t.Errorf("Flag continuation: %q", b.Flags[0].Description)
	}
	if b.Options[0].Description != "Path to config file" {
		t.Errorf("Option continuation: %q", b.Options[0].Description)
	}
	if b.Operands[0].Description != "The name of the thing" {
		t.Errorf("Operand continuation: %q", b.Operands[0].Description)
	}
	if b.Reads[0].Description != "Read from config dir" {
		t.Errorf("Reads continuation: %q", b.Reads[0].Description)
	}
	if b.Stdin.Description != "Reads data from pipe" {
		t.Errorf("Stdin continuation: %q", b.Stdin.Description)
	}
	if b.Exit[0].Description != "Everything went fine" {
		t.Errorf("Exit continuation: %q", b.Exit[0].Description)
	}
	if b.Stdout.Description != "Outputs the result text" {
		t.Errorf("Stdout continuation: %q", b.Stdout.Description)
	}
	if b.Stderr.Description != "Error and diagnostic messages" {
		t.Errorf("Stderr continuation: %q", b.Stderr.Description)
	}
	if b.Sets[0].Description != "Sets this variable" {
		t.Errorf("Sets continuation: %q", b.Sets[0].Description)
	}
	if b.Writes[0].Description != "Writes output to file" {
		t.Errorf("Writes continuation: %q", b.Writes[0].Description)
	}
}

func TestParseDeprecatedContinuation(t *testing.T) {
	input := `#!/bin/bash
#@/subcommand old
 # @deprecated This is deprecated.
 #             Use something else.
 ##
`
	doc := mustParse(t, input)
	b := doc.Blocks[0]
	if b.Deprecated == nil {
		t.Fatal("Deprecated is nil")
	}
	if b.Deprecated.Message != "This is deprecated. Use something else." {
		t.Errorf("Deprecated continuation: %q", b.Deprecated.Message)
	}
}

func TestParseShedocBlockEOF(t *testing.T) {
	// EOF while inside a #?/ block — should finalize gracefully.
	input := `#!/bin/bash
#?/description
 # This description has no closing marker`
	doc := mustParse(t, input)
	if doc.Meta.Description != "This description has no closing marker" {
		t.Errorf("Description = %q", doc.Meta.Description)
	}
}

func TestParseSheblockEOF(t *testing.T) {
	// EOF while inside a #@/ block — should finalize gracefully.
	input := `#!/bin/bash
#@/command
 # A command with no closing marker
 # @flag -v | --verbose Verbose`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	if doc.Blocks[0].Description != "A command with no closing marker" {
		t.Errorf("Description = %q", doc.Blocks[0].Description)
	}
	if len(doc.Blocks[0].Flags) != 1 {
		t.Fatalf("got %d flags, want 1", len(doc.Blocks[0].Flags))
	}
}

func TestParseShedocBlockInterrupted(t *testing.T) {
	// A non-continuation line inside a #?/ block should finalize and reprocess.
	input := `#!/bin/bash
#?/description
 # Some description
#?/version 1.0.0
`
	doc := mustParse(t, input)
	if doc.Meta.Description != "Some description" {
		t.Errorf("Description = %q", doc.Meta.Description)
	}
	if doc.Meta.Version != "1.0.0" {
		t.Errorf("Version = %q", doc.Meta.Version)
	}
}

func TestParseSheblockInterrupted(t *testing.T) {
	// A non-continuation line inside a #@/ block should finalize and reprocess.
	input := `#!/bin/bash
#@/public
 # A function.
 # @flag -v Verbose
some_func() { :; }
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	// Function detected after block was finalized by non-continuation line.
	if doc.Blocks[0].FunctionName != "some_func" {
		t.Errorf("FunctionName = %q, want %q", doc.Blocks[0].FunctionName, "some_func")
	}
}

func TestParseUnknownVisibility(t *testing.T) {
	input := `#!/bin/bash
#@/foobar
 # Unknown visibility defaults to public.
 ##
my_func() { :; }
`
	doc := mustParse(t, input)
	if len(doc.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(doc.Blocks))
	}
	if doc.Blocks[0].Visibility != VisibilityPublic {
		t.Errorf("Visibility = %q, want %q", doc.Blocks[0].Visibility, VisibilityPublic)
	}
}

func TestParseWarningOnBadTag(t *testing.T) {
	input := `#!/bin/bash
#@/command
 # A command.
 # @bogustag some value
 ##
`
	doc := mustParse(t, input)
	if len(doc.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(doc.Warnings))
	}
	if !strings.Contains(doc.Warnings[0].Message, "bogustag") {
		t.Errorf("Warning = %q", doc.Warnings[0].Message)
	}
}

func TestParseTagWithNoContent(t *testing.T) {
	// @tag with no following text on the line, just the tag name.
	input := `#!/bin/bash
#@/command
 # @stdout
 ##
`
	doc := mustParse(t, input)
	if doc.Blocks[0].Stdout == nil {
		t.Fatal("Stdout is nil")
	}
	if doc.Blocks[0].Stdout.Description != "" {
		t.Errorf("Stdout.Description = %q, want empty", doc.Blocks[0].Stdout.Description)
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/path/to/script.sh")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseTagContinuationNoInitialDescription(t *testing.T) {
	// Tag with no description on the @tag line, only continuation lines.
	// This exercises joinDesc with empty existing string.
	input := `#!/bin/bash
#@/command
 # @flag -v
 #       Verbose output mode
 ##
`
	doc := mustParse(t, input)
	if len(doc.Blocks[0].Flags) != 1 {
		t.Fatalf("got %d flags, want 1", len(doc.Blocks[0].Flags))
	}
	if doc.Blocks[0].Flags[0].Description != "Verbose output mode" {
		t.Errorf("Flag.Description = %q, want %q", doc.Blocks[0].Flags[0].Description, "Verbose output mode")
	}
}

func mustParse(t *testing.T, input string) *Document {
	t.Helper()
	doc, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseReader() error: %v", err)
	}
	return doc
}
