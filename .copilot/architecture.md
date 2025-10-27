# TCG Card Generator Architecture

## Project Structure
```
tcg-cardgen/
├── cmd/
│   └── tcg-cardgen/          # CLI entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── metadata/             # Markdown metadata parsing
│   ├── templates/            # Template engine and management
│   ├── assets/               # Asset loading and caching
│   ├── icons/                # Icon replacement system
│   ├── renderer/             # Image generation and rendering
│   └── cache/                # Asset caching system
├── pkg/
│   └── cardgen/              # Public API for library usage
├── templates/                # Built-in card templates
│   └── mtg/                  # MTG template definitions
├── examples/                 # Example cards and usage
└── docs/                     # Documentation
```

## Core Components

### 1. Metadata Parser (`internal/metadata/`)
- Parse YAML/TOML frontmatter from markdown
- Validate required fields per template
- Handle multi-TCG metadata (mtg.*, pkm.*, card.*)
- Default value resolution (filename → card.title)

### 2. Template Engine (`internal/templates/`)
- Template definition format (JSON/YAML)
- Layout region mapping (title, body, power/toughness)
- Style-specific field validation
- Template discovery and loading

### 3. Asset Manager (`internal/assets/`)
- Local asset loading from `$HOME/.tcg-cardgen/card-art`
- URL fetching with local caching
- Memory-efficient asset lifecycle
- PNG/SVG support

### 4. Icon System (`internal/icons/`)
- `{{set.type}}` and `{{set.type(parameter)}}` parsing
- Dynamic icon generation (colorless mana numbers)
- Icon pack management
- Template integration

### 5. Renderer (`internal/renderer/`)
- PNG output generation
- Layer composition
- Text rendering with positioning
- Asset overlaying

### 6. Cache System (`internal/cache/`)
- Asset memory management
- Batch processing optimization
- Cache invalidation strategies

## Data Flow

1. **Input**: Markdown file with frontmatter
2. **Parse**: Extract metadata and content
3. **Validate**: Check against template requirements
4. **Load**: Fetch template and required assets
5. **Process**: Replace icons, render text
6. **Compose**: Layer assets according to template
7. **Output**: Generate PNG to `.tcg-cardgen-out/`
8. **Cleanup**: Unload card-specific data, keep reusable assets

## Template Definition Format

```yaml
# templates/mtg/basic.yaml
name: "MTG Basic Card"
tcg: "mtg"
version: "1.0"
dimensions:
  width: 750   # 2.5" at 300 DPI
  height: 1050 # 3.5" at 300 DPI

layers:
  - name: "background"
    type: "image"
    source: "{{template_dir}}/backgrounds/{{card.rarity}}.png"
    
  - name: "artwork"
    type: "image"
    source: "{{card.artwork}}"
    region: { x: 60, y: 100, width: 630, height: 460 }
    
  - name: "title"
    type: "text"
    content: "{{card.title}}"
    region: { x: 60, y: 60, width: 500, height: 40 }
    font: { family: "MTG-Font", size: 24, color: "#000000" }
    
  - name: "mana_cost"
    type: "text"
    content: "{{card.mana_cost}}"
    region: { x: 600, y: 60, width: 90, height: 40 }
    icon_replace: true
    
  - name: "body"
    type: "text"
    content: "{{card.body}}"
    region: { x: 60, y: 600, width: 630, height: 300 }
    font: { size: "{{card.body.size|16}}", centered: "{{card.body.centered|false}}" }
    icon_replace: true

required_fields:
  - card.title
  - card.mana_cost
  - card.body
  
optional_fields:
  - card.artwork
  - mtg.power
  - mtg.toughness
  - card.rarity
```

## CLI Interface Design

```bash
# Single file
tcg-cardgen ./cards/lightning_bolt.md

# Batch directory
tcg-cardgen ./cards/

# With options
tcg-cardgen ./cards/ --template-dir ./custom-templates --output-dir ./output --verbose

# List available templates
tcg-cardgen --list-templates

# Validate without generating
tcg-cardgen ./cards/ --validate-only
```

## Icon Replacement Examples

```markdown
---
card.title: "Lightning Bolt"
card.mana_cost: "{{mtg.mana_red}}"
mtg.power: "{{mtg.power_undefined}}"
mtg.toughness: "{{mtg.toughness_undefined}}"
---

# Lightning Bolt
Lightning Bolt deals 3 damage to any target. 

{{mtg.mana_red}}{{mtg.mana_red}}: Draw a card.
```

The `card.body` would be the entire markdown content after the frontmatter, with icon replacements processed during rendering. This would render mana symbols and power/toughness appropriately in the MTG template.