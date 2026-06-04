package renderer

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/expr"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
	"github.com/fogleman/gg"
)

func renderShapeLayer(dc *gg.Context, layer types.Layer, proc *expr.Processor, vars map[string]string) error {
	fill := proc.Substitute(layer.Fill)
	stroke := proc.Substitute(layer.Stroke)

	x := float64(layer.Region.X)
	y := float64(layer.Region.Y)
	w := float64(layer.Region.Width)
	h := float64(layer.Region.Height)

	switch layer.Shape {
	case "rectangle":
		return renderRect(dc, x, y, w, h, layer.CornerRadius, fill, stroke, layer.StrokeWidth)
	case "circle":
		return renderCircle(dc, x, y, w, h, fill, stroke, layer.StrokeWidth)
	case "ellipse":
		return renderEllipse(dc, x, y, w, h, fill, stroke, layer.StrokeWidth)
	case "polygon":
		return renderPolygon(dc, x, y, w, h, layer.Points, fill, stroke, layer.StrokeWidth)
	default:
		return fmt.Errorf("unknown shape %q", layer.Shape)
	}
}

func renderRect(dc *gg.Context, x, y, w, h, radius float64, fill, stroke string, strokeWidth float64) error {
	if radius > 0 {
		dc.DrawRoundedRectangle(x, y, w, h, radius)
	} else {
		dc.DrawRectangle(x, y, w, h)
	}
	return applyFillStroke(dc, fill, stroke, strokeWidth)
}

func renderCircle(dc *gg.Context, x, y, w, h float64, fill, stroke string, strokeWidth float64) error {
	r := math.Min(w, h) / 2
	dc.DrawCircle(x+w/2, y+h/2, r)
	return applyFillStroke(dc, fill, stroke, strokeWidth)
}

func renderEllipse(dc *gg.Context, x, y, w, h float64, fill, stroke string, strokeWidth float64) error {
	dc.DrawEllipse(x+w/2, y+h/2, w/2, h/2)
	return applyFillStroke(dc, fill, stroke, strokeWidth)
}

func renderPolygon(dc *gg.Context, x, y, w, h float64, points [][]float64, fill, stroke string, strokeWidth float64) error {
	if len(points) < 3 {
		return fmt.Errorf("polygon needs at least 3 points, got %d", len(points))
	}
	for i, pt := range points {
		if len(pt) != 2 {
			return fmt.Errorf("point %d must have 2 coordinates", i)
		}
		px := x + (pt[0]/100.0)*w
		py := y + (pt[1]/100.0)*h
		if i == 0 {
			dc.MoveTo(px, py)
		} else {
			dc.LineTo(px, py)
		}
	}
	dc.ClosePath()
	return applyFillStroke(dc, fill, stroke, strokeWidth)
}

func applyFillStroke(dc *gg.Context, fill, stroke string, strokeWidth float64) error {
	if fill != "" {
		c, err := parseColor(fill)
		if err != nil {
			return fmt.Errorf("fill color: %w", err)
		}
		dc.SetColor(c)
		if stroke != "" {
			dc.FillPreserve()
		} else {
			dc.Fill()
		}
	}

	if stroke != "" && strokeWidth > 0 {
		c, err := parseColor(stroke)
		if err != nil {
			return fmt.Errorf("stroke color: %w", err)
		}
		dc.SetColor(c)
		dc.SetLineWidth(strokeWidth)
		dc.Stroke()
	}

	return nil
}

// ─── color parsing ─────────────────────────────────────────────────────────────

// parseColor parses #RGB, #RGBA, #RRGGBB, #RRGGBBAA hex color strings.
func parseColor(s string) (color.Color, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "#") {
		return nil, fmt.Errorf("color must start with #, got %q", s)
	}

	hex := s[1:]
	var r, g, b, a uint8 = 0, 0, 0, 255

	parseHex1 := func(h string) (uint8, error) {
		var v uint64
		_, err := fmt.Sscanf(h, "%x", &v)
		if err != nil || v > 15 {
			return 0, fmt.Errorf("invalid hex nibble %q", h)
		}
		return uint8(v * 17), nil
	}
	parseHex2 := func(h string) (uint8, error) {
		var v uint64
		_, err := fmt.Sscanf(h, "%x", &v)
		if err != nil || v > 255 {
			return 0, fmt.Errorf("invalid hex byte %q", h)
		}
		return uint8(v), nil
	}

	var err error
	switch len(hex) {
	case 3:
		if r, err = parseHex1(hex[0:1]); err != nil {
			return nil, err
		}
		if g, err = parseHex1(hex[1:2]); err != nil {
			return nil, err
		}
		if b, err = parseHex1(hex[2:3]); err != nil {
			return nil, err
		}
	case 4:
		if r, err = parseHex1(hex[0:1]); err != nil {
			return nil, err
		}
		if g, err = parseHex1(hex[1:2]); err != nil {
			return nil, err
		}
		if b, err = parseHex1(hex[2:3]); err != nil {
			return nil, err
		}
		if a, err = parseHex1(hex[3:4]); err != nil {
			return nil, err
		}
	case 6:
		if r, err = parseHex2(hex[0:2]); err != nil {
			return nil, err
		}
		if g, err = parseHex2(hex[2:4]); err != nil {
			return nil, err
		}
		if b, err = parseHex2(hex[4:6]); err != nil {
			return nil, err
		}
	case 8:
		if r, err = parseHex2(hex[0:2]); err != nil {
			return nil, err
		}
		if g, err = parseHex2(hex[2:4]); err != nil {
			return nil, err
		}
		if b, err = parseHex2(hex[4:6]); err != nil {
			return nil, err
		}
		if a, err = parseHex2(hex[6:8]); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported color format %q (expected #RGB, #RGBA, #RRGGBB, #RRGGBBAA)", s)
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}
