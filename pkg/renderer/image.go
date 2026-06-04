package renderer

import (
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Merith-TK/tcg-cardgen/pkg/types"
	"github.com/fogleman/gg"
)

// ─── global image cache ───────────────────────────────────────────────────────

var (
	imageCache   sync.Map // map[string]image.Image
)

// loadImage loads an image from a local path or HTTP/HTTPS URL with
// process-lifetime caching.
func loadImage(path string) (image.Image, error) {
	if v, ok := imageCache.Load(path); ok {
		return v.(image.Image), nil
	}

	var img image.Image
	var err error

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		img, err = downloadImage(path)
	} else {
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			return nil, fmt.Errorf("image not found: %s", path)
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".svg" {
			return nil, fmt.Errorf("SVG not yet supported (use PNG/JPG): %s", path)
		}
		img, err = gg.LoadImage(path)
	}

	if err != nil {
		return nil, err
	}

	imageCache.Store(path, img)
	return img, nil
}

// downloadImage downloads an image from a URL.
func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url) // #nosec G107 — URL is intentionally user-supplied
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", url, err)
	}
	return img, nil
}

// createFittedImage scales/crops an image to fit a region.
func createFittedImage(img image.Image, region types.Region, fitMode string) image.Image {
	bounds := img.Bounds()
	iw := float64(bounds.Dx())
	ih := float64(bounds.Dy())
	rw := float64(region.Width)
	rh := float64(region.Height)

	dc := gg.NewContext(region.Width, region.Height)

	switch fitMode {
	case "fill":
		sx := rw / iw
		sy := rh / ih
		s := sx
		if sy > sx {
			s = sy
		}
		dc.Scale(s, s)
		dx := (rw/s - iw) / 2
		dy := (rh/s - ih) / 2
		dc.DrawImage(img, int(dx), int(dy))

	case "fit":
		sx := rw / iw
		sy := rh / ih
		s := sx
		if sy < sx {
			s = sy
		}
		dc.Scale(s, s)
		dx := (rw/s - iw) / 2
		dy := (rh/s - ih) / 2
		dc.DrawImage(img, int(dx), int(dy))

	case "stretch":
		dc.ScaleAbout(rw/iw, rh/ih, 0, 0)
		dc.DrawImage(img, 0, 0)

	case "center":
		dx := (rw - iw) / 2
		dy := (rh - ih) / 2
		dc.DrawImage(img, int(dx), int(dy))

	default:
		return createFittedImage(img, region, "fill")
	}

	return dc.Image()
}

// renderPlaceholder draws a grey placeholder rectangle with a label.
func renderPlaceholder(dc *gg.Context, layer types.Layer, label string) {
	x := float64(layer.Region.X)
	y := float64(layer.Region.Y)
	w := float64(layer.Region.Width)
	h := float64(layer.Region.Height)

	dc.SetColor(color.RGBA{200, 200, 200, 255})
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()

	dc.SetColor(color.RGBA{100, 100, 100, 255})
	dc.SetLineWidth(2)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	dc.SetColor(color.RGBA{50, 50, 50, 255})
	dc.DrawStringAnchored(label, x+w/2, y+h/2, 0.5, 0.5)
}
