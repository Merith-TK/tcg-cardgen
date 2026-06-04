// Package designer provides the embedded HTTP server for the card style designer.
package designer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/card"
	"github.com/Merith-TK/tcg-cardgen/pkg/renderer"
	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
	"gopkg.in/yaml.v3"

	"embed"
)

//go:embed static/*
var staticFiles embed.FS

// Server is the card style designer HTTP server.
type Server struct {
	tmgr *templates.Manager
	mux  *http.ServeMux
}

// New creates a Server ready to serve.
func New() *Server {
	s := &Server{
		tmgr: templates.NewManager(""),
		mux:  http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Static files (including the Preact+HTM app at /).
	s.mux.HandleFunc("/", s.handleStatic)

	// API endpoints.
	s.mux.HandleFunc("/api/templates", withCORS(s.apiListTemplates))
	s.mux.HandleFunc("/api/template", withCORS(s.apiTemplate))
	s.mux.HandleFunc("/api/render", withCORS(s.apiRender))
	s.mux.HandleFunc("/api/export", withCORS(s.apiExport))
}

// ─── static file serving ──────────────────────────────────────────────────────

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	// Strip leading slash and prefix with "static".
	fsPath := "static" + path

	data, err := staticFiles.ReadFile(fsPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType(filepath.Ext(path)))
	w.Write(data)
}

// ─── API: list templates ──────────────────────────────────────────────────────

func (s *Server) apiListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	all, err := s.tmgr.ListAvailableCardstyles()
	if err != nil {
		jsonError(w, "list templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Group by TCG.
	grouped := make(map[string][]types.CardStyleInfo)
	for _, cs := range all {
		grouped[cs.TCG] = append(grouped[cs.TCG], cs)
	}

	jsonOK(w, grouped)
}

// ─── API: get / save template ─────────────────────────────────────────────────

func (s *Server) apiTemplate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.apiGetTemplate(w, r)
	case http.MethodPost:
		s.apiSaveTemplate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) apiGetTemplate(w http.ResponseWriter, r *http.Request) {
	tcg := r.URL.Query().Get("tcg")
	name := r.URL.Query().Get("name")
	if tcg == "" || name == "" {
		jsonError(w, "missing tcg or name query parameter", http.StatusBadRequest)
		return
	}

	t, err := s.tmgr.LoadTemplate(tcg, name)
	if err != nil {
		jsonError(w, "load template: "+err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, t)
}

func (s *Server) apiSaveTemplate(w http.ResponseWriter, r *http.Request) {
	var t templates.Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if t.TCG == "" || t.Name == "" {
		jsonError(w, "template must have tcg and name fields", http.StatusBadRequest)
		return
	}

	// Resolve output path: workspace .tcg-cardstyles/<tcg>/<name>.yaml
	outDir := filepath.Join(".tcg-cardstyles", t.TCG)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		jsonError(w, "create dir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	outPath := filepath.Join(outDir, t.Name+".yaml")

	data, err := yaml.Marshal(&t)
	if err != nil {
		jsonError(w, "marshal template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		jsonError(w, "write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "saved", "path": outPath})
}

// ─── API: render ──────────────────────────────────────────────────────────────

// renderRequest is the JSON body for POST /api/render
type renderRequest struct {
	// Inline card frontmatter fields — same as a .md file's YAML block.
	Card map[string]interface{} `json:"card"`
	// Template can be a TCG/name reference or an inline template object.
	TemplateTCG  string             `json:"template_tcg"`
	TemplateName string             `json:"template_name"`
	Template     *templates.Template `json:"template,omitempty"`
}

func (s *Server) apiRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req renderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Build card from the provided map.
	c := buildCardFromMap(req.Card)

	// Resolve template.
	var t *templates.Template
	var err error
	if req.Template != nil {
		t = req.Template
	} else {
		tcg := req.TemplateTCG
		name := req.TemplateName
		if tcg == "" {
			tcg = c.TCG
		}
		if name == "" {
			name = c.CardStyle
		}
		t, err = s.tmgr.LoadTemplate(tcg, name)
		if err != nil {
			jsonError(w, "load template: "+err.Error(), http.StatusNotFound)
			return
		}
	}

	// Render to a temp file then read it back.
	tmp, err := os.CreateTemp("", "tcg-render-*.png")
	if err != nil {
		jsonError(w, "create temp: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := renderer.RenderCard(c, t, tmp.Name()); err != nil {
		jsonError(w, "render: "+err.Error(), http.StatusInternalServerError)
		return
	}

	imgData, err := os.ReadFile(tmp.Name())
	if err != nil {
		jsonError(w, "read render: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Accept header: return raw PNG or base64 JSON.
	if strings.Contains(r.Header.Get("Accept"), "image/png") {
		w.Header().Set("Content-Type", "image/png")
		w.Write(imgData)
	} else {
		jsonOK(w, map[string]string{
			"image": "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgData),
		})
	}
}

// ─── API: export ──────────────────────────────────────────────────────────────

func (s *Server) apiExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var t templates.Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	data, err := yaml.Marshal(&t)
	if err != nil {
		jsonError(w, "marshal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	name := t.Name
	if name == "" {
		name = "template"
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.yaml"`, name))
	w.Write(data)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func buildCardFromMap(m map[string]interface{}) *card.Card {
	c := &card.Card{
		Metadata: m,
	}
	// Pull well-known fields if present.
	if v, ok := m["tcg"].(string); ok {
		c.TCG = v
	}
	if v, ok := m["cardstyle"].(string); ok {
		c.CardStyle = v
	}
	if v, ok := m["title"].(string); ok {
		c.Title = v
	}
	if v, ok := m["rarity"].(string); ok {
		c.Rarity = v
	}
	if v, ok := m["set"].(string); ok {
		c.Set = v
	}
	if v, ok := m["artist"].(string); ok {
		c.Artist = v
	}
	if v, ok := m["rules_text"].(string); ok {
		c.RulesText = v
	}
	if v, ok := m["flavor_text"].(string); ok {
		c.FlavorText = v
	}
	// Apply defaults.
	if c.TCG == "" {
		c.TCG = "mtg"
	}
	if c.CardStyle == "" {
		c.CardStyle = "basic"
	}
	if c.Title == "" {
		c.Title = "Preview Card"
	}
	return c
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func contentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}


