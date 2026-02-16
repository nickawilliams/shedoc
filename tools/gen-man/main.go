package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nickawilliams/shedoc/internal/cli"
	"github.com/spf13/pflag"
)

func main() {
	outDir := os.Getenv("MAN_OUT_DIR")
	if outDir == "" {
		outDir = filepath.Join("contrib", "man")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("unable to create man dir: %v", err)
	}

	root := cli.NewRootCmd("dev")
	outPath := filepath.Join(outDir, "shedoc.1")

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("unable to create man page file: %v", err)
	}
	defer f.Close()

	date := time.Now().Format("Jan 2006")

	// Header
	fmt.Fprintln(f, ".nh")
	fmt.Fprintf(f, ".TH \"SHEDOC\" \"1\" \"%s\" \"shedoc\" \"User Commands\"\n", date)

	// NAME
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, ".SH NAME")
	fmt.Fprintf(f, "shedoc \\- %s\n", root.Short)

	// SYNOPSIS
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, ".SH SYNOPSIS")
	fmt.Fprintln(f, `\fBshedoc\fP [\fIflags\fP] <\fIfile\fP...>`)

	// DESCRIPTION
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, ".SH DESCRIPTION")
	desc := root.Short
	if root.Long != "" {
		desc = root.Long
	}
	fmt.Fprintln(f, escapeManPage(desc))

	// OPTIONS
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, ".SH OPTIONS")
	root.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		writeOption(f, flag)
	})
	// Help flag (Cobra adds it automatically but not via Flags())
	fmt.Fprintln(f, ".TP")
	fmt.Fprintln(f, `\fB\-h\fP, \fB\-\-help\fP`)
	fmt.Fprintln(f, "Display help and exit.")
	// Version flag
	fmt.Fprintln(f, ".TP")
	fmt.Fprintln(f, `\fB\-v\fP, \fB\-\-version\fP`)
	fmt.Fprintln(f, "Display version and exit.")

	// SUBCOMMANDS
	subs := root.Commands()
	if len(subs) > 0 {
		fmt.Fprintln(f, "")
		fmt.Fprintln(f, ".SH SUBCOMMANDS")
		for _, sub := range subs {
			if sub.Hidden {
				continue
			}
			fmt.Fprintln(f, ".TP")
			fmt.Fprintf(f, "\\fB%s\\fP\n", sub.Name())
			fmt.Fprintln(f, sub.Short)
		}
	}

	// EXIT STATUS
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, ".SH EXIT STATUS")
	fmt.Fprintln(f, ".TP")
	fmt.Fprintln(f, "0")
	fmt.Fprintln(f, "Success.")
	fmt.Fprintln(f, ".TP")
	fmt.Fprintln(f, "1")
	fmt.Fprintln(f, "An error occurred.")

	// SEE ALSO
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, ".SH SEE ALSO")
	fmt.Fprintln(f, `\fBbash\fP(1), \fBzsh\fP(1), \fBfish\fP(1)`)

	fmt.Printf("wrote man page to %s\n", outPath)
}

func escapeManPage(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "-", `\-`)
	return s
}

func writeOption(f *os.File, flag *pflag.Flag) {
	fmt.Fprintln(f, ".TP")

	var sig strings.Builder
	if flag.Shorthand != "" {
		sig.WriteString(fmt.Sprintf(`\fB\-%s\fP, `, flag.Shorthand))
	}
	escapedName := strings.ReplaceAll(flag.Name, "-", `\-`)
	sig.WriteString(fmt.Sprintf(`\fB\-\-%s\fP`, escapedName))

	if flag.Value.Type() != "bool" {
		sig.WriteString(fmt.Sprintf(`=\fI%s\fP`, strings.ToUpper(flag.Name)))
	}

	fmt.Fprintln(f, sig.String())

	desc := flag.Usage
	def := flag.DefValue
	if def != "" && def != "false" && def != "0" && def != `""` {
		fmt.Fprintf(f, "%s (default: %s).\n", capitalizeFirst(desc), def)
	} else {
		fmt.Fprintln(f, capitalizeFirst(desc)+".")
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
