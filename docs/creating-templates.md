# Creating Templates

Learn how to create custom cardstyles (templates) for the TCG card generator. Templates define the visual layout, styling, and validation rules for cards.

## 🎨 Template Basics

Templates are YAML files that define:
- **Layout regions** - Where text and images appear
- **Visual styling** - Fonts, colors, frames
- **Validation rules** - Required/optional fields
- **Smart features** - Dynamic content based on card properties

## 📁 Template Organization

### Built-in Templates
```
templates/
├── mtg/
│   ├── basic.yaml       # Standard MTG cards
│   ├── token.yaml       # Token creatures  
│   └── legendary.yaml   # Legendary cards
└── pokemon/
    └── basic.yaml       # Basic Pokémon cards
```

### User Templates
```
$HOME/.tcg-cardgen/cardstyles/
├── mtg/
│   └── my_style.yaml    # User MTG templates
└── custom_tcg/
    └── basic.yaml       # New TCG support
```

### Project Templates
```
.tcg-cardstyles/
├── mtg/
│   └── project_style.yaml  # Project-specific overrides
└── custom/
    └── special.yaml         # Project-only templates
```

## 🏗️ Template Structure

### Basic Template Format
```yaml
name: "My Card Style"
tcg: "mtg"
version: "1.0.0"
description: "Custom MTG cardstyle with special features"

# Card dimensions
dimensions:
  width: 750
  height: 1050
  dpi: 300

# Required fields for validation
required_fields:
  - card.tcg
  - card.title
  - mtg.color

# Optional fields with defaults
optional_fields:
  mtg.cmc: 0
  card.rarity: "common"

# Visual layers (rendered in order)
layers:
  - name: "background"
    type: "image"
    path: "frames/{{mtg.color|colorless}}_frame.png"
    region: { x: 0, y: 0, width: 750, height: 1050 }
    
  - name: "title"
    type: "text"
    content: "{{card.title}}"
    region: { x: 50, y: 50, width: 650, height: 60 }
    font:
      family: "Arial"
      size: 24
      weight: "bold"
      color: "#000000"
```

## 🧩 Template Inheritance

Templates can extend other templates to reduce duplication:

### Base Template (`mtg/base.yaml`)
```yaml
name: "MTG Base Template"
tcg: "mtg"
version: "1.0.0"

dimensions:
  width: 750
  height: 1050
  dpi: 300

required_fields:
  - card.tcg
  - card.title
  - mtg.color

# Common layers
layers:
  - name: "background"
    type: "image"  
    path: "frames/{{mtg.color|colorless}}_frame.png"
    region: { x: 0, y: 0, width: 750, height: 1050 }
    
  - name: "title"
    type: "text"
    content: "{{card.title}}"
    region: { x: 50, y: 60, width: 650, height: 40 }
    font:
      family: "Arial"
      size: 22
      weight: "bold"
      color: "#000000"
```

### Extended Template (`mtg/legendary.yaml`)
```yaml
name: "MTG Legendary Card"
extends: "./base.yaml"          # Inherit from base template
tcg: "mtg"
version: "1.0.0"
description: "Legendary MTG cards with special border"

# Override layers
overrides:
  - layer: "background"
    path: "frames/legendary_{{mtg.color|colorless}}_frame.png"

# Add new layers
additional_layers:
  - name: "legendary_crown"
    type: "image"
    path: "overlays/legendary_crown.png"
    region: { x: 300, y: 20, width: 150, height: 60 }
```

## 🎯 Layer Types

### Image Layers
```yaml
- name: "background"
  type: "image"
  path: "frames/blue_frame.png"     # Static path
  # OR
  path: "frames/{{mtg.color}}_frame.png"  # Dynamic path
  region: { x: 0, y: 0, width: 750, height: 1050 }
  fit: "stretch"                    # stretch | contain | cover
```

### Text Layers
```yaml
- name: "title"
  type: "text"
  content: "{{card.title}}"
  region: { x: 50, y: 60, width: 650, height: 40 }
  font:
    family: "Arial"                 # Font family
    size: 24                        # Fixed size
    # OR
    size: "{{mtg.font_size.title}}" # Dynamic size
    weight: "bold"                  # normal | bold
    style: "normal"                 # normal | italic
    color: "#000000"                # Hex color
  align: "center"                   # left | center | right
  valign: "middle"                  # top | middle | bottom
  condition: "{{card.title}}"       # Only render if condition is true
  icon_replace: true                # Process icon replacements
```

## 🔤 Template Variables

### Card Variables
```yaml
content: "{{card.title}}"          # Card title
content: "{{card.type}}"           # Card type
content: "{{card.body}}"           # Rules text
content: "{{card.footer}}"         # Flavor text
content: "{{card.artist}}"         # Artist name
content: "{{card.set}}"            # Set code
content: "{{card.rarity}}"         # Rarity
```

### TCG-Specific Variables (MTG)
```yaml
content: "{{mtg.cmc}}"             # Converted mana cost
content: "{{mtg.color}}"           # Color (red, blue, etc.)
content: "{{mtg.type_line}}"       # Full type line
content: "{{mtg.power}}"           # Creature power
content: "{{mtg.toughness}}"       # Creature toughness
content: "{{mtg.mana_cost}}"       # Mana cost array
```

### TCG-Specific Variables (Pokémon)
```yaml
content: "{{pkm.hp}}"              # Hit points
content: "{{pkm.type}}"            # Pokémon type
content: "{{pkm.stage}}"           # Evolution stage
content: "{{pkm.weakness}}"        # Weakness type
content: "{{pkm.resistance}}"      # Resistance type
```

## 🎨 Smart Features

### Dynamic Paths with Fallbacks
```yaml
path: "frames/{{mtg.color|colorless}}_frame.png"
```
- Uses `red_frame.png` if `mtg.color` is `red`
- Falls back to `colorless_frame.png` if `mtg.color` is empty

### Conditional Rendering
```yaml
- name: "power_toughness"
  type: "text"
  content: "{{mtg.power}}/{{mtg.toughness}}"
  condition: "{{mtg.power}} && {{mtg.toughness}}"  # Only show for creatures
  region: { x: 600, y: 950, width: 100, height: 40 }
```

### Style Tokens
```yaml
style_tokens:
  font_large: "Arial"
  font_small: "Arial"
  color_title: "#000000"
  color_body: "#333333"

layers:
  - name: "title"
    type: "text"
    font:
      family: "{{style_tokens.font_large}}"
      color: "{{style_tokens.color_title}}"
```

## 🔧 Advanced Features

### Icon Replacement
```yaml
icons:
  mtg.mana_red: "icons/mana_red.png"
  mtg.mana_blue: "icons/mana_blue.png"
  mtg.mana_tap: "icons/tap.png"

layers:
  - name: "rules_text"
    type: "text"
    content: "{{card.body}}"
    icon_replace: true              # Replace {{mtg.mana_red}} with icon
```

### Layer Overrides
```yaml
# In extending template
overrides:
  - layer: "background"             # Override this layer
    path: "special_background.png"  # New background
    tint: "#FF0000"                 # Add red tint
    
  - layer: "title"
    font:
      size: 28                      # Larger title font
      color: "#FFFFFF"              # White title text
```

### Multiple Conditions
```yaml
- name: "planeswalker_loyalty"
  type: "text"
  content: "{{mtg.loyalty}}"
  condition: "{{mtg.loyalty}} && {{mtg.type_line|contains:Planeswalker}}"
  region: { x: 650, y: 950, width: 80, height: 80 }
```

## 📐 Layout Guidelines

### Standard MTG Dimensions
```yaml
dimensions:
  width: 750                        # 2.5 inches at 300 DPI
  height: 1050                      # 3.5 inches at 300 DPI  
  dpi: 300                          # Print quality
```

### Common Regions
```yaml
# Title area
title_region: { x: 50, y: 60, width: 650, height: 40 }

# Mana cost (top right)
mana_region: { x: 620, y: 60, width: 80, height: 40 }

# Main text area
body_region: { x: 50, y: 500, width: 650, height: 300 }

# Power/Toughness (bottom right)
pt_region: { x: 600, y: 950, width: 100, height: 80 }

# Artist credit (bottom left)
artist_region: { x: 50, y: 980, width: 300, height: 20 }
```

## 🎮 Adding New TCGs

### 1. Create TCG Directory
```bash
mkdir templates/my_tcg
```

### 2. Define Base Template
```yaml
# templates/my_tcg/basic.yaml
name: "My TCG Basic Card"
tcg: "my_tcg"
version: "1.0.0"
description: "Basic card template for My TCG"

dimensions:
  width: 800
  height: 1200
  dpi: 300

required_fields:
  - card.tcg
  - card.title
  - my_tcg.level              # TCG-specific required field

optional_fields:
  my_tcg.level: 1
  my_tcg.element: "neutral"
  card.rarity: "common"

layers:
  - name: "background"
    type: "image"
    path: "frames/{{my_tcg.element|neutral}}_frame.png"
    region: { x: 0, y: 0, width: 800, height: 1200 }
    
  - name: "title"
    type: "text"
    content: "{{card.title}}"
    region: { x: 60, y: 70, width: 680, height: 50 }
    font:
      family: "Arial"
      size: 26
      weight: "bold"
      color: "#000000"
      
  - name: "level"
    type: "text"
    content: "Level {{my_tcg.level}}"
    region: { x: 650, y: 70, width: 90, height: 30 }
    font:
      family: "Arial"
      size: 18
      color: "#FFFFFF"
```

### 3. Create Card Example
```markdown
---
card:
  tcg: my_tcg
  cardstyle: basic
  title: "Fire Dragon"
  
my_tcg:
  level: 5
  element: fire
  attack: 1200
  defense: 800
---

# Fire Dragon

A powerful dragon that breathes fire and soars through the skies.

**Flame Breath**: Deal 800 damage to target.
```

## 🔍 Testing Templates

### Validate Template
```bash
# Test with example card
tcg-cardgen --validate-only examples/my_card.md

# Verbose output for debugging
tcg-cardgen --verbose examples/my_card.md
```

### Template Discovery
```bash
# List all available templates
tcg-cardgen --list-templates

# Should show your new template:
# 🎮 MY_TCG:
#   📄 my_tcg/basic (My TCG Basic Card)
#      Basic card template for My TCG
#      Source: templates/my_tcg/basic.yaml
```

## ❌ Common Issues

### Path Resolution
```yaml
# ❌ Wrong - absolute paths
path: "/Users/me/frames/red.png"

# ✅ Correct - relative to template directory  
path: "frames/red_frame.png"
```

### Missing Variables
```yaml
# ❌ Wrong - undefined variable
content: "{{undefined_var}}"

# ✅ Correct - defined variable or fallback
content: "{{mtg.power|0}}"
```

### Invalid YAML
```yaml
# ❌ Wrong - inconsistent indentation
layers:
  - name: "title"
  type: "text"

# ✅ Correct - proper indentation
layers:
  - name: "title"
    type: "text"
```

## 🎯 Best Practices

### 1. **Use Template Inheritance**
- Create base templates for common layouts
- Extend bases for specialized variants
- Reduce code duplication

### 2. **Smart Defaults**
```yaml
optional_fields:
  mtg.cmc: 0                        # Default mana cost
  mtg.color: "colorless"            # Default color
  card.rarity: "common"             # Default rarity
```

### 3. **Flexible Paths**
```yaml
# Support multiple frame types
path: "frames/{{mtg.color|colorless}}_{{card.rarity|common}}_frame.png"
```

### 4. **Clear Naming**
- Use descriptive layer names: `"mana_cost"` not `"layer1"`
- Consistent file naming: `red_frame.png`, `blue_frame.png`
- Logical template names: `basic.yaml`, `token.yaml`

### 5. **Validation**
```yaml
# Require essential fields
required_fields:
  - card.tcg
  - card.title
  - mtg.color

# Provide sensible defaults
optional_fields:
  mtg.cmc: 0
  card.artist: "Unknown Artist"
```

## 📚 Next Steps

- **[Card Creation Guide](creating-cards.md)** - Learn to write cards for your templates
- **[Examples](examples.md)** - See template examples in action
- **[API Reference](api.md)** - Use templates programmatically

Happy template creation! 🎨