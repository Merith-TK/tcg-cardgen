# Creating Cards

Learn how to create beautiful TCG cards using the card generator. Cards are written as Markdown files with YAML frontmatter containing the card's metadata.

## 📝 Basic Card Structure

Every card file follows this structure:

```markdown
---
# YAML frontmatter with card metadata
card:
  tcg: mtg                    # Which TCG (mtg, pokemon, etc.)
  cardstyle: basic            # Which template style
  title: "Card Name"          # Card title
  
# TCG-specific fields
mtg:
  color: red                  # Smart color selection
  cmc: 3                      # Converted mana cost
  mana_cost: ["{{mtg.mana_red}}", "{{mtg.mana_red}}", "{{mtg.mana_red}}"]
---

# Card Title

**Card text goes here** with *formatting*.

*Flavor text is italic and appears at the bottom.*
```

## 🎮 TCG-Specific Formats

### Magic: The Gathering (MTG)

#### Required Fields
```yaml
card:
  tcg: mtg                    # Must be "mtg"
  cardstyle: basic            # basic | token | legendary
  title: "Lightning Bolt"     # Card name

mtg:
  color: red                  # Color affinity for frame selection
  type_line: "Instant"        # Full type line
```

#### Optional Fields
```yaml
mtg:
  cmc: 1                      # Converted mana cost
  mana_cost: ["{{mtg.mana_red}}"]  # Mana symbols
  power: 2                    # Creature power
  toughness: 2                # Creature toughness
  rarity: "rare"              # Card rarity
  set: "LEA"                  # Set code
  artist: "Mark Poole"        # Artist name
```

#### Color Options
- `red` - Red frame and styling
- `blue` - Blue frame and styling  
- `black` - Black frame and styling
- `white` - White frame and styling
- `green` - Green frame and styling
- `colorless` - Colorless/artifact frame

#### Cardstyle Options
- `basic` - Standard MTG cards (spells, creatures, artifacts)
- `token` - Token creatures with special border
- `legendary` - Legendary permanents with unique styling

### Pokémon (PKM)

#### Required Fields
```yaml
card:
  tcg: pokemon               # Must be "pokemon"
  cardstyle: basic           # Currently only "basic"
  title: "Pikachu"           # Pokémon name

pkm:
  hp: 60                     # Hit points
  type: "Lightning"          # Pokémon type
```

#### Optional Fields
```yaml
pkm:
  stage: "Basic"             # Basic | Stage 1 | Stage 2
  evolves_from: "Pichu"      # Previous evolution
  weakness: "Fighting"       # Weakness type
  resistance: "Metal"        # Resistance type
  retreat_cost: 1            # Retreat cost
  attacks: []                # Attack list
  rarity: "common"           # Card rarity
```

## ✍️ Writing Card Text

### Basic Formatting

```markdown
# Card Title

**Bold text** for card names and keywords.

*Italic text* for flavor text and emphasis.

Regular text for rules text.
```

### Rules Text Patterns

#### Magic: The Gathering
```markdown
**Lightning Bolt** deals 3 damage to any target.

**Counterspell** counters target spell.

When **Goblin Token** enters the battlefield, create a 1/1 red Goblin creature token.
```

#### Pokémon
```markdown
**Thunder Shock** - 20 damage
Flip a coin. If heads, the Defending Pokémon is now Paralyzed.

**Agility** - 30 damage  
Flip a coin. If heads, prevent all effects of attacks, including damage, done to **Pikachu** during your opponent's next turn.
```

### Flavor Text
```markdown
# Card rules text here...

*"The spark that ignites the flame of victory."*  
*—Chandra Nalaar*
```

## 🔧 Advanced Features

### Mana Symbols (MTG)
Use template variables for mana symbols:

```yaml
mtg:
  mana_cost: 
    - "{{mtg.mana_red}}"      # Red mana
    - "{{mtg.mana_blue}}"     # Blue mana  
    - "{{mtg.mana_white}}"    # White mana
    - "{{mtg.mana_black}}"    # Black mana
    - "{{mtg.mana_green}}"    # Green mana
    - "{{mtg.mana_colorless}}" # Generic mana
```

### Power/Toughness (MTG)
```yaml
mtg:
  power: 2                   # Creature power
  toughness: 3               # Creature toughness
```

### Artwork
```yaml
card:
  artwork: "path/to/image.png"  # Local file
  # OR
  artwork: "https://example.com/image.png"  # URL
```

### Advanced Metadata
```yaml
card:
  set: "MY_SET"              # Set code
  rarity: "mythic"           # Rarity
  artist: "Artist Name"      # Artist credit
  print_this: 1              # Collector number
  print_total: 100           # Total in set
```

## 📋 Complete Examples

### Lightning Bolt (MTG Instant)
```markdown
---
card:
  tcg: mtg
  cardstyle: basic
  title: "Lightning Bolt"
  type: "Instant"
  rarity: "common"
  set: "LEA"
  artist: "Christopher Rush"

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

### Goblin Token (MTG Token)
```markdown
---
card:
  tcg: mtg
  cardstyle: token
  title: "Goblin Token"
  type: "Token Creature"

mtg:
  cmc: 0
  color: red
  type_line: "Creature — Goblin"
  power: 1
  toughness: 1
---

# Goblin Token

A 1/1 red Goblin creature token.

*"They may be small, but they're surprisingly vicious."*
```

### Mox Emerald (MTG Legendary)
```markdown
---
card:
  tcg: mtg
  cardstyle: legendary
  title: "Mox Emerald"
  type: "Legendary Artifact"
  rarity: "mythic"

mtg:
  cmc: 0
  color: green
  type_line: "Legendary Artifact"
---

# Mox Emerald

**{{T}}**: Add **{{G}}** to your mana pool.

*One of the most powerful artifacts ever created, sought after by planeswalkers across the multiverse.*
```

### Pikachu (Pokémon)
```markdown
---
card:
  tcg: pokemon
  cardstyle: basic
  title: "Pikachu"
  rarity: "common"

pkm:
  hp: 60
  type: "Lightning"
  stage: "Basic"
  weakness: "Fighting"
  retreat_cost: 1
---

# Pikachu

**Thunder Shock** - 20  
Flip a coin. If heads, the Defending Pokémon is now Paralyzed.

**Agility** - 30  
Flip a coin. If heads, prevent all effects of attacks, including damage, done to **Pikachu** during your opponent's next turn.

*When several of these Pokémon gather, their electricity could build and cause lightning storms.*
```

## 🎯 Best Practices

### 1. **Consistent Naming**
- Use clear, descriptive filenames: `lightning_bolt_red.md`
- Match card titles exactly: `title: "Lightning Bolt"`

### 2. **Proper Metadata**
- Always specify `tcg` and `cardstyle`
- Use appropriate color affinity for frames
- Include required fields for your TCG

### 3. **Clear Text Formatting**
- **Bold** for card names and keywords
- *Italic* for flavor text
- Regular text for rules text

### 4. **Organize Your Cards**
```
my-project/
├── cards/
│   ├── creatures/
│   │   ├── goblin_warrior.md
│   │   └── lightning_elemental.md
│   ├── spells/
│   │   ├── lightning_bolt.md
│   │   └── counterspell.md
│   └── artifacts/
│       └── mox_emerald.md
└── output/
    └── .tcg-cardgen-out/
```

### 5. **Test Early and Often**
```bash
# Validate without generating
tcg-cardgen --validate-only cards/

# Generate and check output
tcg-cardgen cards/new_card.md
```

## ❌ Common Mistakes

### Missing Required Fields
```yaml
# ❌ Wrong - missing required fields
card:
  title: "My Card"

# ✅ Correct - includes TCG and cardstyle
card:
  tcg: mtg
  cardstyle: basic
  title: "My Card"
```

### Incorrect Color Values
```yaml
# ❌ Wrong - invalid color
mtg:
  color: purple

# ✅ Correct - valid MTG color
mtg:
  color: red
```

### Malformed YAML
```yaml
# ❌ Wrong - inconsistent indentation
card:
  tcg: mtg
    cardstyle: basic

# ✅ Correct - proper indentation
card:
  tcg: mtg
  cardstyle: basic
```

## 🔍 Validation and Debugging

### Validate Cards
```bash
# Check all cards for errors
tcg-cardgen --validate-only examples/

# Verbose output for debugging
tcg-cardgen --verbose --validate-only my_card.md
```

### Common Error Messages

**"Required field missing"**
- Add the missing field to your YAML frontmatter

**"Invalid TCG"**
- Check that `card.tcg` is a supported value (`mtg`, `pokemon`)

**"Template not found"**
- Verify `card.cardstyle` exists for your TCG
- Use `--list-templates` to see available options

**"Invalid color"**
- Use valid color values for your TCG (e.g., `red`, `blue`, `colorless` for MTG)

## 🎨 Next Steps

- **[Creating Templates](creating-templates.md)** - Build custom cardstyles
- **[Examples](examples.md)** - Explore more example cards  
- **[API Reference](api.md)** - Use the generator programmatically

Happy card creation! 🎴