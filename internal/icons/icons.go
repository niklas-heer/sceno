package icons

import (
	"fmt"
	"strings"
)

// paths maps icon id → SVG fragment (filled by catalog.go init).
var paths map[string]string

// Names returns icon ids.
func Names() []string {
	return SortedIDs()
}

// Has reports whether an icon exists.
func Has(name string) bool {
	_, ok := paths[name]
	return ok
}

func fmtFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", f), "0"), ".")
}
