# TCG Card Generator Project Log

## Project Overview
- **Goal**: Create a markdown-based TCG card generator written in Go
- **Key Features**:
  - Cards defined by markdown files with metadata
  - Text-to-icon replacement system (e.g., `{{mtg.mana_black}}`, `mtg.mana_colorless(7)`)
  - Flexible card format determination via metadata

## Session Log

### 2025-10-26 - Initial Planning Session
- User requested creation of `.copilot` directory for agent bootstrapping
- Project scope: Markdown-based TCG card generator in Go
- Card format determined by metadata in markdown files
- Support for text2icon replacements with syntax like `{{mtg.mana_black}}` or `mtg.mana_colorless(7)`

#### Requirements Gathered:

**Output & Formats:**
- Primary: PNG (standard TCG size 2.5" x 3.5")
- Future: Multiple format support
- Output location: Same folder as markdown in `.tcg-cardgen-out/` subfolder

**Architecture:**
- Phase 1: CLI tool for batch/single file processing
- Phase 2: Web service with live preview and export
- Template system: Metadata-driven card layouts
- Asset storage: `$HOME/.tcg-cardgen/card-art` for local packs

**Icon System:**
- Syntax: `{{set.type}}` and `{{set.type(parameter)}}`
- Local filesystem + external URL fetching with caching
- PNG preferred, SVG supported
- Parameter support for dynamic icons (e.g., colorless mana numbers)

**Metadata Structure:**
- Header metadata determines card design/format
- Multi-TCG support in single card (mtg.power + pkm.hp)
- Generic `card.*` namespace for common properties
- `card.tcg` to specify target game
- **IMPORTANT**: `card.body` is the markdown content after frontmatter, NOT a metadata field

**Performance Requirements:**
- Memory efficient: Load assets as needed, cache during batch
- Unload card data after generation, keep reusable assets
- Support hundreds to thousands of cards

**Key Metadata Fields:**
- card.print_this/card.print_total (default 1/1)
- card.title (defaults to filename)
- card.set, card.lang, card.designer
- card.artwork (local/URL with caching)
- card.body.size, card.body.centered (for formatting the markdown content)
- Template-specific fields (mtg.power, pkm.hp, etc.)

#### Corrections Made:
- Fixed architecture example: `card.body` is markdown content, not frontmatter field
- Clarified: Markdown headers (# Title) should NOT be rendered in card body for MTG cards
- **NEW**: Reviewed GPT feedback on cardstyle format - excellent suggestions for universal base spec
- **NEW**: Need to refactor fields to be more TCG-specific (mtg.mana_cost vs card.mana_cost)
- **NEW**: Implement universal base cardstyle that other TCGs can extend
- **NEW**: Add style_tokens, layer roles, and conditional rendering
- **REFINEMENT**: Cardstyles should validate against their own TCG, not universal
- **REFINEMENT**: Cross-TCG icon support (mtg.mana_red works in pokemon cardstyle)
- **REFINEMENT**: Cardstyle-defined icon aliases (tcg.cost_red = mtg.mana_red = pkm.energy_red)

#### Architecture Refinements:
1. **TCG-Specific Validation**: Pokemon cardstyle requires pokemon metadata, not universal
2. **Cross-TCG Icon Support**: Any cardstyle can use any TCG's icons
3. **Icon Aliases**: Cardstyles define their own cross-references
4. **Streamlined Approach**: Less universal base, more focused TCG implementations

#### Implementation Plan:
1. ✅ Create basic project structure with sample cards and test cardstyle
2. ✅ Implement metadata parser  
3. ✅ Basic CLI interface
4. 🔄 Template system foundation (basic loading/validation complete)
5. 🆕 Refactor to universal base spec with TCG-specific extensions
6. 🆕 Update sample cards with proper TCG-specific field naming
7. 🆕 Implement layer roles and style tokens

#### Progress Log:

**2025-10-26 - Basic Structure Complete**
- Created full Go project structure with proper module layout
- Implemented YAML frontmatter parser with markdown body processing
- Created MTG template definition system
- Built working CLI tool with validation and basic generation framework
- Created 3 sample cards: Lightning Bolt, Serra Angel, Black Lotus
- Fixed template YAML syntax issue with optional_fields

**Current State:**
- ✅ CLI accepts files/directories and validates cards
- ✅ Metadata parsing strips markdown headers from card body
- ✅ Template loading and validation working
- ✅ Output path generation working
- 🔄 Next: Implement actual image rendering

**Test Results:**
```
.\tcg-cardgen.exe --validate-only ./examples/
✓ All 3 sample cards validate successfully

.\tcg-cardgen.exe --verbose ./examples/lightning_bolt.md  
✓ Parsing and generation pipeline working
```

## Next Steps
1. Design project structure and Go modules
2. Create metadata schema definitions
3. Implement asset management system
4. Build template engine
5. Create CLI interface