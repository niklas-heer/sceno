package theme

import (
	"testing"

	"github.com/niklas-heer/sceno/internal/model"
)

func TestResolveDark(t *testing.T) {
	p := Resolve(model.ThemeConfig{Mode: "dark"})
	if p.FgPrimary != "#fafafa" {
		t.Fatalf("dark fg = %q", p.FgPrimary)
	}
}

func TestResolveTransparent(t *testing.T) {
	p := Resolve(model.ThemeConfig{Transparent: true})
	if p.BgCanvas != "none" {
		t.Fatalf("transparent bg = %q", p.BgCanvas)
	}
}

func TestResolveCustomVar(t *testing.T) {
	p := Resolve(model.ThemeConfig{Vars: map[string]string{"card": "#112233"}})
	if p.BgCard != "#112233" {
		t.Fatalf("custom card = %q", p.BgCard)
	}
}
