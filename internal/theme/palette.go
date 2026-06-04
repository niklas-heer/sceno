package theme

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/model"
)

// Palette holds resolved colors for one render (light, dark, or custom vars).
type Palette struct {
	Mode        string
	Transparent bool
	BgCanvas    string
	BgCard      string
	BgMuted     string
	BgLane      string
	BgTextbox   string
	BgCode      string
	Border      string
	BorderStrong string
	FgPrimary   string
	FgMuted     string
	Accent      string
	AccentBrand string
	EdgeDefault string
	EdgeOpacity string
	Shadow      string
	Ring        string
	CodeFg      string
	CodeKeyword string
	CodeString  string
	CodeComment string
	CodeNumber  string
}

// LightPalette is the default zinc/shadcn light theme.
func LightPalette() Palette {
	return Palette{
		Mode:         "light",
		BgCanvas:     BgCanvas,
		BgCard:       BgCard,
		BgMuted:      BgMuted,
		BgLane:       BgLane,
		BgTextbox:    BgTextbox,
		BgCode:       "#f4f4f5",
		Border:       Border,
		BorderStrong: BorderStrong,
		FgPrimary:    FgPrimary,
		FgMuted:      FgMuted,
		Accent:       Accent,
		AccentBrand:  AccentBrand,
		EdgeDefault:  EdgeDefault,
		EdgeOpacity:  EdgeOpacity,
		Shadow:       Shadow,
		Ring:         Ring,
		CodeFg:       "#18181b",
		CodeKeyword:  "#7c3aed",
		CodeString:   "#059669",
		CodeComment:  "#71717a",
		CodeNumber:   "#d97706",
	}
}

// DarkPalette is zinc/shadcn dark — tuned for slides and exports.
func DarkPalette() Palette {
	return Palette{
		Mode:         "dark",
		BgCanvas:     "#09090b",
		BgCard:       "#18181b",
		BgMuted:      "#27272a",
		BgLane:       "#0f0f12",
		BgTextbox:    "#27272a",
		BgCode:       "#0c0c0e",
		Border:       "#3f3f46",
		BorderStrong: "#52525b",
		FgPrimary:    "#fafafa",
		FgMuted:      "#a1a1aa",
		Accent:       "#fafafa",
		AccentBrand:  "#a78bfa",
		EdgeDefault:  "#71717a",
		EdgeOpacity:  "0.95",
		Shadow:       "#000000",
		Ring:         "rgba(250,250,250,0.06)",
		CodeFg:       "#e4e4e7",
		CodeKeyword:  "#c4b5fd",
		CodeString:   "#6ee7b7",
		CodeComment:  "#71717a",
		CodeNumber:   "#fcd34d",
	}
}

// Resolve builds a palette from diagram/deck theme config.
func Resolve(cfg model.ThemeConfig) Palette {
	p := LightPalette()
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "dark" {
		p = DarkPalette()
	}
	p.Transparent = cfg.Transparent
	for k, v := range cfg.Vars {
		p.applyVar(k, v)
	}
	if cfg.Transparent {
		p.BgCanvas = "none"
	}
	return p
}

func (p *Palette) applyVar(key, value string) {
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if value == "transparent" {
		value = "none"
	}
	switch key {
	case "background", "bg", "canvas":
		p.BgCanvas = value
	case "card":
		p.BgCard = value
	case "muted":
		p.BgMuted = value
	case "lane":
		p.BgLane = value
	case "textbox":
		p.BgTextbox = value
	case "code", "codebg", "code-background":
		p.BgCode = value
	case "border":
		p.Border = value
	case "foreground", "fg", "text":
		p.FgPrimary = value
	case "muted-foreground", "mutedfg", "muted-text":
		p.FgMuted = value
	case "accent":
		p.AccentBrand = value
	case "primary":
		p.Accent = value
	case "edge":
		p.EdgeDefault = value
	case "code-fg", "codefg":
		p.CodeFg = value
	case "code-keyword":
		p.CodeKeyword = value
	case "code-string":
		p.CodeString = value
	case "code-comment":
		p.CodeComment = value
	case "code-number":
		p.CodeNumber = value
	case "transparent":
		if value == "true" || value == "1" || value == "none" {
			p.Transparent = true
			p.BgCanvas = "none"
		}
	}
}

// CSSVars emits :root custom properties for HTML/slides.
func (p Palette) CSSVars() string {
	bg := p.BgCanvas
	if bg == "none" {
		bg = "transparent"
	}
	return fmt.Sprintf(`:root{
--background:%s;
--foreground:%s;
--card:%s;
--card-foreground:%s;
--muted:%s;
--muted-foreground:%s;
--border:%s;
--ring:%s;
--primary:%s;
--accent:%s;
--code-bg:%s;
--code-fg:%s;
--code-keyword:%s;
--code-string:%s;
--code-comment:%s;
--radius:0.5rem;
}`, bg, p.FgPrimary, p.BgCard, p.FgPrimary, p.BgMuted, p.FgMuted, p.Border, p.Ring, p.Accent, p.AccentBrand,
		p.BgCode, p.CodeFg, p.CodeKeyword, p.CodeString, p.CodeComment)
}

// SlideCSS returns extra rules for code blocks and transparent decks.
func (p Palette) SlideCSS() string {
	innerBg := p.BgCard
	if p.Transparent {
		innerBg = "transparent"
	}
	s := fmt.Sprintf(`
.slide-inner{background:%s}
pre.code-block{margin:0;padding:14px 16px;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:12px;line-height:1.5;background:var(--code-bg);color:var(--code-fg);border-radius:calc(var(--radius));overflow:auto;max-height:100%%}
.code-kw{color:var(--code-keyword)}.code-str{color:var(--code-string)}.code-cm{color:var(--code-comment)}.code-num{color:var(--code-number)}
.slide-layout{display:flex;flex-direction:column;gap:16px;width:100%%;max-width:min(96vw,1200px)}
.slide-layout .slide-inner{flex:1;min-height:0}
.slide-code{width:100%%;max-width:min(96vw,1200px)}
`, innerBg)
	if p.Mode == "dark" {
		s += `body{color-scheme:dark}`
	}
	return s
}
