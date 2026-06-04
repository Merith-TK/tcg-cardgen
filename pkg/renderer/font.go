package renderer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

// fontVariant selects between the four font variants we need.
type fontVariant int

const (
	varRegular fontVariant = iota
	varBold
	varItalic
	varBoldItalic
)

// fontCacheKey uniquely identifies a loaded font variant.
type fontCacheKey struct {
	family  string
	variant fontVariant
}

var (
	fontCache   = make(map[fontCacheKey]*truetype.Font)
	fontCacheMu sync.RWMutex
)

// loadFont resolves a font by family name, checking in order:
//  1. <templateDir>/fonts/<family>[-Bold|-Italic|-BoldItalic].ttf / .otf
//  2. Common system font directories
//  3. Go built-in fonts (always succeeds)
func loadFont(family, templateDir string, v fontVariant) *truetype.Font {
	key := fontCacheKey{family: strings.ToLower(family), variant: v}

	fontCacheMu.RLock()
	if f, ok := fontCache[key]; ok {
		fontCacheMu.RUnlock()
		return f
	}
	fontCacheMu.RUnlock()

	f := resolveFont(family, templateDir, v)

	fontCacheMu.Lock()
	fontCache[key] = f
	fontCacheMu.Unlock()

	return f
}

// resolveFont does the actual search without touching the cache.
func resolveFont(family, templateDir string, v fontVariant) *truetype.Font {
	if family != "" {
		// --- 1. Template-relative fonts/ directory ---
		if templateDir != "" {
			if f := tryFontFiles(filepath.Join(templateDir, "fonts"), family, v); f != nil {
				return f
			}
		}

		// --- 2. System font directories ---
		for _, dir := range systemFontDirs() {
			if f := tryFontFiles(dir, family, v); f != nil {
				return f
			}
		}
	}

	// --- 3. Go built-in fallback ---
	return builtinFont(v)
}

// tryFontFiles attempts to load a font from dir using common naming conventions.
func tryFontFiles(dir, family string, v fontVariant) *truetype.Font {
	suffixes := variantSuffixes(v)
	exts := []string{".ttf", ".otf", ".TTF", ".OTF"}

	for _, suffix := range suffixes {
		for _, ext := range exts {
			candidates := []string{
				filepath.Join(dir, family+suffix+ext),
				filepath.Join(dir, strings.ToLower(family)+strings.ToLower(suffix)+ext),
				// Some fonts use hyphen separator: family-Bold.ttf
				filepath.Join(dir, family+"-"+strings.TrimPrefix(suffix, "-")+ext),
			}
			for _, path := range candidates {
				if f := parseFontFile(path); f != nil {
					return f
				}
			}
		}
	}
	return nil
}

// variantSuffixes returns the file-name suffixes to try for a given variant.
// The first element is the canonical suffix; others are common alternatives.
func variantSuffixes(v fontVariant) []string {
	switch v {
	case varBold:
		return []string{"-Bold", "Bold", "b", "-b"}
	case varItalic:
		return []string{"-Italic", "Italic", "-Oblique", "Oblique", "i", "-i"}
	case varBoldItalic:
		return []string{"-BoldItalic", "BoldItalic", "-BoldOblique", "BoldOblique", "bi", "-bi"}
	default: // varRegular
		return []string{"", "-Regular", "Regular", "r", "-r"}
	}
}

func parseFontFile(path string) *truetype.Font {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	f, err := truetype.Parse(data)
	if err != nil {
		return nil
	}
	return f
}

func builtinFont(v fontVariant) *truetype.Font {
	var data []byte
	switch v {
	case varBold:
		data = gobold.TTF
	case varItalic:
		data = goitalic.TTF
	case varBoldItalic:
		data = gobolditalic.TTF
	default:
		data = goregular.TTF
	}
	f, _ := truetype.Parse(data)
	return f
}

// systemFontDirs returns a list of directories to search for system fonts on
// the current OS.
func systemFontDirs() []string {
	switch runtime.GOOS {
	case "linux":
		dirs := []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
		}
		if home, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs,
				filepath.Join(home, ".fonts"),
				filepath.Join(home, ".local", "share", "fonts"),
			)
		}
		return dirs

	case "darwin":
		dirs := []string{
			"/Library/Fonts",
			"/System/Library/Fonts",
		}
		if home, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
		}
		return dirs

	case "windows":
		systemRoot := os.Getenv("SYSTEMROOT")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		dirs := []string{filepath.Join(systemRoot, "Fonts")}
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"))
		}
		return dirs

	default:
		return nil
	}
}
