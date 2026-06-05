// Package theme holds shared visual tokens (shadcn / zinc-inspired) for export parity.
package theme

// Font family name used in SVG, HTML, PDF, and PNG.
const FontFamily = "Inter"

// Colors — zinc palette aligned with shadcn/ui defaults.
const (
	BgCanvas    = "#fafafa" // zinc-50
	BgCard      = "#ffffff"
	BgMuted     = "#f4f4f5" // zinc-100
	BgLane      = "#fafafa"
	BgTextbox   = "#f4f4f5"
	Border      = "#e4e4e7" // zinc-200
	BorderStrong = "#d4d4d8" // zinc-300
	FgPrimary   = "#09090b" // zinc-950
	FgMuted     = "#71717a" // zinc-500
	Accent      = "#18181b" // zinc-900 (primary actions)
	AccentBrand = "#7c3aed" // violet callouts
	EdgeDefault = "#a1a1aa" // zinc-400
	EdgeOpacity = "0.92"
	Shadow      = "#09090b"
	Ring        = "rgba(9,9,11,0.05)"
)

// Type scale (px).
const (
	TitleSize     = 28
	SubtitleSize  = 14
	NodeSize      = 13
	LaneLabelSize = 10
	SubSize       = 11
)

// Radii — shadcn rounded-lg / md.
const (
	RadiusCard    = 8
	RadiusTextbox = 6
	RadiusLane    = 12
	RadiusSlide   = 12
)

// Edge geometry.
const (
	EdgeWidth       = 1.75
	EdgeCorner      = 10.0
	ArrowMarkerSize = 9.0
)

// CSSVars returns :root custom properties for HTML/slides (light default).
func CSSVars() string {
	return LightPalette().CSSVars()
}
