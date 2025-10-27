# API Reference

Complete reference for using the TCG Card Generator programmatically.

## 📦 Packages

### `pkg/cardgen`
Main card generation API
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/cardgen"
```

### `pkg/metadata` 
Card metadata parsing and validation
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/metadata"
```

### `pkg/renderer`
Image rendering and template processing  
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/renderer"
```

### `pkg/templates`
Template discovery and management
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/templates"
```

### `pkg/types`
Shared types and configuration
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/types"
```

## 🚀 Quick Start

### Generate a Single Card
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/Merith-TK/tcg-cardgen/pkg/cardgen"
    "github.com/Merith-TK/tcg-cardgen/pkg/types"
)

func main() {
    // Create generator with default configuration
    generator, err := cardgen.New(types.Config{
        OutputDir:    "./output",
        TemplateDir:  "./templates",
        UserDataDir:  "~/.tcg-cardgen",
        Verbose:      false,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate card from markdown file
    ctx := context.Background()
    outputPath, err := generator.GenerateCard(ctx, "cards/lightning_bolt.md")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Generated card: %s\n", outputPath)
}
```

### Batch Generation
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/Merith-TK/tcg-cardgen/pkg/cardgen"
    "github.com/Merith-TK/tcg-cardgen/pkg/types"
)

func main() {
    generator, err := cardgen.New(types.Config{
        OutputDir: "./output",
        Verbose:   true,  // Enable detailed logging
    })
    if err != nil {
        log.Fatal(err)
    }
    
    cardFiles := []string{
        "cards/lightning_bolt.md",
        "cards/dark_ritual.md", 
        "cards/giant_growth.md",
    }
    
    ctx := context.Background()
    results, err := generator.GenerateBatch(ctx, cardFiles)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("Failed %s: %v\n", result.InputPath, result.Error)
        } else {
            fmt.Printf("Generated %s -> %s\n", result.InputPath, result.OutputPath)
        }
    }
}
```

## 📋 Core Types

### `types.Config`
Main configuration structure:
```go
type Config struct {
    OutputDir    string  // Output directory for generated cards
    TemplateDir  string  // Template directory (optional, uses embedded if empty)
    UserDataDir  string  // User data directory for custom templates
    Verbose      bool    // Enable verbose logging
}
```

### `types.CardStyleInfo`
Template information:
```go
type CardStyleInfo struct {
    Name        string // Template display name
    TCG         string // Target TCG (mtg, pokemon, etc.)
    Description string // Template description  
    Version     string // Template version
    Source      string // File path or "embedded"
    FilePath    string // Full file path
}
```

### `cardgen.BatchResult`
Batch generation result:
```go
type BatchResult struct {
    InputPath  string // Input markdown file
    OutputPath string // Generated image file
    Error      error  // Generation error (if any)
}
```

## 🎯 Core APIs

### `cardgen.New(config Config) (*Generator, error)`
Create a new card generator instance.

**Parameters:**
- `config`: Configuration object

**Returns:**
- `*Generator`: Generator instance
- `error`: Initialization error

**Example:**
```go
generator, err := cardgen.New(types.Config{
    OutputDir:   "./output",
    UserDataDir: "~/.tcg-cardgen",
    Verbose:     true,
})
```

### `(*Generator).GenerateCard(ctx context.Context, inputPath string) (string, error)`
Generate a single card from a markdown file.

**Parameters:**
- `ctx`: Context for cancellation
- `inputPath`: Path to markdown card file

**Returns:**
- `string`: Path to generated image
- `error`: Generation error

**Example:**
```go
ctx := context.Background()
outputPath, err := generator.GenerateCard(ctx, "cards/my_card.md")
```

### `(*Generator).GenerateBatch(ctx context.Context, inputPaths []string) ([]BatchResult, error)`
Generate multiple cards in batch.

**Parameters:**
- `ctx`: Context for cancellation  
- `inputPaths`: Slice of markdown file paths

**Returns:**
- `[]BatchResult`: Results for each input file
- `error`: Critical error (individual errors in results)

**Example:**
```go
results, err := generator.GenerateBatch(ctx, []string{
    "cards/card1.md",
    "cards/card2.md",
})

for _, result := range results {
    if result.Error != nil {
        log.Printf("Failed %s: %v", result.InputPath, result.Error)
    }
}
```

### `(*Generator).ValidateCard(inputPath string) error`
Validate a card without generating the image.

**Parameters:**
- `inputPath`: Path to markdown card file

**Returns:**
- `error`: Validation error

**Example:**
```go
err := generator.ValidateCard("cards/test_card.md")
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
}
```

### `(*Generator).ListTemplates() (map[string][]CardStyleInfo, error)`
List all available templates organized by TCG.

**Returns:**
- `map[string][]CardStyleInfo`: Templates grouped by TCG
- `error`: Discovery error

**Example:**
```go
templates, err := generator.ListTemplates()
if err != nil {
    log.Fatal(err)
}

for tcg, cardstyles := range templates {
    fmt.Printf("TCG: %s\n", tcg)
    for _, cardstyle := range cardstyles {
        fmt.Printf("  %s (%s)\n", cardstyle.Name, cardstyle.Version)
    }
}
```

## 🔍 Template Discovery

### `templates.Discover(templateDir, userDataDir string) (map[string][]CardStyleInfo, error)`
Discover templates from multiple sources.

**Parameters:**
- `templateDir`: Project template directory (can be empty)
- `userDataDir`: User data directory (can be empty)

**Returns:**
- `map[string][]CardStyleInfo`: Discovered templates by TCG
- `error`: Discovery error

**Example:**
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/templates"

discovered, err := templates.Discover("./templates", "~/.tcg-cardgen")
if err != nil {
    log.Fatal(err)
}
```

### `templates.LoadTemplate(tcg, cardstyle, templateDir, userDataDir string) (*CardStyle, error)`
Load a specific template.

**Parameters:**
- `tcg`: Target TCG name
- `cardstyle`: Cardstyle name  
- `templateDir`: Project template directory
- `userDataDir`: User data directory

**Returns:**
- `*CardStyle`: Loaded template
- `error`: Loading error

**Example:**
```go
cardStyle, err := templates.LoadTemplate("mtg", "basic", "./templates", "~/.tcg-cardgen")
if err != nil {
    log.Fatal(err)
}
```

## 📄 Metadata Parsing

### `metadata.ParseCard(filePath string) (*CardData, error)`
Parse card metadata from markdown file.

**Parameters:**
- `filePath`: Path to markdown file

**Returns:**
- `*CardData`: Parsed card data
- `error`: Parsing error

**Example:**
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/metadata"

cardData, err := metadata.ParseCard("cards/my_card.md")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Card: %s\n", cardData.Metadata["card.title"])
fmt.Printf("TCG: %s\n", cardData.Metadata["card.tcg"])
```

### `(*CardData).GetString(key string) string`
Get string value from card metadata.

**Example:**
```go
title := cardData.GetString("card.title")
tcg := cardData.GetString("card.tcg")
```

### `(*CardData).GetInt(key string) int`
Get integer value from card metadata.

**Example:**
```go
cmc := cardData.GetInt("mtg.cmc")
hp := cardData.GetInt("pkm.hp")
```

### `(*CardData).GetSlice(key string) []interface{}`
Get slice value from card metadata.

**Example:**
```go
manaCost := cardData.GetSlice("mtg.mana_cost")
types := cardData.GetSlice("pkm.types")
```

## 🖼️ Rendering

### `renderer.RenderCard(cardData *CardData, cardStyle *CardStyle, outputPath string) error`
Render a card to an image file.

**Parameters:**
- `cardData`: Parsed card data
- `cardStyle`: Loaded template
- `outputPath`: Output image path

**Returns:**
- `error`: Rendering error

**Example:**
```go
import (
    "github.com/Merith-TK/tcg-cardgen/pkg/metadata"
    "github.com/Merith-TK/tcg-cardgen/pkg/renderer"
    "github.com/Merith-TK/tcg-cardgen/pkg/templates"
)

// Parse card
cardData, err := metadata.ParseCard("cards/my_card.md")
if err != nil {
    log.Fatal(err)
}

// Load template
cardStyle, err := templates.LoadTemplate("mtg", "basic", "", "~/.tcg-cardgen")
if err != nil {
    log.Fatal(err)
}

// Render
err = renderer.RenderCard(cardData, cardStyle, "output/my_card.png")
if err != nil {
    log.Fatal(err)
}
```

## 🔧 Advanced Usage

### Custom Configuration
```go
import (
    "os"
    "path/filepath"
    
    "github.com/Merith-TK/tcg-cardgen/pkg/cardgen"
    "github.com/Merith-TK/tcg-cardgen/pkg/types"
)

func main() {
    // Custom output directory
    homeDir, _ := os.UserHomeDir()
    customOutputDir := filepath.Join(homeDir, "Desktop", "cards")
    
    // Custom user data directory  
    customUserDataDir := filepath.Join(homeDir, "Documents", "tcg-templates")
    
    generator, err := cardgen.New(types.Config{
        OutputDir:    customOutputDir,
        TemplateDir:  "./project-templates",  // Project-specific templates
        UserDataDir:  customUserDataDir,      // Custom user templates
        Verbose:      true,                   // Detailed logging
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Use generator...
}
```

### Error Handling
```go
func generateWithRetry(generator *cardgen.Generator, inputPath string, maxRetries int) (string, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        outputPath, err := generator.GenerateCard(ctx, inputPath)
        if err == nil {
            return outputPath, nil
        }
        
        lastErr = err
        log.Printf("Attempt %d failed: %v", i+1, err)
        time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
    }
    
    return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
```

### Template Validation
```go
import "github.com/Merith-TK/tcg-cardgen/pkg/templates"

func validateTemplate(tcg, cardstyle string) error {
    cardStyle, err := templates.LoadTemplate(tcg, cardstyle, "", "~/.tcg-cardgen")
    if err != nil {
        return fmt.Errorf("failed to load template: %w", err)
    }
    
    // Check required fields
    if len(cardStyle.RequiredFields) == 0 {
        return fmt.Errorf("template has no required fields")
    }
    
    // Check dimensions
    if cardStyle.Dimensions.Width <= 0 || cardStyle.Dimensions.Height <= 0 {
        return fmt.Errorf("invalid dimensions: %dx%d", cardStyle.Dimensions.Width, cardStyle.Dimensions.Height)
    }
    
    return nil
}
```

### Concurrent Generation
```go
import (
    "sync"
    "context"
)

func generateConcurrent(generator *cardgen.Generator, inputPaths []string, maxWorkers int) []cardgen.BatchResult {
    jobs := make(chan string, len(inputPaths))
    results := make(chan cardgen.BatchResult, len(inputPaths))
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < maxWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for inputPath := range jobs {
                ctx := context.Background()
                outputPath, err := generator.GenerateCard(ctx, inputPath)
                results <- cardgen.BatchResult{
                    InputPath:  inputPath,
                    OutputPath: outputPath,
                    Error:      err,
                }
            }
        }()
    }
    
    // Submit jobs
    for _, inputPath := range inputPaths {
        jobs <- inputPath
    }
    close(jobs)
    
    // Wait for completion
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var batchResults []cardgen.BatchResult
    for result := range results {
        batchResults = append(batchResults, result)
    }
    
    return batchResults
}
```

## 🐛 Error Types

### Common Errors
```go
// Template not found
var ErrTemplateNotFound = errors.New("template not found")

// Invalid card format
var ErrInvalidCard = errors.New("invalid card format")

// Missing required field
var ErrMissingField = errors.New("missing required field")

// Rendering error
var ErrRenderingFailed = errors.New("rendering failed")
```

### Error Handling Patterns
```go
func handleGenerationError(err error) {
    switch {
    case errors.Is(err, ErrTemplateNotFound):
        fmt.Println("Template not found. Use --list-templates to see available templates.")
        
    case errors.Is(err, ErrMissingField):
        fmt.Println("Card is missing required fields. Check template requirements.")
        
    case errors.Is(err, ErrInvalidCard):
        fmt.Println("Card format is invalid. Check YAML frontmatter syntax.")
        
    default:
        fmt.Printf("Generation failed: %v\n", err)
    }
}
```

## 📚 Best Practices

### 1. **Resource Management**
```go
// Create generator once, reuse for multiple cards
generator, err := cardgen.New(config)
if err != nil {
    return err
}
defer generator.Close() // If cleanup is needed

// Reuse for multiple generations
for _, cardPath := range cardPaths {
    _, err := generator.GenerateCard(ctx, cardPath)
    // handle error
}
```

### 2. **Context Usage**
```go
// Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

outputPath, err := generator.GenerateCard(ctx, inputPath)
```

### 3. **Error Handling**
```go
// Validate before generating
if err := generator.ValidateCard(inputPath); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

// Generate with proper error handling
outputPath, err := generator.GenerateCard(ctx, inputPath)
if err != nil {
    return fmt.Errorf("generation failed for %s: %w", inputPath, err)
}
```

### 4. **Configuration**
```go
// Use environment variables for configuration
config := types.Config{
    OutputDir:   getEnvWithDefault("TCG_OUTPUT_DIR", "./output"),
    UserDataDir: getEnvWithDefault("TCG_USER_DATA_DIR", "~/.tcg-cardgen"),
    Verbose:     getEnvBool("TCG_VERBOSE", false),
}

func getEnvWithDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## 🔗 Related Documentation

- **[Creating Cards](creating-cards.md)** - Learn card file format
- **[Creating Templates](creating-templates.md)** - Build custom templates  
- **[CLI Usage](cli.md)** - Command-line interface reference
- **[Examples](examples.md)** - Complete code examples