package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niklas-heer/sceno/internal/model"
)

// LoadFile reads a .kdl sceno spec.
func LoadFile(path string) (model.Spec, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".kdl" {
		return model.Spec{}, fmt.Errorf("sceno only supports KDL specs (.kdl); got %q — run: sceno init -o sceno.kdl", ext)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Spec{}, err
	}
	return LoadKDL(data)
}

func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func defaults(s *model.Spec) {
	if s.Layout == "" {
		s.Layout = model.LayoutAuto
	}
	if s.Style == "" {
		s.Style = model.StylePolished
	}
	if s.Gap <= 0 {
		s.Gap = 28
	}
	if s.Padding <= 0 {
		s.Padding = 20
	}
	normalizeNode := func(n *model.NodeSpec) {
		n.Kind = model.NormalizeShape(n.Kind)
		if n.Kind == "" {
			n.Kind = model.ShapeBox
		}
		if n.Stroke == "" {
			n.Stroke = "#e2e8f0"
		}
		if n.FontSize <= 0 {
			n.FontSize = 14
		}
		switch n.Kind {
		case model.ShapeInfobox, model.ShapeCallout:
			if n.Accent == "" {
				n.Accent = "#7c3aed"
			}
		case model.ShapeNote:
			if n.Fill == "" {
				n.Fill = "#fef9c3"
			}
		case model.ShapeTextbox:
			if n.Fill == "" {
				n.Fill = "#f8fafc"
			}
		case model.ShapeCode:
			if n.Code == "" && strings.Contains(n.Label, "\n") {
				n.Code = n.Label
			}
		}
	}
	for i := range s.Nodes {
		normalizeNode(&s.Nodes[i])
	}
	for i := range s.Slides {
		for j := range s.Slides[i].Nodes {
			normalizeNode(&s.Slides[i].Nodes[j])
		}
	}
}
