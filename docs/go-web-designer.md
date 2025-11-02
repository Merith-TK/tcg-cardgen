# Go-Hosted Cardstyle Designer

Secure web-based cardstyle editor with zero npm dependencies.

## 🏗️ Architecture

### Technology Stack
- **Backend**: Go 1.19+ with embedded static assets
- **Frontend**: Vanilla JS + HTML5 Canvas + CSS Grid
- **Build**: Go embed for static files (no webpack/npm)
- **Security**: No external dependencies, controlled supply chain

### Project Structure
```
cmd/
├── tcg-cardgen/           # CLI tool (existing)
└── cardstyle-designer/    # Web designer server
    ├── main.go           # Web server entry point
    ├── handlers/         # HTTP handlers
    ├── api/             # REST API endpoints
    └── static/          # Embedded web assets
        ├── index.html
        ├── css/
        ├── js/
        └── assets/

pkg/
├── designer/            # Designer-specific logic
│   ├── server.go       # HTTP server setup
│   ├── api.go          # API handlers
│   └── export.go       # Template export/import
└── web/                # Web utilities
    ├── templates.go    # Go HTML templates
    └── assets.go       # Static asset embedding
```

## 🔒 Security Benefits

### Go vs npm Ecosystem
| Aspect | Go Modules | npm Packages |
|--------|------------|--------------|
| Dependencies | Vetted, minimal | Sprawling tree |
| Supply Chain | Cryptographic checksums | Trust-based |
| Install Time | Compile-time checked | Runtime surprises |
| Malware Vector | Very low | High (Shai-halud style) |

### Our Approach
- ✅ **Zero npm dependencies**: All JS written in-house
- ✅ **Embedded assets**: No CDN dependencies
- ✅ **Single binary**: No runtime installations
- ✅ **Controlled supply chain**: Only Go standard library + vetted modules

## 🎨 Frontend Architecture

### Vanilla JavaScript Structure
```javascript
// No frameworks, just clean modular JS
class CardstyleDesigner {
  constructor() {
    this.canvas = new CanvasRenderer();
    this.layerPanel = new LayerPanel();
    this.propertiesPanel = new PropertiesPanel();
    this.template = new TemplateManager();
  }
}

class CanvasRenderer {
  // HTML5 Canvas rendering
  // Mouse interaction handling
  // Layer selection/dragging
}

class LayerPanel {
  // Layer list management
  // Drag-to-reorder
  // Visibility toggles
}

class PropertiesPanel {
  // Form-based property editing
  // Real-time updates
  // Color picker components
}
```

### No External Dependencies
- **Canvas API**: Native HTML5 canvas for rendering
- **Drag & Drop**: Native HTML5 drag/drop API
- **Color Picker**: Custom implementation or `<input type="color">`
- **HTTP**: Native `fetch()` API for server communication

## 🚀 Implementation Plan

### Phase 1: Go Web Server Foundation
1. Create web server with embedded static files
2. Basic HTML/CSS layout
3. Canvas element with basic rendering
4. REST API endpoints for template CRUD

### Phase 2: Interactive Canvas
1. Layer rendering on HTML5 canvas
2. Mouse selection and dragging
3. Resize handles for layers
4. Live preview updates

### Phase 3: UI Panels
1. Layer management panel
2. Properties editing panel
3. Template metadata panel
4. Export/import functionality

## 🛠️ Development Start

Let's begin implementation: