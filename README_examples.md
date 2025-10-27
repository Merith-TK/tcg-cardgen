# MTG Card Examples

This directory contains example cards demonstrating the 3 essential MTG cardstyles with smart color affinity.

## Cardstyle Examples

### 🎴 **Basic Cards** (`cardstyle: basic`)
Standard MTG cards with smart color frame selection:

- **Lightning Bolt** (`mtg.color: red`) - Red instant spell
- **Counterspell** (`mtg.color: blue`) - Blue instant spell  
- **Sol Ring** (`mtg.color: colorless`) - Colorless artifact

### 🪙 **Token Cards** (`cardstyle: token`)
Token creatures with special styling:

- **Goblin Token** (`mtg.color: red`) - Red token creature

### 👑 **Legendary Cards** (`cardstyle: legendary`)
Legendary permanents with special border:

- **Mox Emerald** (`mtg.color: green`) - Legendary green artifact

## Key Features Demonstrated

### 🎨 **Color Affinity System**
All cards use `mtg.color` to automatically select the appropriate frame:
```yaml
mtg:
  color: red        # Uses red_frame.png
  color: blue       # Uses blue_frame.png
  color: colorless  # Uses colorless_frame.png
  # etc.
```

### 📋 **TCG-Specific Metadata**
Cards use proper MTG metadata structure:
```yaml
card:
  tcg: mtg                    # Specifies Magic: The Gathering
  cardstyle: basic            # Chooses the cardstyle

mtg:
  cmc: 1                      # Converted mana cost
  color: red                  # Color affinity
  type_line: Instant          # Full type line
  mana_cost: ["{{mtg.mana_red}}"]  # Mana cost with icons
  power: 2                    # Creature power (if applicable)
  toughness: 2                # Creature toughness (if applicable)
```

## Generation

Generate all examples:
```bash
tcg-cardgen ./examples/
```

Generate single card:
```bash
tcg-cardgen ./examples/lightning_bolt_red.md
```

List available cardstyles:
```bash
tcg-cardgen --list-templates
```