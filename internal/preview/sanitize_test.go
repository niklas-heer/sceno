package preview

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/diag"
	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"
	"github.com/niklas-heer/sceno/internal/repair"
)

func TestSafeSVGRemovesActiveContentAndExternalResources(t *testing.T) {
	source := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" onload="alert(1)">
 <script>alert(1)</script><foreignObject><div xmlns="http://www.w3.org/1999/xhtml">HTML</div></foreignObject>
 <animate attributeName="href" values="javascript:alert(1)"/>
 <image href="https://example.com/a.png"/><use xlink:href="https://example.com/a.svg#x"/>
 <path d="M 0 0 L 10 10" fill="url(https://example.com/paint)" stroke="url(javascript:alert(1))" onclick="alert(1)"/>
 <rect fill="red; background:url(https://example.com)" style="fill:url(https://example.com)"/>
 <style>@import url(https://example.com/evil.css);</style>
 <text>Safe &amp; visible</text></svg>`
	got := safeSVG(source)
	if got == "" {
		t.Fatal("sanitizer dropped valid containing SVG")
	}
	for _, bad := range []string{"script", "foreignObject", "animate", "image", "onload", "onclick", "https://example.com", "javascript:", "style="} {
		if strings.Contains(got, bad) {
			t.Fatalf("unsafe %q retained: %s", bad, got)
		}
	}
	if !strings.Contains(got, "Safe &amp; visible") || !strings.Contains(got, `d="M 0 0 L 10 10"`) {
		t.Fatalf("valid geometry/text removed: %s", got)
	}
}

func TestSafeSVGPreservesRendererGeometryAndEmbeddedFonts(t *testing.T) {
	d := model.Diagram{Nodes: []model.Node{{ID: "a", Kind: model.ShapeBox, Label: "A", Rect: model.Rect{X: 20, Y: 20, W: 100, H: 80}}}}
	raw := render.PolishedSVG(d)
	got := safeSVG(raw)
	if got == "" || !strings.Contains(got, "data:font/ttf;base64,") || !strings.Contains(got, "font-family:Inter") {
		t.Fatal("renderer SVG lost embedded Inter fonts")
	}
	for _, needle := range []string{`viewBox=`, `<rect`, `<text`, `width="100.0"`} {
		if !strings.Contains(got, needle) {
			t.Fatalf("renderer geometry missing %q", needle)
		}
	}
	if err := xml.Unmarshal([]byte(got), new(any)); err != nil {
		t.Fatalf("sanitized SVG is invalid XML: %v", err)
	}
	if !strings.Contains(got, fonts.SVGStyle()) {
		t.Fatal("embedded stylesheet changed")
	}
}

func TestSafeSVGAllowsOnlyInternalPaintAndUseReferences(t *testing.T) {
	got := safeSVG(`<svg><defs><linearGradient id="paint"><stop offset="0" stop-color="#fff"/></linearGradient><path id="icon" d="M0 0"/></defs><rect fill="url(#paint)" filter="url('#shadow')"/><use href="#icon"/><path fill="u&#114;l(https://example.com)" stroke="\75rl(#paint)"/></svg>`)
	if got == "" || !strings.Contains(got, `fill="url(#paint)"`) || !strings.Contains(got, `href="#icon"`) {
		t.Fatalf("internal reference removed: %s", got)
	}
	if strings.Contains(got, "example.com") || strings.Contains(got, `\75`) {
		t.Fatalf("obfuscated paint retained: %s", got)
	}
}

func TestSafeSVGFailsClosedOnMalformedXML(t *testing.T) {
	for _, source := range []string{
		`<svg><rect></svg>`, `<svg/><svg/>`, `<svg onload="a" onload="b"/>`, `<svg><text>&evil;</text></svg>`,
		`<!DOCTYPE svg [<!ENTITY x "x">]><svg/>`, `<svg xmlns="http://www.w3.org/1999/xhtml"><script>x</script></svg>`, `<svg/>trailing`,
	} {
		if got := safeSVG(source); got != "" {
			t.Fatalf("unsafe/malformed XML accepted %q => %s", source, got)
		}
	}
}

func TestSafeSVGRejectsNamespaceAndCSSBypasses(t *testing.T) {
	got := safeSVG(`<svg xmlns:html="http://www.w3.org/1999/xhtml"><html:script>x</html:script><g xmlns="http://www.w3.org/1999/xhtml"><script>x</script></g><rect ONLOAD="x" fill="var(--untrusted)"/><style>text{fill:red}/* external arbitrary CSS */</style><use href="&#x6a;avascript:alert(1)"/></svg>`)
	if got == "" {
		t.Fatal("outer SVG should survive")
	}
	for _, bad := range []string{"script", "ONLOAD", "var(", "style", "javascript"} {
		if strings.Contains(got, bad) {
			t.Fatalf("bypass survived: %s", got)
		}
	}
}

func TestPreviewStateIssueIDWorksForInMemoryRepair(t *testing.T) {
	source := []byte("diagram layout=free gap=20 {\n shape box a \"A\" x=0 y=150 w=100 h=80\n shape box b \"B\" x=60 y=150 w=100 h=80\n}\n")
	state, _ := buildState(source, "/actual/project/diagram.kdl")
	for _, issue := range state.Issues {
		if issue.Code == diag.CodeCollision && len(issue.Repairs) > 0 {
			candidate := issue.Repairs[0]
			preview, err := repair.Preview(source, repair.Request{IssueID: issue.ID, SlideIndex: issue.SlideIndex, Target: candidate.Target, Properties: candidate.Properties})
			if err != nil || !preview.Applicable {
				t.Fatalf("state evidence rejected in memory: %v %+v", err, preview)
			}
			return
		}
	}
	t.Fatal("no collision in state")
}
