package card_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Merith-TK/tcg-cardgen/pkg/card"
)

func writeCard(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.md")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestParseBasicFrontmatter(t *testing.T) {
	path := writeCard(t, `---
tcg: mtg
cardstyle: basic
title: Lightning Bolt
rarity: rare
set: Alpha
artist: Christopher Rush
---

Deal 3 damage to any target.
`)
	c, err := card.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if c.TCG != "mtg" {
		t.Errorf("TCG: want mtg, got %q", c.TCG)
	}
	if c.CardStyle != "basic" {
		t.Errorf("CardStyle: want basic, got %q", c.CardStyle)
	}
	if c.Title != "Lightning Bolt" {
		t.Errorf("Title: want 'Lightning Bolt', got %q", c.Title)
	}
	if c.Rarity != "rare" {
		t.Errorf("Rarity: want rare, got %q", c.Rarity)
	}
	if c.Artist != "Christopher Rush" {
		t.Errorf("Artist: %q", c.Artist)
	}
	if c.RulesText != "Deal 3 damage to any target." {
		t.Errorf("RulesText: %q", c.RulesText)
	}
}

func TestParseBodyTitle(t *testing.T) {
	path := writeCard(t, `# Counterspell

Counter target spell.
`)
	c, err := card.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Counterspell" {
		t.Errorf("Title: want 'Counterspell', got %q", c.Title)
	}
	if c.RulesText != "Counter target spell." {
		t.Errorf("RulesText: %q", c.RulesText)
	}
}

func TestParseFlavorText(t *testing.T) {
	path := writeCard(t, `---
tcg: mtg
cardstyle: basic
---

Deal 3 damage.

---

*A simple bolt, yet devastating.*
`)
	c, err := card.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.RulesText != "Deal 3 damage." {
		t.Errorf("RulesText: %q", c.RulesText)
	}
	if c.FlavorText != "A simple bolt, yet devastating." {
		t.Errorf("FlavorText: %q", c.FlavorText)
	}
}

func TestParseManaCost(t *testing.T) {
	path := writeCard(t, `---
tcg: mtg
cardstyle: basic
---

> {{mtg.cost}}

> **Instant**

Deal 3 damage.
`)
	c, err := card.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.ManaCost != "{{mtg.cost}}" {
		t.Errorf("ManaCost: want '{{mtg.cost}}', got %q", c.ManaCost)
	}
	if c.Type != "Instant" {
		t.Errorf("Type: want 'Instant', got %q", c.Type)
	}
}

func TestParseDefaultsFromFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my_cool_card.md")
	os.WriteFile(path, []byte("Some rules text.\n"), 0644)

	c, err := card.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "My cool card" {
		t.Errorf("Title default: got %q", c.Title)
	}
	if c.TCG != "mtg" {
		t.Errorf("TCG default: got %q", c.TCG)
	}
	if c.CardStyle != "basic" {
		t.Errorf("CardStyle default: got %q", c.CardStyle)
	}
}

func TestParseNestedFrontmatter(t *testing.T) {
	path := writeCard(t, `---
tcg: mtg
cardstyle: basic
title: Test Card
mtg:
  color: red
  cmc: 3
  power: 2
  toughness: 1
---
`)
	c, err := card.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Nested keys should be in Metadata and accessible by their section.
	mtgMap, ok := c.Metadata["mtg"].(map[string]interface{})
	if !ok {
		t.Fatal("expected mtg to be a map in Metadata")
	}
	if color, _ := mtgMap["color"].(string); color != "red" {
		t.Errorf("mtg.color: want red, got %q", color)
	}
}
