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
	// Core card data
	TCG    string `yaml:"card.tcg"`
	Title  string `yaml:"card.title"`
	Type   string `yaml:"card.type"`
	Rarity string `yaml:"card.rarity"`
	Set    string `yaml:"card.set"`
	Artist string `yaml:"card.artist"`

	// Print information
	PrintThis  int `yaml:"card.print_this"`
	PrintTotal int `yaml:"card.print_total"`

	// Content
	Body     string `yaml:"-"` // Markdown content after frontmatter
	BodySize int    `yaml:"card.body.size"`

	// Artwork
	Artwork string `yaml:"card.artwork"`

	// Raw metadata for template-specific fields
	Metadata map[string]interface{} `yaml:",inline"`

	// Source file info
	SourceFile string `yaml:"-"`
}

// Parser handles parsing markdown files with YAML frontmatter
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

	// Check for YAML frontmatter
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("missing YAML frontmatter")
	}

	// Read frontmatter
	var frontmatterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	// Read remaining content (card body)
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse YAML frontmatter
	frontmatter := strings.Join(frontmatterLines, "\n")
	card := &Card{
		Metadata:   make(map[string]interface{}),
		SourceFile: filePath,
	}

	if err := yaml.Unmarshal([]byte(frontmatter), &card.Metadata); err != nil {
		return nil, fmt.Errorf("error parsing YAML frontmatter: %v", err)
	}

	// Also parse into struct fields
	if err := yaml.Unmarshal([]byte(frontmatter), card); err != nil {
		return nil, fmt.Errorf("error parsing YAML into struct: %v", err)
	}

	// Process body content - strip headers and clean up
	card.Body = p.processBody(bodyLines)

	// Set defaults
	p.setDefaults(card, filePath)

	return card, nil
}

// processBody cleans up the markdown content, removing headers
func (p *Parser) processBody(lines []string) string {
	var cleanLines []string

	for _, line := range lines {
		// Skip markdown headers (lines starting with #)
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	// Join and trim whitespace
	body := strings.Join(cleanLines, "\n")
	return strings.TrimSpace(body)
}

// setDefaults sets default values for missing fields
func (p *Parser) setDefaults(card *Card, filePath string) {
	// Default title to filename if not set
	if card.Title == "" {
		baseFilename := filepath.Base(filePath)
		nameWithoutExt := baseFilename[:len(baseFilename)-len(filepath.Ext(baseFilename))]
		// Convert underscores to spaces and title case
		card.Title = strings.Title(strings.ReplaceAll(nameWithoutExt, "_", " "))
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

	// Default body size
	if card.BodySize == 0 {
		card.BodySize = 12
	}
}
