package cardgen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/tcg-cardgen/internal/metadata"
	"github.com/Merith-TK/tcg-cardgen/internal/templates"
)

// Config holds configuration for the card generator
type Config struct {
	TemplateDir  string
	OutputDir    string
	ValidateOnly bool
	Verbose      bool
}

// Generator handles card generation
type Generator struct {
	config          *Config
	templateManager *templates.Manager
	metadataParser  *metadata.Parser
}

// NewGenerator creates a new card generator with the given config
func NewGenerator(config *Config) *Generator {
	if config.OutputDir == "" {
		config.OutputDir = ".tcg-cardgen-out"
	}

	return &Generator{
		config:          config,
		templateManager: templates.NewManager(config.TemplateDir),
		metadataParser:  metadata.NewParser(),
	}
}

// GenerateCard processes a single markdown file and generates a card
func (g *Generator) GenerateCard(filePath string) error {
	if g.config.Verbose {
		fmt.Printf("Parsing metadata from: %s\n", filePath)
	}

	// Parse the markdown file
	card, err := g.metadataParser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", filePath, err)
	}

	if g.config.Verbose {
		fmt.Printf("Card TCG: %s, Title: %s\n", card.TCG, card.Title)
	}

	// Load appropriate template
	templateName := "basic"
	if card.TCG == "mtg" {
		templateName = "streamlined" // Use our new MTG template for testing
	}

	template, err := g.templateManager.LoadTemplate(card.TCG, templateName)
	if err != nil {
		return fmt.Errorf("failed to load template for %s: %v", card.TCG, err)
	}

	// Validate card against template
	if err := template.ValidateCard(card); err != nil {
		return fmt.Errorf("card validation failed: %v", err)
	}

	if g.config.ValidateOnly {
		fmt.Printf("✓ %s is valid\n", filePath)
		return nil
	}

	// Create output directory
	outputDir := filepath.Join(filepath.Dir(filePath), g.config.OutputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate output filename
	baseFilename := filepath.Base(filePath)
	nameWithoutExt := baseFilename[:len(baseFilename)-len(filepath.Ext(baseFilename))]
	outputPath := filepath.Join(outputDir, nameWithoutExt+".png")

	if g.config.Verbose {
		fmt.Printf("Output path: %s\n", outputPath)
	}

	// TODO: Implement actual rendering
	fmt.Printf("Would generate: %s -> %s\n", filePath, outputPath)

	return nil
}
