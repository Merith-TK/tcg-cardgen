# MTG Frame Assets

This directory contains the Magic: The Gathering card frame assets extracted from the sprite sheet.

## Frame Sprite Sheet Layout

**Row 1 (Special Frames):**
- Column 1: Colorless frame + mana symbol sprite sheet
- Column 2: Legendary frame  
- Column 3: Green frame
- Column 4: Red frame
- Column 5: Black frame
- Column 6: Blue frame
- Column 7: White frame

**Row 2 (Basic Frames):**
- Column 1: Card back
- Columns 2-8: Basic color frames corresponding to Row 1

## Usage

The cardstyles reference these frames using:
```yaml
source: "{{template_dir}}/frames/white_frame.png"
```

Individual frame files should be extracted from the sprite sheet and saved here.

## Icon Sprites

The top-left frame contains mana symbol sprites that can be used for icon replacement:
- Numbers 0-9 for colorless mana
- WUBRG mana symbols
- Tap/untap symbols
- Special symbols