---
card.tcg: "mtg"
card.title: "Tiny Text Spell"
card.type: "Instant"
card.rarity: "uncommon"
card.set: "Verbose Set"
card.artist: "Wordsmith"
card.print_this: 1
card.print_total: 1
mtg.mana_cost: "{{tcg.cost_blue}}"
mtg.cmc: 1
mtg.color_identity: ["blue"]
# Very small fonts for cards with lots of text
mtg.font_size.title: 24         # Much smaller title
mtg.font_size.card_text: 16     # Much smaller body text
mtg.font_size.type_line: 20     # Smaller type line
---

# Tiny Text Spell

This card has **incredibly verbose rules text** that requires smaller fonts to fit properly. It demonstrates how card designers can specify exact font sizes for their specific needs.

**Choose three different modes:**

* Target player draws a card, then discards a card
* Target creature gets +1/+1 until end of turn
* Target player gains 3 life
* Deal 1 damage to any target
* Counter target spell unless its controller pays {{tcg.cost_colorless(1)}}

**Additional rules:** If you control three or more artifacts, you may choose an additional mode. If you control five or more artifacts, you may choose all modes instead.

## Footer

*"Sometimes you need small text to fit big ideas."*