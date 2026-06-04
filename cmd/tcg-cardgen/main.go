package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/generator"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
)

func main() {
	var (
		templateDir  = flag.String("template-dir", "", "Extra template search directory")
		outputDir    = flag.String("output-dir", "", "Output directory (default: .tcg-cardgen-out next to each card)")
		validateOnly = flag.Bool("validate-only", false, "Validate cards without rendering")
		listTpl      = flag.Bool("list-templates", false, "List available card styles and exit")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		concurrency  = flag.Int("concurrency", 1, "Number of parallel render workers")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file.md|directory> ...\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg := &types.Config{
		TemplateDir:  *templateDir,
		OutputDir:    *outputDir,
		ValidateOnly: *validateOnly,
		Verbose:      *verbose,
		Concurrency:  *concurrency,
	}

	gen := generator.New(cfg)

	if *listTpl {
		if err := listCardstyles(gen); err != nil {
			log.Fatalf("error: %v", err)
		}
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	exitCode := 0
	for _, arg := range args {
		if err := processInput(gen, cfg, arg); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func processInput(gen *generator.Generator, cfg *types.Config, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", path, err)
	}
	if info.IsDir() {
		return gen.GenerateDirectory(path)
	}
	return gen.GenerateCard(path)
}

func listCardstyles(gen *generator.Generator) error {
	styles, err := gen.ListCardstyles()
	if err != nil {
		return err
	}
	if len(styles) == 0 {
		fmt.Println("No card styles found.")
		return nil
	}

	// Group by TCG.
	byTCG := make(map[string][]types.CardStyleInfo)
	var tcgOrder []string
	for _, s := range styles {
		if _, exists := byTCG[s.TCG]; !exists {
			tcgOrder = append(tcgOrder, s.TCG)
		}
		byTCG[s.TCG] = append(byTCG[s.TCG], s)
	}

	fmt.Println("Available card styles:")
	for _, tcg := range tcgOrder {
		fmt.Printf("\n  %s\n", strings.ToUpper(tcg))
		for _, s := range byTCG[tcg] {
			line := fmt.Sprintf("    %s/%s", tcg, s.Name)
			if s.DisplayName != "" && s.DisplayName != s.Name {
				line += fmt.Sprintf("  (%s)", s.DisplayName)
			}
			fmt.Println(line)
			if s.Description != "" {
				fmt.Printf("      %s\n", s.Description)
			}
			if s.Extends != "" {
				fmt.Printf("      extends: %s\n", filepath.Base(s.Extends))
			}
			if s.Source != "" && s.Source != "embedded" {
				fmt.Printf("      source: %s\n", s.Source)
			}
		}
	}
	return nil
}
