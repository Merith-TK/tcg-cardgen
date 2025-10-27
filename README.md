# TCG Card Generator

A powerful, extensible trading card game (TCG) card generator written in Go. Create beautiful card images from Markdown files with YAML frontmatter using customizable templates.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-green.svg)

## ✨ Features

- 🎴 **Multi-TCG Support** - Built-in templates for Magic: The Gathering and Pokémon
- 🎨 **Smart Color Affinity** - Dynamic frame selection based on card properties
- 📝 **Markdown-Based Cards** - Write cards in simple Markdown with YAML metadata
- 🔧 **Template Inheritance** - Extend base templates for consistent styling
- 🏗️ **Extensible Architecture** - Add new TCGs without code changes
- 📦 **Self-Contained** - Embedded templates work out-of-the-box
- 🎯 **Workspace Customization** - Project-specific template overrides
- 🔍 **Template Discovery** - List and explore available cardstyles

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Merith-TK/tcg-cardgen.git
cd tcg-cardgen

# Build the application
go build ./cmd/tcg-cardgen

# Or install directly
go install ./cmd/tcg-cardgen
```

### Basic Usage

```bash
# Generate a single card
./tcg-cardgen examples/lightning_bolt_red.md

# Generate all cards in a directory
./tcg-cardgen examples/

# List available templates
./tcg-cardgen --list-templates

# Validate cards without generating images
./tcg-cardgen --validate-only examples/
```

### Your First Card

Create a file `my_card.md`:

```markdown
---
card:
  tcg: mtg
  cardstyle: basic
  title: "Lightning Bolt"
  type: "Instant"
  rarity: "common"

mtg:
  cmc: 1
  color: red
  mana_cost: ["{{mtg.mana_red}}"]
  type_line: "Instant"
---

# Lightning Bolt

**Lightning Bolt** deals 3 damage to any target.

*The spark of an idea, the flash of inspiration, the bolt of lightning that changes everything.*
```

Generate it:

```bash
./tcg-cardgen my_card.md
```

## 📚 Documentation

- **[Creating Cards](docs/creating-cards.md)** - Learn how to write card files and use the generator
- **[Creating Templates](docs/creating-templates.md)** - Build custom cardstyles and templates  
- **[API Reference](docs/api-reference.md)** - Use the generator programmatically

### Quick Links
- [Installation](#installation) - Get started quickly
- [Template Discovery](#template-discovery) - Find available cardstyles
- [Smart Color System](#smart-features) - Dynamic frame selection
- [Template Inheritance](#template-inheritance) - Extensible template system

## 🎮 Supported TCGs

### Magic: The Gathering (MTG)
- **Basic Cards** - Standard spells, creatures, artifacts
- **Token Cards** - Creature tokens with special styling  
- **Legendary Cards** - Legendary permanents with unique borders

### Pokémon (PKM)
- **Basic Cards** - Standard Pokémon cards
- Extensible for Trainer cards, Energy cards, etc.

## 🏗️ Architecture

```
tcg-cardgen/
├── cmd/               # CLI applications
├── pkg/               # Public API packages
│   ├── cardgen/      # Main generator
│   ├── metadata/     # Card parsing
│   ├── renderer/     # Image rendering
│   ├── templates/    # Template system
│   └── types/        # Common types
├── templates/        # Built-in templates
├── examples/         # Example cards
└── docs/            # Documentation
```

## 🎨 Customization

### Project Templates

Create `.tcg-cardstyles/` in your project:

```
my-project/
├── .tcg-cardstyles/          # Project-specific templates
│   └── mtg/
│       └── custom.yaml       # Custom MTG cardstyle
├── cards/
│   └── my_card.md           # Your card files
└── output/                   # Generated images
```

### User Templates

Global templates in `$HOME/.tcg-cardgen/cardstyles/`:

```
~/.tcg-cardgen/
└── cardstyles/
    ├── mtg/
    │   └── my_style.yaml    # User MTG cardstyles
    └── custom_tcg/
        └── basic.yaml       # New TCG support
```

## 📖 Examples

### Lightning Bolt (Red Instant)
```yaml
card:
  tcg: mtg
  cardstyle: basic
mtg:
  color: red                 # Smart color frame selection
  cmc: 1
  mana_cost: ["{{mtg.mana_red}}"]
```

### Goblin Token
```yaml
card:
  tcg: mtg
  cardstyle: token          # Special token styling
mtg:
  color: red
  power: 1
  toughness: 1
```

### Legendary Artifact
```yaml
card:
  tcg: mtg
  cardstyle: legendary      # Legendary border
mtg:
  color: colorless
  cmc: 0
```

## 🛠️ Development

### Building from Source

```bash
# Install dependencies
go mod download

# Build
go build ./cmd/tcg-cardgen

# Run tests
go test ./...

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Adding New TCGs

1. Create templates in `templates/new_tcg/`
2. Define required/optional fields
3. Add validation rules
4. Test with example cards

See [Template Development Guide](docs/creating-templates.md) for details.

## 🤝 Contributing

Contributions are welcome! Please see our [Contributing Guide](docs/contributing.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and documentation
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [gg](https://github.com/fogleman/gg) for image generation
- Inspired by the Magic: The Gathering and Pokémon communities
- Uses Go's powerful template and embedding systems

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/Merith-TK/tcg-cardgen/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/Merith-TK/tcg-cardgen/discussions)
- 📧 **Contact**: [GitHub Profile](https://github.com/Merith-TK)

---

Made with ❤️ for the TCG community