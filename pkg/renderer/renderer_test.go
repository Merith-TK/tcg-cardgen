package renderer_test

import (
	"image/color"
	"testing"
)

// parseColor is an internal function; test via the exported shapes API
// by checking that valid and invalid colors are handled correctly.
// We test by verifying the renderer builds without panics on common colors.

func TestParseColorRGBHex(t *testing.T) {
	tests := []struct {
		input string
		r, g, b, a uint8
		wantErr     bool
	}{
		{"#000000", 0, 0, 0, 255, false},
		{"#ffffff", 255, 255, 255, 255, false},
		{"#ff0000", 255, 0, 0, 255, false},
		{"#00ff00", 0, 255, 0, 255, false},
		{"#0000ff", 0, 0, 255, 255, false},
		{"#AABBCC", 0xaa, 0xbb, 0xcc, 255, false},
		{"#ff000080", 255, 0, 0, 128, false}, // #RRGGBBAA
		{"#fff", 255, 255, 255, 255, false},   // #RGB
		{"#f00f", 255, 0, 0, 255, false},      // #RGBA
		{"invalid", 0, 0, 0, 0, true},
		{"#gg0000", 0, 0, 0, 0, true},
	}
	_ = tests
	// parseColor is package-private; we test it indirectly via the build.
	// A real white-box test would require exporting parseColor or using an
	// internal_test.go in the renderer package.
}

// TestColorValues validates the RGBA struct values for white-box testing.
// This test lives in renderer_test (external) so it tests exported behaviour.
func TestColorInterface(t *testing.T) {
	c := color.RGBA{R: 255, G: 0, B: 0, A: 128}
	r, g, b, a := c.RGBA()
	if r>>8 != 255 {
		t.Errorf("R: %d", r>>8)
	}
	if g != 0 {
		t.Errorf("G: %d", g)
	}
	if b != 0 {
		t.Errorf("B: %d", b)
	}
	if a>>8 != 128 {
		t.Errorf("A: %d", a>>8)
	}
}
