package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/cardgen"
)

func main() {
	var (
		templateDir   = flag.String("template-dir", "", "Custom template directory")
		outputDir     = flag.String("output-dir", "", "Custom output directory (default: .tcg-cardgen-out)")
		validateOnly  = flag.Bool("validate-only", false, "Validate cards without generating")
		listTemplates = flag.Bool("list-templates", false, "List available templates")
		verbose       = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	if *listTemplates {
		// Initialize template manager to discover cardstyles
		generator := cardgen.NewGenerator(&cardgen.Config{
			TemplateDir: *templateDir,
		})

		if err := listAvailableCardstyles(generator); err != nil {
			log.Fatalf("Error listing templates: %v", err)
		}
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file_or_directory>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputPath := args[0]

	// Initialize the card generator
	generator := cardgen.NewGenerator(&cardgen.Config{
		TemplateDir:  *templateDir,
		OutputDir:    *outputDir,
		ValidateOnly: *validateOnly,
		Verbose:      *verbose,
	})

	// Process input
	err := processInput(generator, inputPath)
	if err != nil {
		log.Fatalf("Error processing input: %v", err)
	}
}

func processInput(generator *cardgen.Generator, inputPath string) error {
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("cannot access %s: %v", inputPath, err)
	}

	if info.IsDir() {
		return processDirectory(generator, inputPath)
	} else {
		return processFile(generator, inputPath)
	}
}

func processDirectory(generator *cardgen.Generator, dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".md" {
			return processFile(generator, path)
		}

		return nil
	})
}

func processFile(generator *cardgen.Generator, filePath string) error {
	fmt.Printf("Processing: %s\n", filePath)
	return generator.GenerateCard(filePath)
}

func listAvailableCardstyles(generator *cardgen.Generator) error {
	cardstyles, err := generator.ListCardstyles()
	if err != nil {
		return fmt.Errorf("failed to discover cardstyles: %v", err)
	}

	if len(cardstyles) == 0 {
		fmt.Println("No cardstyles found.")
		return nil
	}

	fmt.Println("Available Cardstyles:")
	fmt.Println()

	// Group by TCG
	tcgGroups := make(map[string][]cardgen.CardStyleInfo)
	for _, style := range cardstyles {
		tcgGroups[style.TCG] = append(tcgGroups[style.TCG], style)
	}

	for tcg, styles := range tcgGroups {
		fmt.Printf("🎮 %s:\n", strings.ToUpper(tcg))
		for _, style := range styles {
			fmt.Printf("  📄 %s/%s", tcg, style.Name)
			if style.DisplayName != "" && style.DisplayName != style.Name {
				fmt.Printf(" (%s)", style.DisplayName)
			}
			fmt.Println()

			if style.Description != "" {
				fmt.Printf("     %s\n", style.Description)
			}

			if style.Extends != "" {
				fmt.Printf("     Extends: %s\n", style.Extends)
			}

			if style.Source != "built-in" {
				fmt.Printf("     Source: %s\n", style.Source)
			}

			fmt.Println()
		}
	}

	return nil
}
