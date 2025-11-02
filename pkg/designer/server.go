package designer

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Merith-TK/tcg-cardgen/pkg/metadata"
	"github.com/Merith-TK/tcg-cardgen/pkg/renderer"
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
	// Add CORS headers to allow React frontend to connect
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api")

	switch {
	case path == "/templates" && r.Method == "GET":
		s.apiListTemplates(w, r)
	case path == "/template" && r.Method == "GET":
		s.apiGetTemplate(w, r)
	case path == "/template" && r.Method == "POST":
		s.apiSaveTemplate(w, r)
	case path == "/template/save" && r.Method == "POST":
		s.apiSaveTemplate(w, r)
	case path == "/template/export" && r.Method == "POST":
		s.apiExportTemplate(w, r)
	case path == "/render" && r.Method == "POST":
		s.apiRenderCard(w, r)
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

// apiSaveTemplate saves a template
func (s *Server) apiSaveTemplate(w http.ResponseWriter, r *http.Request) {
	var templateData struct {
		Name   string               `json:"name"`
		Format templates.Dimensions `json:"format"`
		Layers []templates.Layer    `json:"layers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&templateData); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// For now, just return success - actual saving would go here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Template '%s' saved successfully", templateData.Name),
	})
}

// apiRenderCard renders a card and returns the image
func (s *Server) apiRenderCard(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Template struct {
			Name   string               `json:"name"`
			Format templates.Dimensions `json:"format"`
			Layers []templates.Layer    `json:"layers"`
		} `json:"template"`
		Card map[string]interface{} `json:"card"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to proper template structure
	template := &templates.Template{
		Name:       requestData.Template.Name,
		Dimensions: requestData.Template.Format,
		Layers:     requestData.Template.Layers,
	}

	// Create renderer
	rendererInstance := renderer.NewRenderer()

	// Create a simple card with the provided data
	cardData := map[string]interface{}{}
	for k, v := range requestData.Card {
		cardData[k] = v
	}

	// Create a metadata card from the request data
	card := &metadata.Card{
		Metadata: cardData,
	}

	// Create a temporary file to render to
	tempFile := fmt.Sprintf("temp_render_%d.png", time.Now().UnixNano())

	// Render the card to temporary file
	err := rendererInstance.RenderCard(card, template, tempFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render card: %v", err), http.StatusInternalServerError)
		return
	}

	// Read the rendered file
	imageData, err := os.ReadFile(tempFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read rendered image: %v", err), http.StatusInternalServerError)
		return
	}

	// Clean up temp file
	os.Remove(tempFile)

	// Return the image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.png\"", requestData.Template.Name))
	w.Write(imageData)
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
