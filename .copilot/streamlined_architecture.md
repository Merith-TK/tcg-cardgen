# Streamlined TCG Card Generator Architecture v2

## Core Principles

1. **TCG-Specific Validation**: Each cardstyle validates against its own TCG metadata
2. **Cross-TCG Icon Support**: Any cardstyle can reference any TCG's icons  
3. **Cardstyle-Defined Aliases**: Templates define their own cross-references
4. **User Freedom**: Cards can mix TCG elements while using appropriate templates

## Template Structure

### Basic Template Format
```yaml
name: "Template Name"
tcg: "mtg"                    # Template validates against this TCG
version: "1.0.0"
description: "Description"

required_fields:              # TCG-specific required fields
  - card.tcg                  # Must match template TCG
  - card.title
  - mtg.mana_cost            # TCG-specific fields

optional_fields:              # Defaults for missing fields
  mtg.power: null
  card.rarity: "common"

layers: [...]                 # Rendering layers

style_tokens:                 # Visual constants
  font_title: "Beleren"
  color_text: "#000000"

icons:                        # Icon definitions with cross-TCG support
  # Native icons
  mtg.mana_red: "path/to/red.png"
  
  # Cross-TCG icons (allows other TCGs in this template)
  pkm.energy_fire: "path/to/pokemon/fire.png"
  
  # Template-defined aliases
  tcg.cost_red: "{{mtg.mana_red}}"     # This template's red cost
  tcg.cost: "{{mtg.mana_colorless}}"   # Generic cost
```

## Validation Rules

1. **TCG Matching**: `card.tcg` must match template `tcg` field
2. **Required Fields**: All template `required_fields` must exist in card
3. **Field Namespacing**: MTG cards use `mtg.*`, Pokemon cards use `pkm.*`
4. **Icon Freedom**: Any template can reference any TCG's icons

## Example Cross-TCG Usage

### MTG Card Using Generic Aliases
```yaml
mtg.mana_cost: "{{tcg.cost_red}}{{tcg.cost_colorless(1)}}"
```
In MTG template: `tcg.cost_red` → `mtg.mana_red`
In Pokemon template: `tcg.cost_red` → `pkm.energy_fire`

### Pokemon Card Using MTG Icons for Custom Effects  
```yaml
# This works in a Pokemon cardstyle
card.body: "Discard a {{mtg.mana_red}} energy..."
```

### Mixed Content Card
```yaml
# MTG card with Pokemon reference (using MTG template)
card.body: "This creature has the energy of {{pkm.energy_electric}}."
```

## Benefits

1. **Strict Validation**: Wrong TCG cards fail template validation
2. **Creative Freedom**: Users can mix elements for custom designs
3. **Template Consistency**: Each TCG has its own visual style
4. **Icon Compatibility**: Cross-references work seamlessly
5. **Extensible**: New TCGs add their own templates and icons

## Implementation Status

- ✅ TCG-specific required fields
- ✅ Cross-TCG icon definitions in templates  
- ✅ Style tokens system
- ✅ Layer roles for semantic meaning
- 🔄 Template validation enforcement (needs debugging)
- 🔄 Icon alias resolution system
- ⏳ Template inheritance system (future)

This approach gives maximum flexibility while maintaining proper validation and visual consistency per TCG.