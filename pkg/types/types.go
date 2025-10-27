package types

// Common types shared across packages

// CardStyleInfo represents information about a discovered cardstyle
type CardStyleInfo struct {
	TCG         string
	Name        string
	DisplayName string
	Description string
	Version     string
	Source      string // "embedded", "workspace", "user", or file path
	Extends     string // Base template it extends
}

// Config holds configuration for the card generator
type Config struct {
	TemplateDir  string
	OutputDir    string
	ValidateOnly bool
	Verbose      bool
}
