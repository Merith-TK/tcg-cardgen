package metadata

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Card represents a parsed card with metadata and content
type Card struct {
	// Core card data (extracted from body or frontmatter)
	TCG    string `yaml:"card.tcg"`
	Title  string `yaml:"card.title"`
	Type   string `yaml:"card.type"`
	Rarity string `yaml:"card.rarity"`
	Set    string `yaml:"card.set"`
	Artist string `yaml:"card.artist"`

	// Print information
	PrintThis  int `yaml:"card.print_this"`
	PrintTotal int `yaml:"card.print_total"`

	// Content sections (parsed from body)
	Body       string `yaml:"-"` // Full markdown content after frontmatter
	RulesText  string `yaml:"-"` // Extracted rules text
	FlavorText string `yaml:"-"` // Extracted flavor text
	ManaCost   string `yaml:"-"` // Extracted mana cost

	// Raw metadata for template-specific fields
	Metadata map[string]interface{} `yaml:",inline"`

	// Source file info
	SourceFile string `yaml:"-"`
}

// Parser handles parsing markdown files with YAML frontmatter and body extraction
type Parser struct{}

// NewParser creates a new metadata parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a markdown file and extracts metadata and content
func (p *Parser) ParseFile(filePath string) (*Card, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Check for YAML frontmatter (optional)
	var frontmatterLines []string
	var bodyLines []string

	if scanner.Scan() && scanner.Text() == "---" {
		// Read frontmatter
		for scanner.Scan() {
			line := scanner.Text()
			if line == "---" {
				break
			}
			frontmatterLines = append(frontmatterLines, line)
		}
	} else {
		// No frontmatter, add first line to body
		bodyLines = append(bodyLines, scanner.Text())
	}

	// Read remaining content (card body)
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Initialize card
	card := &Card{
		Metadata:   make(map[string]interface{}),
		SourceFile: filePath,
		Body:       strings.Join(bodyLines, "\n"),
	}

	// Parse YAML frontmatter if present
	if len(frontmatterLines) > 0 {
		frontmatter := strings.Join(frontmatterLines, "\n")

		if err := yaml.Unmarshal([]byte(frontmatter), &card.Metadata); err != nil {
			return nil, fmt.Errorf("error parsing YAML frontmatter: %v", err)
		}

		// Also parse into struct fields
		if err := yaml.Unmarshal([]byte(frontmatter), card); err != nil {
			return nil, fmt.Errorf("error parsing YAML into struct: %v", err)
		}
	}

	// Parse structured data from markdown body
	if err := p.parseBodyContent(card); err != nil {
		return nil, fmt.Errorf("error parsing body content: %v", err)
	}

	// Set defaults
	p.setDefaults(card, filePath)

	return card, nil
}

// parseBodyContent extracts structured data from the markdown body
func (p *Parser) parseBodyContent(card *Card) error {
	lines := strings.Split(card.Body, "\n")

	var rulesLines []string
	var flavorLines []string
	inFlavorSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract title from # Header (only if not set in frontmatter)
		if card.Title == "" && strings.HasPrefix(line, "# ") {
			card.Title = strings.TrimSpace(line[2:])
			continue
		}

		// Extract mana cost from > {{mtg.cost...}} blockquote
		if strings.HasPrefix(line, "> {{") && strings.HasSuffix(line, "}}") {
			if card.ManaCost == "" { // Only set if not already set
				card.ManaCost = strings.TrimSpace(line[2:]) // Remove "> "
			}
			continue
		}

		// Extract type from > **Type** blockquote
		if strings.HasPrefix(line, "> **") && strings.HasSuffix(line, "**") {
			if card.Type == "" { // Only set if not already set
				// Extract text between > ** and **
				typeText := line[4 : len(line)-2] // Remove "> **" and "**"
				card.Type = strings.TrimSpace(typeText)
			}
			continue
		}

		// Check for flavor text separator (horizontal rule)
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "-----") {
			inFlavorSection = true
			continue
		}

		// Skip empty lines and headers
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract flavor text (italic lines after separator)
		if inFlavorSection {
			if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") && len(line) > 2 {
				// Remove surrounding asterisks
				flavorText := line[1 : len(line)-1]
				flavorLines = append(flavorLines, flavorText)
			}
			continue
		}

		// Everything else is rules text
		if line != "" {
			rulesLines = append(rulesLines, line)
		}
	}

	// Join the extracted content
	card.RulesText = strings.Join(rulesLines, "\n\n")
	card.FlavorText = strings.Join(flavorLines, "\n")

	return nil
}

// setDefaults sets default values for missing fields
func (p *Parser) setDefaults(card *Card, filePath string) {
	// Default title to filename if not set
	if card.Title == "" {
		baseFilename := filepath.Base(filePath)
		nameWithoutExt := baseFilename[:len(baseFilename)-len(filepath.Ext(baseFilename))]
		// Convert underscores to spaces and capitalize first letter
		titleText := strings.ReplaceAll(nameWithoutExt, "_", " ")
		if len(titleText) > 0 {
			titleText = strings.ToUpper(titleText[:1]) + titleText[1:]
		}
		card.Title = titleText
	}

	// Default print info
	if card.PrintThis == 0 {
		card.PrintThis = 1
	}
	if card.PrintTotal == 0 {
		card.PrintTotal = 1
	}

	// Default rarity
	if card.Rarity == "" {
		card.Rarity = "common"
	}

	// Default set
	if card.Set == "" {
		card.Set = "Unknown"
	}

	// Default artist
	if card.Artist == "" {
		card.Artist = "Unknown Artist"
	}

	// Default TCG
	if card.TCG == "" {
		card.TCG = "mtg" // Default to MTG for now
	}
}
