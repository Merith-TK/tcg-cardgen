package types

// Config holds runtime configuration for the card generator.
type Config struct {
	TemplateDir  string
	OutputDir    string
	ValidateOnly bool
	Verbose      bool
	Concurrency  int  // number of parallel render workers (0 = runtime.NumCPU)
	JSONLog      bool // emit JSON-structured log lines (for CI/GitHub Actions)
}

// CardStyleInfo describes a discovered card style template.
type CardStyleInfo struct {
	TCG         string
	Name        string
	DisplayName string
	Description string
	Version     string
	// Source is "embedded", "workspace", "user", or an absolute file path.
	Source  string
	Extends string
}

// Dimensions defines output image size.
type Dimensions struct {
	Width  int `yaml:"width"  json:"width"`
	Height int `yaml:"height" json:"height"`
	DPI    int `yaml:"dpi"    json:"dpi"`
}

// Region defines a rectangular area within a card image.
type Region struct {
	X      int `yaml:"x"      json:"x"`
	Y      int `yaml:"y"      json:"y"`
	Width  int `yaml:"width"  json:"width"`
	Height int `yaml:"height" json:"height"`
}

// Font defines text rendering properties for a layer.
type Font struct {
	// Family is resolved via FontLoader: template fonts/ dir → system fonts → monospace fallback.
	Family string `yaml:"family,omitempty" json:"family,omitempty"`
	// Size can be a number or a {{variable}} expression.
	Size   interface{} `yaml:"size"            json:"size"`
	Weight string      `yaml:"weight,omitempty" json:"weight,omitempty"`
	Style  string      `yaml:"style,omitempty"  json:"style,omitempty"`
	Color  string      `yaml:"color"            json:"color"`
}

// Layer represents a single rendering layer in a card template.
type Layer struct {
	Name         string      `yaml:"name"                   json:"name"`
	Role         string      `yaml:"role,omitempty"         json:"role,omitempty"`
	Type         string      `yaml:"type"                   json:"type"`
	Source       string      `yaml:"source,omitempty"       json:"source,omitempty"`
	Content      string      `yaml:"content,omitempty"      json:"content,omitempty"`
	Region       Region      `yaml:"region"                 json:"region"`
	Font         *Font       `yaml:"font,omitempty"         json:"font,omitempty"`
	FitMode      string      `yaml:"fit_mode,omitempty"     json:"fit_mode,omitempty"`
	IconReplace  bool        `yaml:"icon_replace,omitempty" json:"icon_replace,omitempty"`
	StripHeaders bool        `yaml:"strip_headers,omitempty" json:"strip_headers,omitempty"`
	// Condition is an expression string; layer is skipped when it evaluates falsy.
	Condition    string      `yaml:"condition,omitempty"    json:"condition,omitempty"`
	Align        string      `yaml:"align,omitempty"        json:"align,omitempty"`
	Fallback     string      `yaml:"fallback,omitempty"     json:"fallback,omitempty"`
	TextType     string      `yaml:"text_type,omitempty"    json:"text_type,omitempty"`
	Shape        string      `yaml:"shape,omitempty"        json:"shape,omitempty"`
	Fill         string      `yaml:"fill,omitempty"         json:"fill,omitempty"`
	Stroke       string      `yaml:"stroke,omitempty"       json:"stroke,omitempty"`
	StrokeWidth  float64     `yaml:"stroke_width,omitempty" json:"stroke_width,omitempty"`
	CornerRadius float64     `yaml:"corner_radius,omitempty" json:"corner_radius,omitempty"`
	Points       [][]float64 `yaml:"points,omitempty"       json:"points,omitempty"`
}

// LayerOverride describes field patches applied to a named base layer.
type LayerOverride struct {
	Layer   string                 `yaml:"layer"    json:"layer"`
	Updates map[string]interface{} `yaml:",inline"  json:"updates,omitempty"`
}
