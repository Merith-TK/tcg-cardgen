package cardgen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/tcg-cardgen/pkg/metadata"
	"github.com/Merith-TK/tcg-cardgen/pkg/renderer"
	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
)

// Generator handles card generation
type Generator struct {
	config          *types.Config
	templateManager *templates.Manager
	metadataParser  *metadata.Parser
	renderer        *renderer.Renderer
}

// NewGenerator creates a new card generator with the given config
func NewGenerator(config *types.Config) *Generator {
	if config.OutputDir == "" {
		config.OutputDir = ".tcg-cardgen-out"
	}

	return &Generator{
		config:          config,
		templateManager: templates.NewManager(config.TemplateDir),
		metadataParser:  metadata.NewParser(),
		renderer:        renderer.NewRenderer(),
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
		fmt.Printf("Card TCG: %s, CardStyle: %s, Title: %s\n", card.TCG, card.CardStyle, card.Title)
	}

	// Load appropriate template based on TCG and cardstyle
	template, err := g.templateManager.LoadTemplate(card.TCG, card.CardStyle)
	if err != nil {
		return fmt.Errorf("failed to load cardstyle %s/%s: %v", card.TCG, card.CardStyle, err)
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

	// Render the card
	if err := g.renderer.RenderCard(card, template, outputPath); err != nil {
		return fmt.Errorf("failed to render card: %v", err)
	}

	if g.config.Verbose {
		fmt.Printf("✓ Generated: %s\n", outputPath)
	} else {
		fmt.Printf("Generated: %s -> %s\n", filePath, outputPath)
	}

	return nil
}

// ListCardstyles discovers and lists all available cardstyles
func (g *Generator) ListCardstyles() ([]types.CardStyleInfo, error) {
	templateInfos, err := g.templateManager.ListAvailableCardstyles()
	if err != nil {
		return nil, err
	}

	// Convert template CardStyleInfo to common types version
	cardstyles := make([]types.CardStyleInfo, len(templateInfos))
	for i, info := range templateInfos {
		cardstyles[i] = types.CardStyleInfo{
			TCG:         info.TCG,
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Description: info.Description,
			Version:     info.Version,
			Source:      info.Source,
			Extends:     info.Extends,
		}
	}

	return cardstyles, nil
}
