package renderer

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/fogleman/gg"

	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
)

// renderShapeLayer renders a shape layer
func (r *Renderer) renderShapeLayer(dc *gg.Context, layer templates.Layer, vars map[string]string) error {
	// Resolve colors with variable substitution
	fillColor := r.variableProcessor.SubstituteVariables(layer.Fill, vars)
	strokeColor := r.variableProcessor.SubstituteVariables(layer.Stroke, vars)

	// Get region
	region := layer.Region
	x := float64(region.X)
	y := float64(region.Y)
	w := float64(region.Width)
	h := float64(region.Height)

	// Draw the shape
	switch layer.Shape {
	case "rectangle":
		return r.renderRectangle(dc, x, y, w, h, layer.CornerRadius, fillColor, strokeColor, layer.StrokeWidth)
	case "circle":
		return r.renderCircle(dc, x, y, w, h, fillColor, strokeColor, layer.StrokeWidth)
	case "ellipse":
		return r.renderEllipse(dc, x, y, w, h, fillColor, strokeColor, layer.StrokeWidth)
	case "polygon":
		return r.renderPolygon(dc, x, y, w, h, layer.Points, fillColor, strokeColor, layer.StrokeWidth)
	default:
		return fmt.Errorf("unknown shape type: %s", layer.Shape)
	}
}

// renderRectangle renders a rectangle or rounded rectangle
func (r *Renderer) renderRectangle(dc *gg.Context, x, y, w, h, cornerRadius float64, fill, stroke string, strokeWidth float64) error {
	if cornerRadius > 0 {
		dc.DrawRoundedRectangle(x, y, w, h, cornerRadius)
	} else {
		dc.DrawRectangle(x, y, w, h)
	}

	// Fill if color specified
	if fill != "" {
		fillCol, err := parseRGBAColor(fill)
		if err != nil {
			return fmt.Errorf("invalid fill color: %v", err)
		}
		dc.SetColor(fillCol)
		if stroke != "" {
			dc.FillPreserve() // Preserve path for stroke
		} else {
			dc.Fill()
		}
	}

	// Stroke if color specified
	if stroke != "" && strokeWidth > 0 {
		strokeCol, err := parseRGBAColor(stroke)
		if err != nil {
			return fmt.Errorf("invalid stroke color: %v", err)
		}
		dc.SetColor(strokeCol)
		dc.SetLineWidth(strokeWidth)
		dc.Stroke()
	}

	return nil
}

// renderCircle renders a circle (using the smaller dimension for radius)
func (r *Renderer) renderCircle(dc *gg.Context, x, y, w, h float64, fill, stroke string, strokeWidth float64) error {
	// Use smaller dimension for radius to fit within region
	radius := math.Min(w, h) / 2
	centerX := x + w/2
	centerY := y + h/2

	dc.DrawCircle(centerX, centerY, radius)

	// Fill if color specified
	if fill != "" {
		fillCol, err := parseRGBAColor(fill)
		if err != nil {
			return fmt.Errorf("invalid fill color: %v", err)
		}
		dc.SetColor(fillCol)
		if stroke != "" {
			dc.FillPreserve()
		} else {
			dc.Fill()
		}
	}

	// Stroke if color specified
	if stroke != "" && strokeWidth > 0 {
		strokeCol, err := parseRGBAColor(stroke)
		if err != nil {
			return fmt.Errorf("invalid stroke color: %v", err)
		}
		dc.SetColor(strokeCol)
		dc.SetLineWidth(strokeWidth)
		dc.Stroke()
	}

	return nil
}

// renderEllipse renders an ellipse
func (r *Renderer) renderEllipse(dc *gg.Context, x, y, w, h float64, fill, stroke string, strokeWidth float64) error {
	centerX := x + w/2
	centerY := y + h/2
	radiusX := w / 2
	radiusY := h / 2

	dc.DrawEllipse(centerX, centerY, radiusX, radiusY)

	// Fill if color specified
	if fill != "" {
		fillCol, err := parseRGBAColor(fill)
		if err != nil {
			return fmt.Errorf("invalid fill color: %v", err)
		}
		dc.SetColor(fillCol)
		if stroke != "" {
			dc.FillPreserve()
		} else {
			dc.Fill()
		}
	}

	// Stroke if color specified
	if stroke != "" && strokeWidth > 0 {
		strokeCol, err := parseRGBAColor(stroke)
		if err != nil {
			return fmt.Errorf("invalid stroke color: %v", err)
		}
		dc.SetColor(strokeCol)
		dc.SetLineWidth(strokeWidth)
		dc.Stroke()
	}

	return nil
}

// renderPolygon renders a polygon from points
func (r *Renderer) renderPolygon(dc *gg.Context, x, y, w, h float64, points [][]float64, fill, stroke string, strokeWidth float64) error {
	if len(points) < 3 {
		return fmt.Errorf("polygon needs at least 3 points")
	}

	// Scale points to region (points are relative to region dimensions)
	for i, point := range points {
		if len(point) != 2 {
			return fmt.Errorf("polygon point %d must have x,y coordinates", i)
		}

		// Scale relative coordinates to actual region
		scaledX := x + (point[0]/100.0)*w // Assume points are 0-100 relative
		scaledY := y + (point[1]/100.0)*h

		if i == 0 {
			dc.MoveTo(scaledX, scaledY)
		} else {
			dc.LineTo(scaledX, scaledY)
		}
	}
	dc.ClosePath()

	// Fill if color specified
	if fill != "" {
		fillCol, err := parseRGBAColor(fill)
		if err != nil {
			return fmt.Errorf("invalid fill color: %v", err)
		}
		dc.SetColor(fillCol)
		if stroke != "" {
			dc.FillPreserve()
		} else {
			dc.Fill()
		}
	}

	// Stroke if color specified
	if stroke != "" && strokeWidth > 0 {
		strokeCol, err := parseRGBAColor(stroke)
		if err != nil {
			return fmt.Errorf("invalid stroke color: %v", err)
		}
		dc.SetColor(strokeCol)
		dc.SetLineWidth(strokeWidth)
		dc.Stroke()
	}

	return nil
}

// parseRGBAColor parses RGBA color strings
// Supports formats: "#RRGGBB", "#RRGGBBAA", "#RGB", "#RGBA"
func parseRGBAColor(colorStr string) (color.Color, error) {
	if !strings.HasPrefix(colorStr, "#") {
		return nil, fmt.Errorf("color must start with #")
	}

	hex := colorStr[1:]

	var r, g, b, a uint8 = 0, 0, 0, 255 // Default alpha to fully opaque

	switch len(hex) {
	case 3: // #RGB
		if rv, err := strconv.ParseUint(hex[0:1], 16, 8); err == nil {
			r = uint8(rv * 17) // Convert single hex digit to full byte
		} else {
			return nil, fmt.Errorf("invalid red component")
		}
		if gv, err := strconv.ParseUint(hex[1:2], 16, 8); err == nil {
			g = uint8(gv * 17)
		} else {
			return nil, fmt.Errorf("invalid green component")
		}
		if bv, err := strconv.ParseUint(hex[2:3], 16, 8); err == nil {
			b = uint8(bv * 17)
		} else {
			return nil, fmt.Errorf("invalid blue component")
		}

	case 4: // #RGBA
		if rv, err := strconv.ParseUint(hex[0:1], 16, 8); err == nil {
			r = uint8(rv * 17)
		} else {
			return nil, fmt.Errorf("invalid red component")
		}
		if gv, err := strconv.ParseUint(hex[1:2], 16, 8); err == nil {
			g = uint8(gv * 17)
		} else {
			return nil, fmt.Errorf("invalid green component")
		}
		if bv, err := strconv.ParseUint(hex[2:3], 16, 8); err == nil {
			b = uint8(bv * 17)
		} else {
			return nil, fmt.Errorf("invalid blue component")
		}
		if av, err := strconv.ParseUint(hex[3:4], 16, 8); err == nil {
			a = uint8(av * 17)
		} else {
			return nil, fmt.Errorf("invalid alpha component")
		}

	case 6: // #RRGGBB
		if rv, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
			r = uint8(rv)
		} else {
			return nil, fmt.Errorf("invalid red component")
		}
		if gv, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
			g = uint8(gv)
		} else {
			return nil, fmt.Errorf("invalid green component")
		}
		if bv, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
			b = uint8(bv)
		} else {
			return nil, fmt.Errorf("invalid blue component")
		}

	case 8: // #RRGGBBAA
		if rv, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
			r = uint8(rv)
		} else {
			return nil, fmt.Errorf("invalid red component")
		}
		if gv, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
			g = uint8(gv)
		} else {
			return nil, fmt.Errorf("invalid green component")
		}
		if bv, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
			b = uint8(bv)
		} else {
			return nil, fmt.Errorf("invalid blue component")
		}
		if av, err := strconv.ParseUint(hex[6:8], 16, 8); err == nil {
			a = uint8(av)
		} else {
			return nil, fmt.Errorf("invalid alpha component")
		}

	default:
		return nil, fmt.Errorf("invalid color format: expected #RGB, #RGBA, #RRGGBB, or #RRGGBBAA")
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}
