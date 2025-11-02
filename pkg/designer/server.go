package designer

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/tcg-cardgen/pkg/templates"
	"github.com/Merith-TK/tcg-cardgen/pkg/types"
)

// Embed static files into the binary
//
//go:embed static/*
var staticFiles embed.FS

// Server represents the cardstyle designer web server
type Server struct {
	templateManager *templates.Manager
	htmlTemplate    *template.Template
}

// NewServer creates a new designer server
func NewServer() (*Server, error) {
	// Create template manager
	templateManager := templates.NewManager("")

	// Parse HTML templates
	htmlTemplate, err := template.ParseFS(staticFiles, "static/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML templates: %v", err)
	}

	return &Server{
		templateManager: templateManager,
		htmlTemplate:    htmlTemplate,
	}, nil
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route the request
	path := r.URL.Path

	switch {
	case path == "/" || path == "/index.html":
		s.handleIndex(w, r)
	case strings.HasPrefix(path, "/api/"):
		s.handleAPI(w, r)
	case strings.HasPrefix(path, "/static/"):
		s.handleStatic(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleIndex serves the main designer page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	data := struct {
		Title   string
		Version string
	}{
		Title:   "TCG Cardstyle Designer",
		Version: "1.0.0",
	}

	if err := s.htmlTemplate.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleStatic serves static files (CSS, JS, images)
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Remove /static/ prefix and serve from embedded filesystem
	path := strings.TrimPrefix(r.URL.Path, "/")
	
	data, err := staticFiles.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set content type based on file extension
	ext := filepath.Ext(path)
	contentType := getContentType(ext)
	w.Header().Set("Content-Type", contentType)
	
	w.Write(data)
}

// handleAPI handles REST API endpoints
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	
	switch {
	case path == "/templates" && r.Method == "GET":
		s.apiListTemplates(w, r)
	case path == "/template" && r.Method == "GET":
		s.apiGetTemplate(w, r)
	case path == "/template" && r.Method == "POST":
		s.apiSaveTemplate(w, r)
	case path == "/template/export" && r.Method == "POST":
		s.apiExportTemplate(w, r)
	case path == "/preview" && r.Method == "POST":
		s.apiGeneratePreview(w, r)
	default:
		http.Error(w, "API endpoint not found", http.StatusNotFound)
	}
}

// apiListTemplates returns all available templates
func (s *Server) apiListTemplates(w http.ResponseWriter, r *http.Request) {
	cardstyles, err := s.templateManager.ListAvailableCardstyles()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	// Group by TCG for easier frontend consumption
	grouped := make(map[string][]types.CardStyleInfo)
	for _, cardstyle := range cardstyles {
		if grouped[cardstyle.TCG] == nil {
			grouped[cardstyle.TCG] = make([]types.CardStyleInfo, 0)
		}
		grouped[cardstyle.TCG] = append(grouped[cardstyle.TCG], types.CardStyleInfo{
			TCG:         cardstyle.TCG,
			Name:        cardstyle.Name,
			DisplayName: cardstyle.DisplayName,
			Description: cardstyle.Description,
			Version:     cardstyle.Version,
			Source:      cardstyle.Source,
			Extends:     cardstyle.Extends,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grouped)
}

// apiGetTemplate returns a specific template for editing
func (s *Server) apiGetTemplate(w http.ResponseWriter, r *http.Request) {
	tcg := r.URL.Query().Get("tcg")
	cardstyle := r.URL.Query().Get("cardstyle")

	if tcg == "" || cardstyle == "" {
		http.Error(w, "Missing tcg or cardstyle parameter", http.StatusBadRequest)
		return
	}

	template, err := s.templateManager.LoadTemplate(tcg, cardstyle)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load template: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

// apiSaveTemplate saves a template (placeholder for now)
func (s *Server) apiSaveTemplate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement template saving
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Template saving not yet implemented",
	})
}

// apiExportTemplate exports a template as YAML (placeholder for now)
func (s *Server) apiExportTemplate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement template export
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Template export not yet implemented",
	})
}

// apiGeneratePreview generates a preview image (placeholder for now)
func (s *Server) apiGeneratePreview(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement preview generation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Preview generation not yet implemented",
	})
}

// getContentType returns the appropriate content type for a file extension
func getContentType(ext string) string {
	switch ext {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}