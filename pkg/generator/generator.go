// Package generator orchestrates the full card generation pipeline:
// parse card → load template → validate → render → save PNG.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Merith-TK/tcg-cardgen/pkg/card"
	"github.com/Merith-TK/tcg-cardgen/pkg/renderer"
	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
)

// Generator runs the card generation pipeline.
type Generator struct {
	cfg  *types.Config
	tmgr *templates.Manager
}

// New creates a Generator with the given config.
func New(cfg *types.Config) *Generator {
	if cfg.OutputDir == "" {
		cfg.OutputDir = ".tcg-cardgen-out"
	}
	return &Generator{
		cfg:  cfg,
		tmgr: templates.NewManager(cfg.TemplateDir),
	}
}

// GenerateCard processes a single .md file and produces a PNG.
func (g *Generator) GenerateCard(filePath string) error {
	g.logf("Parsing: %s", filePath)

	c, err := card.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filePath, err)
	}

	g.logf("TCG=%s  cardstyle=%s  title=%q", c.TCG, c.CardStyle, c.Title)

	t, err := g.tmgr.LoadTemplate(c.TCG, c.CardStyle)
	if err != nil {
		return fmt.Errorf("load template %s/%s: %w", c.TCG, c.CardStyle, err)
	}

	if err := t.ValidateCard(c); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	if g.cfg.ValidateOnly {
		fmt.Printf("valid: %s\n", filePath)
		return nil
	}

	// Determine output path.
	// If OutputDir is absolute, write all cards there flat.
	// If relative, write next to the source file.
	var outDir string
	if filepath.IsAbs(g.cfg.OutputDir) {
		outDir = g.cfg.OutputDir
	} else {
		outDir = filepath.Join(filepath.Dir(filePath), g.cfg.OutputDir)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create output dir %s: %w", outDir, err)
	}

	base := filepath.Base(filePath)
	nameNoExt := base[:len(base)-len(filepath.Ext(base))]
	outPath := filepath.Join(outDir, nameNoExt+".png")

	g.logf("Rendering → %s", outPath)

	if err := renderer.RenderCard(c, t, outPath); err != nil {
		return fmt.Errorf("render %s: %w", filePath, err)
	}

	if g.cfg.Verbose {
		fmt.Printf("generated: %s\n", outPath)
	} else {
		fmt.Printf("%s → %s\n", filePath, outPath)
	}

	return nil
}

// GenerateDirectory processes all .md files under dir (excluding outDir).
// If concurrency > 1, files are processed in parallel.
func (g *Generator) GenerateDirectory(dir string) error {
	files, err := collectCards(dir, g.cfg.OutputDir)
	if err != nil {
		return err
	}

	workers := g.cfg.Concurrency
	if workers <= 0 {
		workers = 1
	}

	if workers == 1 {
		var firstErr error
		for _, f := range files {
			if err := g.GenerateCard(f); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				if firstErr == nil {
					firstErr = err
				}
			}
		}
		return firstErr
	}

	// Parallel mode.
	sem := make(chan struct{}, workers)
	var (
		mu      sync.Mutex
		firstErr error
	)
	var wg sync.WaitGroup
	for _, f := range files {
		f := f
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := g.GenerateCard(f); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return firstErr
}

// ListCardstyles returns all discovered card styles.
func (g *Generator) ListCardstyles() ([]types.CardStyleInfo, error) {
	return g.tmgr.ListAvailableCardstyles()
}

// TemplateManager exposes the template manager (used by the designer server).
func (g *Generator) TemplateManager() *templates.Manager {
	return g.tmgr
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (g *Generator) logf(format string, args ...interface{}) {
	if g.cfg.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// collectCards walks dir and returns all .md files, skipping outDir.
func collectCards(dir, outDir string) ([]string, error) {
	var files []string
	absOut := ""
	if outDir != "" {
		if filepath.IsAbs(outDir) {
			absOut = outDir
		} else {
			absOut, _ = filepath.Abs(filepath.Join(dir, outDir))
		}
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			absPath, _ := filepath.Abs(path)
			// Skip the output directory.
			if absOut != "" && absPath == absOut {
				return filepath.SkipDir
			}
			// Skip hidden directories.
			base := filepath.Base(path)
			if len(base) > 1 && base[0] == '.' && path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
