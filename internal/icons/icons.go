package icons

import (
	"fmt"
	"strings"
)

// SVG path data (24x24 viewBox) for inline rendering.
var paths = map[string]string{
	"cloud":    `<path d="M17.5 19H9a7 7 0 1 1-.5-14 9 9 0 0 1 9 9 2.5 2.5 0 0 1 0 5Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"database": `<ellipse cx="12" cy="5.5" rx="7" ry="3" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M5 5.5v13c0 1.66 3.13 3 7 3s7-1.34 7-3v-13" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M5 12c0 1.66 3.13 3 7 3s7-1.34 7-3" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"server":   `<rect x="4" y="4" width="16" height="6" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><rect x="4" y="14" width="16" height="6" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="8" cy="7" r=".75" fill="currentColor"/><circle cx="8" cy="17" r=".75" fill="currentColor"/>`,
	"user":     `<circle cx="12" cy="8" r="3.5" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M5 20c0-3.87 3.13-7 7-7s7 3.13 7 7" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"users":    `<path d="M9 12a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M16 13a3 3 0 1 0 0-6" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M3 20c0-2.76 2.69-5 6-5M13 20c0-2.2 2.24-4 5-4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"lock":     `<rect x="6" y="11" width="12" height="9" rx="2" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M8 11V8a4 4 0 0 1 8 0v3" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"shield":   `<path d="M12 3 5 6v6c0 4.42 3 7.56 7 8 4-0.44 7-3.58 7-8V6l-7-3Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/>`,
	"workflow": `<circle cx="6" cy="12" r="2.5" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="18" cy="6" r="2.5" fill="none" stroke="currentColor" stroke-width="1.75"/><circle cx="18" cy="18" r="2.5" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M8.5 11 15 7M8.5 13 15 17" fill="none" stroke="currentColor" stroke-width="1.75"/>`,
	"queue":    `<path d="M5 7h14M5 12h14M5 17h10" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"api":      `<path d="M8 8 4 12l4 4M16 8l4 4-4 4M14 5l-4 14" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"storage":  `<path d="M4 7V5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v2M4 7v10a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M10 11h4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
	"k8s":      `<path d="M12 3 4 7v10l8 4 8-4V7l-8-4Z" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/><circle cx="12" cy="12" r="2" fill="currentColor"/>`,
	"policy":   `<path d="M12 3 5 6v5c0 3.5 3 6 7 7 4-1 7-3.5 7-7V6l-7-3Z" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M9 12l2 2 4-4" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"/>`,
	"info":     `<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="1.75"/><path d="M12 11v5M12 8h.01" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>`,
}

// Names returns icon ids.
func Names() []string {
	out := make([]string, 0, len(paths))
	for k := range paths {
		out = append(out, k)
	}
	return out
}

// Has reports whether an icon exists.
func Has(name string) bool {
	_, ok := paths[name]
	return ok
}

func fmtFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", f), "0"), ".")
}
