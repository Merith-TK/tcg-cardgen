# Cardstyle Designer Technical Specification

## 🎯 Project Overview

The Cardstyle Designer is a web-based visual editor for creating TCG cardstyles without requiring manual YAML editing. Users can drag, drop, and configure layers with real-time preview.

## 🏗️ Architecture

### Frontend (React + TypeScript)
```
src/
├── components/
│   ├── Canvas/
│   │   ├── CardCanvas.tsx          # Main canvas component
│   │   ├── LayerRenderer.tsx       # Individual layer rendering
│   │   ├── SelectionHandles.tsx    # Resize/move handles
│   │   └── GridOverlay.tsx         # Alignment grid
│   ├── Panels/
│   │   ├── LayerPanel.tsx          # Layer list and controls
│   │   ├── PropertiesPanel.tsx     # Layer property editor
│   │   ├── TemplatePanel.tsx       # Template metadata
│   │   └── ToolPanel.tsx           # Add layer tools
│   ├── Controls/
│   │   ├── ColorPicker.tsx         # RGBA color picker
│   │   ├── FontSelector.tsx        # Font family/size controls
│   │   ├── ShapeSelector.tsx       # Shape type selector
│   │   └── RegionEditor.tsx        # X/Y/W/H number inputs
│   └── Export/
│       ├── ExportDialog.tsx        # Package export UI
│       └── PreviewGenerator.tsx    # Generate preview images
├── types/
│   ├── CardstyleTypes.ts          # TypeScript interfaces
│   └── CanvasTypes.ts             # Canvas-specific types
├── utils/
│   ├── YamlGenerator.ts           # Generate YAML from state
│   ├── TemplateLoader.ts          # Load existing templates
│   └── ColorUtils.ts              # RGBA conversion utilities
└── hooks/
    ├── useCanvas.ts               # Canvas state management
    ├── useTemplate.ts             # Template CRUD operations
    └── useExport.ts               # Export/import functionality
```

### Backend (Go Web Server)
```
cmd/cardstyle-designer/
├── main.go                        # Web server entry point
├── handlers/
│   ├── template.go                # Template CRUD API
│   ├── export.go                  # Package export/import
│   ├── preview.go                 # Generate preview images
│   └── assets.go                  # Static file serving
├── models/
│   ├── designer.go                # Designer-specific types
│   └── package.go                 # Package format handling
└── services/
    ├── template_service.go        # Template business logic
    ├── preview_service.go         # Image generation
    └── package_service.go         # Package creation/extraction
```

## 🎨 Core Types

### TypeScript Interfaces
```typescript
interface DesignerLayer {
  id: string;
  name: string;
  type: 'text' | 'shape' | 'image';
  visible: boolean;
  locked: boolean;
  zIndex: number;
  
  // Position and size
  region: {
    x: number;
    y: number;
    width: number;
    height: number;
  };
  
  // Type-specific properties
  textProps?: TextProperties;
  shapeProps?: ShapeProperties;
  imageProps?: ImageProperties;
  
  // Common properties
  condition?: string;
  role?: string;
}

interface TextProperties {
  textType: 'static' | 'dynamic';
  content: string;
  font: {
    family: string;
    size: number;
    weight: 'normal' | 'bold';
    style: 'normal' | 'italic';
    color: string; // RGBA hex
  };
  align: 'left' | 'center' | 'right';
  iconReplace: boolean;
  stripHeaders: boolean;
}

interface ShapeProperties {
  shape: 'rectangle' | 'circle' | 'ellipse' | 'polygon';
  fill: string;           // RGBA hex
  stroke: string;         // RGBA hex
  strokeWidth: number;
  cornerRadius?: number;  // For rectangles
  points?: number[][];    // For polygons
}

interface ImageProperties {
  source: string;
  fallback?: string;
  fitMode: 'fill' | 'fit' | 'stretch' | 'center';
}

interface CardstyleTemplate {
  name: string;
  tcg: string;
  version: string;
  description: string;
  extends?: string;
  
  dimensions: {
    width: number;
    height: number;
    dpi: number;
  };
  
  layers: DesignerLayer[];
  requiredFields: string[];
  optionalFields: Record<string, any>;
  styleTokens: Record<string, string>;
  icons: Record<string, string>;
}
```

## 🛠️ Core Features

### 1. Interactive Canvas
- **Drag & Drop**: Move layers by dragging
- **Resize Handles**: Corner/edge handles for resizing
- **Multi-selection**: Select multiple layers with Ctrl+click
- **Alignment Tools**: Snap to grid, align to other layers
- **Zoom Controls**: Zoom in/out for precision editing

### 2. Layer Management
- **Layer List**: Hierarchical list with eye (visibility) and lock icons
- **Reordering**: Drag layers to change z-order
- **Grouping**: Group related layers together
- **Duplication**: Duplicate layers with Ctrl+D

### 3. Property Editing
- **Live Preview**: Changes appear immediately on canvas
- **Color Picker**: Full RGBA color picker with transparency
- **Font Controls**: Family, size, weight, style selectors
- **Content Editor**: Rich text editor for static content
- **Variable Autocomplete**: Suggest {{card.title}} etc. for dynamic content

### 4. Template Management
- **New Template**: Start from scratch or copy existing
- **Load Template**: Import existing YAML templates
- **Save Template**: Generate and download YAML
- **Template Validation**: Real-time validation feedback

## 🎯 Key UI Components

### Canvas Component
```typescript
interface CanvasProps {
  template: CardstyleTemplate;
  selectedLayers: string[];
  onLayerSelect: (layerId: string, multiSelect: boolean) => void;
  onLayerMove: (layerId: string, newRegion: Region) => void;
  onLayerResize: (layerId: string, newRegion: Region) => void;
}

const CardCanvas: React.FC<CanvasProps> = ({
  template,
  selectedLayers,
  onLayerSelect,
  onLayerMove,
  onLayerResize
}) => {
  // Render each layer
  // Handle mouse interactions
  // Show selection handles
  // Implement drag/resize logic
};
```

### Layer Panel
```typescript
interface LayerPanelProps {
  layers: DesignerLayer[];
  selectedLayers: string[];
  onLayerSelect: (layerId: string) => void;
  onLayerReorder: (fromIndex: number, toIndex: number) => void;
  onLayerToggleVisible: (layerId: string) => void;
  onLayerDelete: (layerId: string) => void;
  onLayerAdd: (type: LayerType) => void;
}
```

### Properties Panel
```typescript
interface PropertiesPanelProps {
  selectedLayer: DesignerLayer | null;
  onLayerUpdate: (layerId: string, updates: Partial<DesignerLayer>) => void;
}
```

## 🔄 Workflow

### 1. Create New Template
```
1. User clicks "New Template"
2. Select TCG type (MTG, Pokemon, Custom)
3. Choose base template or start blank
4. Set dimensions and metadata
5. Begin adding layers
```

### 2. Add Layers
```
1. Click "Add Layer" button
2. Choose layer type (Text/Shape/Image)
3. Layer appears on canvas with default properties
4. User drags to position, resizes as needed
5. Properties panel updates with layer settings
```

### 3. Edit Properties
```
1. Select layer on canvas or in layer panel
2. Properties panel shows layer-specific controls
3. Changes update live on canvas
4. Validation feedback for invalid values
```

### 4. Export Template
```
1. Click "Export" button
2. Choose export format:
   - YAML file only
   - .tcgpack with assets
   - Share online (if implemented)
3. Download generated package
```

## 🎨 Canvas Rendering Strategy

### Fabric.js Implementation
```typescript
// Create Fabric.js canvas
const canvas = new fabric.Canvas('cardstyle-canvas', {
  width: 750,
  height: 1050,
  backgroundColor: '#ffffff'
});

// Add layer as Fabric object
const addTextLayer = (layer: DesignerLayer) => {
  const text = new fabric.Textbox(layer.textProps.content, {
    left: layer.region.x,
    top: layer.region.y,
    width: layer.region.width,
    height: layer.region.height,
    fontFamily: layer.textProps.font.family,
    fontSize: layer.textProps.font.size,
    fill: layer.textProps.font.color
  });
  
  canvas.add(text);
};

const addShapeLayer = (layer: DesignerLayer) => {
  let shape: fabric.Object;
  
  switch (layer.shapeProps.shape) {
    case 'rectangle':
      shape = new fabric.Rect({
        left: layer.region.x,
        top: layer.region.y,
        width: layer.region.width,
        height: layer.region.height,
        fill: layer.shapeProps.fill,
        stroke: layer.shapeProps.stroke,
        strokeWidth: layer.shapeProps.strokeWidth,
        rx: layer.shapeProps.cornerRadius,
        ry: layer.shapeProps.cornerRadius
      });
      break;
      
    case 'circle':
      shape = new fabric.Circle({
        left: layer.region.x,
        top: layer.region.y,
        radius: Math.min(layer.region.width, layer.region.height) / 2,
        fill: layer.shapeProps.fill,
        stroke: layer.shapeProps.stroke,
        strokeWidth: layer.shapeProps.strokeWidth
      });
      break;
  }
  
  canvas.add(shape);
};
```

## 📦 Package Format

### .tcgpack Structure
```
my-cardstyle.tcgpack (ZIP file)
├── manifest.yaml
├── templates/
│   ├── mtg/
│   │   ├── awesome.yaml
│   │   └── awesome_token.yaml
│   └── pokemon/
│       └── custom.yaml
├── assets/
│   ├── frames/
│   ├── icons/
│   ├── fonts/
│   └── artwork/
└── previews/
    ├── mtg_awesome.png
    └── pokemon_custom.png
```

### Installation Commands
```bash
# Install package to user directory
tcg-cardgen install my-cardstyle.tcgpack

# Install to workspace
tcg-cardgen install --workspace my-cardstyle.tcgpack

# List installed packages
tcg-cardgen packages list

# Uninstall package
tcg-cardgen packages remove my-cardstyle
```

## 🚀 Implementation Phases

### Phase 1: Core Canvas (Week 1)
- Basic React app with canvas
- Layer rendering (text, basic shapes)
- Selection and drag/drop
- Properties panel

### Phase 2: Advanced Features (Week 2)
- Shape tools (polygon, circles)
- Color picker with RGBA support
- Font controls
- Layer panel with reordering

### Phase 3: Template System (Week 3)
- YAML generation/loading
- Template validation
- Variable autocomplete
- Preview generation

### Phase 4: Package System (Week 4)
- Export to .tcgpack format
- Install/uninstall commands
- Asset management
- Online sharing (optional)

## 🎯 Success Metrics

- **Ease of Use**: Non-programmers can create templates
- **Feature Parity**: Can recreate existing YAML templates
- **Performance**: Smooth 60fps canvas interactions
- **Adoption**: Community creates and shares templates

This would revolutionize TCG cardstyle creation, making it accessible to designers who aren't comfortable with YAML! 🎨✨