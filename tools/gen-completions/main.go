package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nickawilliams/shedoc/internal/cli"
)

func main() {
	outDir := os.Getenv("COMPLETIONS_OUT_DIR")
	if outDir == "" {
		outDir = filepath.Join("contrib", "completions")
	}

	cmd := cli.NewRootCmd("dev")

	// Bash
	bashDir := filepath.Join(outDir, "bash")
	if err := os.MkdirAll(bashDir, 0o755); err != nil {
		log.Fatalf("unable to create bash completions dir: %v", err)
	}
	bashFile, err := os.Create(filepath.Join(bashDir, "shedoc.bash"))
	if err != nil {
		log.Fatalf("unable to create bash completion file: %v", err)
	}
	if err := cmd.GenBashCompletionV2(bashFile, true); err != nil {
		log.Fatalf("unable to generate bash completion: %v", err)
	}
	bashFile.Close()

	// Zsh
	zshDir := filepath.Join(outDir, "zsh")
	if err := os.MkdirAll(zshDir, 0o755); err != nil {
		log.Fatalf("unable to create zsh completions dir: %v", err)
	}
	zshFile, err := os.Create(filepath.Join(zshDir, "shedoc.zsh"))
	if err != nil {
		log.Fatalf("unable to create zsh completion file: %v", err)
	}
	if err := cmd.GenZshCompletion(zshFile); err != nil {
		log.Fatalf("unable to generate zsh completion: %v", err)
	}
	zshFile.Close()

	// Fish
	fishDir := filepath.Join(outDir, "fish")
	if err := os.MkdirAll(fishDir, 0o755); err != nil {
		log.Fatalf("unable to create fish completions dir: %v", err)
	}
	fishFile, err := os.Create(filepath.Join(fishDir, "shedoc.fish"))
	if err != nil {
		log.Fatalf("unable to create fish completion file: %v", err)
	}
	if err := cmd.GenFishCompletion(fishFile, true); err != nil {
		log.Fatalf("unable to generate fish completion: %v", err)
	}
	fishFile.Close()

	fmt.Printf("wrote completions to %s\n", outDir)
}
