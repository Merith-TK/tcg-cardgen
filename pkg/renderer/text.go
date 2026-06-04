package renderer

import (
	"image/color"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

// ─── types ────────────────────────────────────────────────────────────────────

type textStyle struct {
	Bold   bool
	Italic bool
}

type formattedSegment struct {
	Content string
	Style   textStyle
}

type lineKind int

const (
	lineNormal lineKind = iota
	lineHeader
	lineHR
)

type formattedLine struct {
	Segments []formattedSegment
	Kind     lineKind
	Level    int // header level 1-6
}

// ─── markdown parser ──────────────────────────────────────────────────────────

// parseMarkdown converts a markdown string into a slice of formatted lines.
func parseMarkdown(content string) []formattedLine {
	raw := strings.Split(content, "\n")
	var out []formattedLine

	for _, raw := range raw {
		line := strings.TrimSpace(raw)

		if line == "" {
			out = append(out, formattedLine{Kind: lineNormal})
			continue
		}

		// Horizontal rule: ---, ***, ===
		if isHR(line) {
			out = append(out, formattedLine{Kind: lineHR})
			continue
		}

		// Heading: # through ######
		if strings.HasPrefix(line, "#") {
			level, text := parseHeading(line)
			if level > 0 {
				out = append(out, formattedLine{
					Segments: parseInline(text),
					Kind:     lineHeader,
					Level:    level,
				})
				continue
			}
		}

		out = append(out, formattedLine{
			Segments: parseInline(line),
			Kind:     lineNormal,
		})
	}

	return out
}

func isHR(s string) bool {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) < 3 {
		return false
	}
	ch := trimmed[0]
	if ch != '-' && ch != '*' && ch != '=' {
		return false
	}
	for _, c := range trimmed {
		if byte(c) != ch && c != ' ' {
			return false
		}
	}
	return true
}

func parseHeading(line string) (level int, text string) {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	if i > 0 && i <= 6 && i < len(line) && line[i] == ' ' {
		return i, strings.TrimSpace(line[i+1:])
	}
	return 0, ""
}

// parseInline parses inline bold/italic markdown into a slice of segments.
// Correctly handles ***bold-italic*** before **bold** before *italic*
// by always matching the longest marker first.
func parseInline(text string) []formattedSegment {
	return parseInlineRecursive(text, textStyle{})
}

func parseInlineRecursive(text string, inherited textStyle) []formattedSegment {
	if text == "" {
		return nil
	}

	// Find the earliest marker and its length.
	// We check *** before ** before * to ensure longest-match semantics.
	type markerMatch struct {
		pos     int
		length  int
		bold    bool
		italic  bool
	}

	best := markerMatch{pos: -1}

	for _, m := range []markerMatch{
		{length: 3, bold: true, italic: true},
		{length: 2, bold: true, italic: false},
		{length: 1, bold: false, italic: true},
	} {
		marker := strings.Repeat("*", m.length)
		idx := strings.Index(text, marker)
		if idx == -1 {
			continue
		}
		// Make sure this isn't the start of a longer marker.
		if m.length < 3 && idx+m.length < len(text) && text[idx+m.length] == '*' {
			continue
		}
		if best.pos == -1 || idx < best.pos {
			best = markerMatch{pos: idx, length: m.length, bold: m.bold, italic: m.italic}
		}
	}

	if best.pos == -1 {
		return []formattedSegment{{Content: text, Style: inherited}}
	}

	var out []formattedSegment

	// Text before the marker.
	if best.pos > 0 {
		out = append(out, formattedSegment{Content: text[:best.pos], Style: inherited})
	}

	marker := strings.Repeat("*", best.length)
	after := text[best.pos+best.length:]
	closeIdx := strings.Index(after, marker)
	if closeIdx == -1 {
		// No closing marker; emit the rest as plain text.
		out = append(out, formattedSegment{Content: text[best.pos:], Style: inherited})
		return out
	}

	innerStyle := textStyle{
		Bold:   inherited.Bold || best.bold,
		Italic: inherited.Italic || best.italic,
	}
	inner := after[:closeIdx]
	out = append(out, parseInlineRecursive(inner, innerStyle)...)

	remainder := after[closeIdx+best.length:]
	out = append(out, parseInlineRecursive(remainder, inherited)...)
	return out
}

// stripMarkdownHeaders removes lines that are headings.
func stripMarkdownHeaders(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), "#") {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

// ─── text rendering ───────────────────────────────────────────────────────────

// drawFormattedText renders a slice of formatted lines into a region.
func drawFormattedText(dc *gg.Context, lines []formattedLine, x, y, w, h float64, align string, baseSize float64, baseColor color.Color, fontFamily, templateDir string) {
	if len(lines) == 0 {
		return
	}

	lineHeight := baseSize * 1.35

	// First pass: measure total height so we can vertically centre the block.
	totalH := 0.0
	for _, l := range lines {
		switch l.Kind {
		case lineHeader:
			totalH += headerSize(baseSize, l.Level) * 1.4
		case lineHR:
			totalH += baseSize * 0.6
		case lineNormal:
			if len(l.Segments) == 0 {
				totalH += lineHeight * 0.5
			} else {
				totalH += lineHeight * float64(countWrappedLines(dc, l.Segments, w, baseSize, fontFamily, templateDir))
			}
		}
	}

	curY := y + (h-totalH)/2
	if curY < y {
		curY = y
	}

	// Second pass: render.
	for _, l := range lines {
		switch l.Kind {
		case lineHeader:
			hs := headerSize(baseSize, l.Level)
			setFontFace(dc, fontFamily, templateDir, hs, true, false, baseColor)
			combined := combineSegments(l.Segments)
			drawSingleLine(dc, combined, x, curY, w, align)
			curY += hs * 1.4

		case lineHR:
			dc.SetColor(color.RGBA{128, 128, 128, 255})
			dc.SetLineWidth(1)
			ruleY := curY + baseSize*0.3
			dc.DrawLine(x+w*0.05, ruleY, x+w*0.95, ruleY)
			dc.Stroke()
			curY += baseSize * 0.6

		case lineNormal:
			if len(l.Segments) == 0 {
				curY += lineHeight * 0.5
			} else {
				curY = drawFormattedLine(dc, l.Segments, x, curY, w, baseSize, baseColor, fontFamily, templateDir, align, lineHeight)
			}
		}
	}
}

func drawFormattedLine(dc *gg.Context, segments []formattedSegment, x, y, w, baseSize float64, baseColor color.Color, fontFamily, templateDir, align string, lineHeight float64) float64 {
	wrapped := wrapSegments(dc, segments, w, baseSize, fontFamily, templateDir)
	curY := y
	for _, wline := range wrapped {
		renderWrappedLine(dc, wline, x, curY, w, baseSize, baseColor, fontFamily, templateDir, align)
		curY += lineHeight
	}
	return curY
}

// wrapSegments wraps formatted segments to fit within maxWidth pixels.
func wrapSegments(dc *gg.Context, segments []formattedSegment, maxWidth, baseSize float64, fontFamily, templateDir string) [][]formattedSegment {
	var lines [][]formattedSegment
	var curLine []formattedSegment
	curWidth := 0.0

	for _, seg := range segments {
		setFontFace(dc, fontFamily, templateDir, baseSize, seg.Style.Bold, seg.Style.Italic, color.Black)
		words := strings.Fields(seg.Content)
		if len(words) == 0 {
			continue
		}

		for wi, word := range words {
			// Add a space before words that are not the first on the line.
			prefix := ""
			if wi > 0 || len(curLine) > 0 {
				prefix = " "
			}
			displayWord := prefix + word
			ww, _ := dc.MeasureString(displayWord)

			if curWidth+ww > maxWidth && len(curLine) > 0 {
				lines = append(lines, curLine)
				curLine = nil
				curWidth = 0
				displayWord = word
				ww, _ = dc.MeasureString(displayWord)
			}
			curLine = append(curLine, formattedSegment{Content: displayWord, Style: seg.Style})
			curWidth += ww
		}
	}

	if len(curLine) > 0 {
		lines = append(lines, curLine)
	}
	return lines
}

// countWrappedLines estimates how many display lines a set of segments will occupy.
func countWrappedLines(dc *gg.Context, segments []formattedSegment, maxWidth, baseSize float64, fontFamily, templateDir string) int {
	wrapped := wrapSegments(dc, segments, maxWidth, baseSize, fontFamily, templateDir)
	if len(wrapped) == 0 {
		return 1
	}
	return len(wrapped)
}

func renderWrappedLine(dc *gg.Context, segments []formattedSegment, x, y, w, baseSize float64, baseColor color.Color, fontFamily, templateDir, align string) {
	if len(segments) == 0 {
		return
	}

	// Measure total width for alignment.
	totalW := 0.0
	for _, seg := range segments {
		setFontFace(dc, fontFamily, templateDir, baseSize, seg.Style.Bold, seg.Style.Italic, baseColor)
		sw, _ := dc.MeasureString(seg.Content)
		totalW += sw
	}

	curX := x
	switch align {
	case "center":
		curX = x + (w-totalW)/2
	case "right":
		curX = x + w - totalW
	}

	for _, seg := range segments {
		setFontFace(dc, fontFamily, templateDir, baseSize, seg.Style.Bold, seg.Style.Italic, baseColor)
		dc.DrawStringAnchored(seg.Content, curX, y, 0, 0)
		sw, _ := dc.MeasureString(seg.Content)
		curX += sw
	}
}

func drawSingleLine(dc *gg.Context, text string, x, y, w float64, align string) {
	switch align {
	case "right":
		dc.DrawStringAnchored(text, x+w, y, 1.0, 0.0)
	case "center":
		dc.DrawStringAnchored(text, x+w/2, y, 0.5, 0.0)
	default:
		dc.DrawStringAnchored(text, x, y, 0.0, 0.0)
	}
}

func combineSegments(segs []formattedSegment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Content)
	}
	return b.String()
}

func headerSize(base float64, level int) float64 {
	scales := []float64{1.8, 1.6, 1.4, 1.2, 1.1, 1.0}
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return base * scales[level-1]
}

// ─── font face helper ─────────────────────────────────────────────────────────

func setFontFace(dc *gg.Context, family, templateDir string, size float64, bold, italic bool, col color.Color) {
	v := varRegular
	switch {
	case bold && italic:
		v = varBoldItalic
	case bold:
		v = varBold
	case italic:
		v = varItalic
	}

	f := loadFont(family, templateDir, v)
	face := truetype.NewFace(f, &truetype.Options{Size: size, DPI: 72})
	dc.SetFontFace(face)
	dc.SetColor(col)
}
