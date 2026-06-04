package model

// Rect is an axis-aligned bounding box in canvas coordinates.
type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Right() float64  { return r.X + r.W }
func (r Rect) Bottom() float64 { return r.Y + r.H }
func (r Rect) CX() float64     { return r.X + r.W / 2 }
func (r Rect) CY() float64     { return r.Y + r.H / 2 }

// ShapeKind determines rendering and anchor geometry.
type ShapeKind string

// Side is which border an edge attaches to.
type Side string

const (
	SideTop    Side = "top"
	SideRight  Side = "right"
	SideBottom Side = "bottom"
	SideLeft   Side = "left"
	SideAuto   Side = "auto"
)

// LayoutMode controls placement strategy.
type LayoutMode string

const (
	LayoutAuto   LayoutMode = "auto"
	LayoutFree   LayoutMode = "free"
	LayoutHybrid LayoutMode = "hybrid"
)

// RenderStyle selects sketch vs polished output.
type RenderStyle string

const (
	StyleSketch   RenderStyle = "sketch"
	StylePolished RenderStyle = "polished"
)

// ThemeConfig controls colors, dark mode, transparency, and CSS/SVG variables.
type ThemeConfig struct {
	Mode        string            `json:"mode,omitempty"`        // light | dark
	Transparent bool              `json:"transparent,omitempty"` // canvas/slide background
	Vars        map[string]string `json:"vars,omitempty"`        // e.g. background, foreground, card
}

// Node is a placed diagram element.
type Node struct {
	ID       string
	Label    string
	Subtitle string
	Kind     ShapeKind
	Icon     string // catalog name, e.g. cloud, database
	CodeLang string // for shape code
	Code     string // source body
	Fill     string
	Stroke   string
	Accent   string // infobox left stripe
	FontSize float64
	Layer    int
	Row      int
	Column   int
	Fixed    bool
	Parent   string
	Rect     Rect
}

// Edge connects two nodes with optional anchor sides.
type Edge struct {
	From, To       string
	FromSide, ToSide Side
	Dashed         bool
	Color          string
}

// RoutedEdge is a laid-out connector.
type RoutedEdge struct {
	Edge   Edge
	Key    string
	Points [][]float64
}

// Diagram is the fully resolved scene.
type Diagram struct {
	Title       string
	Subtitle    string
	Layout      LayoutMode
	Style       RenderStyle
	Gap         float64
	Padding     float64
	SlideAspect string // e.g. 16x9 — slide export framing
	Theme       ThemeConfig
	Nodes       []Node
	Edges       []Edge
	Routed      []RoutedEdge
	EdgePaths   map[string][][]float64 // legacy access by key
}

// SlideSpec is one declarative slide (PowerPoint-like deck from KDL).
type SlideSpec struct {
	Title string
	Nodes []NodeSpec
	Edges []EdgeSpec
}

// Deck is a multi-slide document after layout.
type Deck struct {
	Title       string
	Subtitle    string
	SlideAspect string
	Theme       ThemeConfig
	Slides      []Diagram
}

// Spec is the input document (KDL).
type Spec struct {
	Title       string       `yaml:"title" json:"title"`
	Subtitle    string       `yaml:"subtitle" json:"subtitle"`
	Layout      LayoutMode   `yaml:"layout" json:"layout"`
	Style       RenderStyle  `yaml:"style" json:"style"`
	Gap         float64      `yaml:"gap" json:"gap"`
	Padding     float64      `yaml:"padding" json:"padding"`
	SlideAspect string       `yaml:"slideAspect" json:"slideAspect"` // slide=16x9
	Theme       ThemeConfig  `yaml:"theme" json:"theme"`
	Slides      []SlideSpec  `yaml:"slides" json:"slides"`
	Nodes       []NodeSpec   `yaml:"nodes" json:"nodes"`
	Edges       []EdgeSpec   `yaml:"edges" json:"edges"`
}

type NodeSpec struct {
	ID       string    `yaml:"id" json:"id"`
	Label    string    `yaml:"label" json:"label"`
	Subtitle string    `yaml:"subtitle" json:"subtitle"`
	Kind     ShapeKind `yaml:"kind" json:"kind"`
	Icon     string    `yaml:"icon" json:"icon"`
	Fill     string    `yaml:"fill" json:"fill"`
	Stroke   string    `yaml:"stroke" json:"stroke"`
	Accent   string    `yaml:"accent" json:"accent"`
	FontSize float64   `yaml:"fontSize" json:"fontSize"`
	Layer    int       `yaml:"layer" json:"layer"`
	Row      int       `yaml:"row" json:"row"`
	Parent   string    `yaml:"parent" json:"parent"`
	W        float64   `yaml:"w" json:"w"`
	H        float64   `yaml:"h" json:"h"`
	X        *float64  `yaml:"x" json:"x"`
	Y        *float64  `yaml:"y" json:"y"`
	CodeLang string    `yaml:"lang" json:"lang,omitempty"`
	Code     string    `yaml:"source" json:"source,omitempty"`
}

type EdgeSpec struct {
	From     string `yaml:"from" json:"from"`
	To       string `yaml:"to" json:"to"`
	FromSide Side   `yaml:"fromSide" json:"fromSide"`
	ToSide   Side   `yaml:"toSide" json:"toSide"`
	Dashed   bool   `yaml:"dashed" json:"dashed"`
	Color    string `yaml:"color" json:"color"`
}

// Collision between nodes.
type Collision struct {
	A, B string
}

// EdgeCollision describes a routing problem.
type EdgeCollision struct {
	EdgeKey string
	Kind    string // node_crossing, edge_crossing
	With    string
}
