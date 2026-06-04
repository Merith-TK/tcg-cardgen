package renderer

import (
	"image/color"
	"testing"
)

func TestParseColorRGB(t *testing.T) {
	c, err := parseColor("#ff0000")
	if err != nil {
		t.Fatal(err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if c != want {
		t.Errorf("want %v, got %v", want, c)
	}
}

func TestParseColorRGBA8(t *testing.T) {
	c, err := parseColor("#ff000080")
	if err != nil {
		t.Fatal(err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 128}
	if c != want {
		t.Errorf("want %v, got %v", want, c)
	}
}

func TestParseColorShort3(t *testing.T) {
	c, err := parseColor("#fff")
	if err != nil {
		t.Fatal(err)
	}
	want := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if c != want {
		t.Errorf("want %v, got %v", want, c)
	}
}

func TestParseColorShort4(t *testing.T) {
	c, err := parseColor("#f00f")
	if err != nil {
		t.Fatal(err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if c != want {
		t.Errorf("want %v, got %v", want, c)
	}
}

func TestParseColorInvalid(t *testing.T) {
	cases := []string{"", "red", "#gg0000", "#12345"}
	for _, tc := range cases {
		_, err := parseColor(tc)
		if err == nil {
			t.Errorf("expected error for %q", tc)
		}
	}
}

func TestParseInlineBoldItalic(t *testing.T) {
	segs := parseInline("***bold italic***")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if !segs[0].Style.Bold || !segs[0].Style.Italic {
		t.Errorf("expected bold+italic, got %+v", segs[0].Style)
	}
	if segs[0].Content != "bold italic" {
		t.Errorf("content: %q", segs[0].Content)
	}
}

func TestParseInlineBold(t *testing.T) {
	segs := parseInline("**bold** normal")
	if len(segs) < 2 {
		t.Fatalf("expected ≥2 segments, got %d: %v", len(segs), segs)
	}
	if !segs[0].Style.Bold {
		t.Errorf("first segment should be bold: %+v", segs[0].Style)
	}
	if segs[0].Content != "bold" {
		t.Errorf("first content: %q", segs[0].Content)
	}
}

func TestParseInlineItalic(t *testing.T) {
	segs := parseInline("*italic*")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if !segs[0].Style.Italic || segs[0].Style.Bold {
		t.Errorf("expected italic-only, got %+v", segs[0].Style)
	}
}

func TestParseInlineMixed(t *testing.T) {
	segs := parseInline("normal **bold** and *italic*")
	// Expect: "normal ", "bold", " and ", "italic"
	if len(segs) != 4 {
		t.Fatalf("expected 4 segments, got %d: %v", len(segs), segs)
	}
}

func TestParseInlinePlain(t *testing.T) {
	segs := parseInline("no formatting here")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if segs[0].Style.Bold || segs[0].Style.Italic {
		t.Error("expected no formatting")
	}
}

func TestIsHR(t *testing.T) {
	if !isHR("---") {
		t.Error("--- should be HR")
	}
	if !isHR("***") {
		t.Error("*** should be HR")
	}
	if isHR("**text**") {
		t.Error("**text** should not be HR")
	}
	if isHR("--") {
		t.Error("-- is too short")
	}
}
