package scene

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/icons"
	"github.com/niklas-heer/sceno/internal/measure"
	"github.com/niklas-heer/sceno/internal/model"
)

func ruleIcons(ctx ruleContext) []Finding {
	var out []Finding
	for _, n := range ctx.d.Nodes {
		if n.Icon == "" || model.IsContainer(n.Kind) {
			continue
		}
		if !icons.Has(n.Icon) {
			out = append(out, Finding{
				RuleID: "icons", Severity: "error", Plane: PlaneLabel,
				Code: string(diag.CodeUnknownIcon),
				Message: fmt.Sprintf("node %q uses unknown icon %q", n.ID, n.Icon),
				Fix:     "Run sceno docs icons --json for allowed names and suggested shapes.",
				Items:   []string{n.ID},
			})
			continue
		}
		if n.Rect.W < measure.IconColumn+measure.PadX+24 {
			out = append(out, Finding{
				RuleID: "icons", Severity: "warning", Plane: PlaneLabel,
				Code: string(diag.CodeTextOverflow),
				Message: fmt.Sprintf("node %q is narrow (%.0fpx) for icon + label — icon may crowd text", n.ID, n.Rect.W),
				Fix:     "Widen with w=, shorten label, or use iconPos=top on tall cards.",
				Items:   []string{n.ID},
			})
		}
		pos := n.IconPos
		if pos == "" {
			pos = model.IconTopLeft
		}
		if pos == model.IconTopLeft && len(strings.Split(n.Label, "\n")) > 2 && n.Rect.H < 72 {
			out = append(out, Finding{
				RuleID: "icons", Severity: "hint", Plane: PlaneLabel,
				Code: string(diag.CodeSuggestAnnotation),
				Message: fmt.Sprintf("node %q has multi-line label with iconPos=top-left — try iconPos=top", n.ID),
				Fix:     "iconPos=top stacks icon above centered label on narrow cards.",
				Items:   []string{n.ID},
			})
		}
		ix, iy := measure.IconRect(n, measure.IconSize)
		iconBox := model.Rect{X: ix, Y: iy, W: measure.IconSize, H: measure.IconSize}
		if overlapsLabel(iconBox, n) {
			out = append(out, Finding{
				RuleID: "icons", Severity: "warning", Plane: PlaneLabel,
				Code: string(diag.CodeMisaligned),
				Message: fmt.Sprintf("node %q icon overlaps label region", n.ID),
				Fix:     "Increase node size, shorten label, or set iconPos=top.",
				Items:   []string{n.ID},
			})
		}
	}
	return out
}

func overlapsLabel(icon model.Rect, n model.Node) bool {
	layout := measure.LabelLayoutFor(n)
	labelTop := layout.ContentY + layout.IconOffsetY
	labelBox := model.Rect{
		X: layout.ContentX,
		Y: labelTop,
		W: layout.ContentW,
		H: n.Rect.Bottom() - labelTop,
	}
	return rectsOverlap(icon, labelBox, 2)
}
