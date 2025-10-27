package renderer

import (
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
	"github.com/fogleman/gg"
)

// ImageProcessor handles all image-related operations
type ImageProcessor struct {
	cache map[string]image.Image
}

// NewImageProcessor creates a new image processor
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		cache: make(map[string]image.Image),
	}
}

// LoadImage loads an image with caching (supports local files and URLs)
func (ip *ImageProcessor) LoadImage(path string) (image.Image, error) {
	// Check cache first
	if img, exists := ip.cache[path]; exists {
		return img, nil
	}

	var img image.Image
	var err error

	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		img, err = ip.downloadImage(path)
	} else {
		// Check if local file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("image file not found: %s", path)
		}

		// Load local image
		img, err = gg.LoadImage(path)
	}

	if err != nil {
		return nil, err
	}

	// Cache it
	ip.cache[path] = img
	return img, nil
}

// downloadImage downloads an image from a URL
func (ip *ImageProcessor) downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}

// CreateFittedImage creates a new image that fits the specified region with the given fit mode
func (ip *ImageProcessor) CreateFittedImage(img image.Image, region templates.Region, fitMode string) image.Image {
	imgBounds := img.Bounds()
	imgWidth := float64(imgBounds.Dx())
	imgHeight := float64(imgBounds.Dy())

	regionWidth := float64(region.Width)
	regionHeight := float64(region.Height)

	// Create a new image context for the fitted result
	fittedDC := gg.NewContext(region.Width, region.Height)

	switch fitMode {
	case "fill": // Scale to fill region completely, crop if necessary
		// Calculate scaling to fill the region (crop if necessary)
		scaleX := regionWidth / imgWidth
		scaleY := regionHeight / imgHeight
		scale := scaleX
		if scaleY > scaleX {
			scale = scaleY // Use larger scale to fill region completely
		}

		// Calculate scaled dimensions
		scaledWidth := imgWidth * scale
		scaledHeight := imgHeight * scale

		// Calculate position to center the scaled image
		drawX := (regionWidth - scaledWidth) / 2
		drawY := (regionHeight - scaledHeight) / 2

		// Scale and draw the image
		fittedDC.Scale(scale, scale)
		fittedDC.DrawImageAnchored(img, int(drawX/scale+imgWidth/2), int(drawY/scale+imgHeight/2), 0.5, 0.5)

	case "fit": // Scale to fit entirely within region, may leave empty space
		// Calculate scaling to fit within the region
		scaleX := regionWidth / imgWidth
		scaleY := regionHeight / imgHeight
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY // Use smaller scale to fit entirely
		}

		// Calculate scaled dimensions
		scaledWidth := imgWidth * scale
		scaledHeight := imgHeight * scale

		// Calculate position to center the scaled image
		drawX := (regionWidth - scaledWidth) / 2
		drawY := (regionHeight - scaledHeight) / 2

		// Scale and draw the image
		fittedDC.Scale(scale, scale)
		fittedDC.DrawImageAnchored(img, int(drawX/scale+imgWidth/2), int(drawY/scale+imgHeight/2), 0.5, 0.5)

	case "stretch": // Stretch to exact region dimensions (may distort)
		fittedDC.DrawImageAnchored(img, region.Width/2, region.Height/2, 0.5, 0.5)

	case "center": // No scaling, just center (may crop or leave empty space)
		drawX := (regionWidth - imgWidth) / 2
		drawY := (regionHeight - imgHeight) / 2
		fittedDC.DrawImageAnchored(img, int(drawX+imgWidth/2), int(drawY+imgHeight/2), 0.5, 0.5)

	default: // Default to fill
		return ip.CreateFittedImage(img, region, "fill")
	}

	return fittedDC.Image()
}

// RenderPlaceholder renders a placeholder rectangle with text
func (ip *ImageProcessor) RenderPlaceholder(dc *gg.Context, layer templates.Layer, text string) {
	// Draw placeholder rectangle
	dc.SetColor(color.RGBA{200, 200, 200, 255})
	dc.DrawRectangle(float64(layer.Region.X), float64(layer.Region.Y),
		float64(layer.Region.Width), float64(layer.Region.Height))
	dc.Fill()

	// Draw border
	dc.SetColor(color.RGBA{100, 100, 100, 255})
	dc.SetLineWidth(2)
	dc.DrawRectangle(float64(layer.Region.X), float64(layer.Region.Y),
		float64(layer.Region.Width), float64(layer.Region.Height))
	dc.Stroke()

	// Draw text
	dc.SetColor(color.RGBA{50, 50, 50, 255})
	dc.DrawStringAnchored(text,
		float64(layer.Region.X+layer.Region.Width/2),
		float64(layer.Region.Y+layer.Region.Height/2),
		0.5, 0.5)
}
