package render

import (
	"encoding/xml"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
	"golang.org/x/image/font"
)

func TestCodeTabsUseFourColumnStops(t *testing.T) {
	for input, want := range map[string]string{
		"\treturn":  "    return",
		"ab\tx":     "ab  x",
		"é\tx":      "é   x",
		"abcd\t\tx": "abcd        x",
	} {
		if got := expandCodeTabs(input); got != want {
			t.Errorf("expandCodeTabs(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestCodeSVGPreservesIndentAndMeasuredUnicodeSpans(t *testing.T) {
	n := model.Node{Kind: model.ShapeCode, Code: "\treturn \"é\" + 12", CodeLang: "go", Rect: model.Rect{X: 40, Y: 80, W: 300, H: 90}}
	useDiagramPalette(model.Diagram{})
	svg := codeBlockSVG(n)
	if strings.Contains(svg, "\t") || !strings.Contains(svg, `xml:space="preserve"`) {
		t.Fatal("code must expand tabs and preserve whitespace in SVG")
	}
	var doc struct {
		Group struct {
			Text []struct {
				X    string `xml:"x,attr"`
				Body string `xml:",chardata"`
			} `xml:"text"`
		} `xml:"g"`
	}
	if err := xml.Unmarshal([]byte("<svg>"+svg+"</svg>"), &doc); err != nil {
		t.Fatal(err)
	}
	face, err := fonts.Face(fonts.WeightRegular, codeFontSize)
	if err != nil {
		t.Fatal(err)
	}
	defer face.Close()
	x := n.Rect.X + codePadX
	var reconstructed strings.Builder
	for _, span := range doc.Group.Text {
		got, err := strconv.ParseFloat(span.X, 64)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-x) > 0.051 {
			t.Errorf("span %q x=%.3f, want measured x=%.3f", span.Body, got, x)
		}
		x += float64(font.MeasureString(face, span.Body)) / 64
		reconstructed.WriteString(span.Body)
	}
	if got := reconstructed.String(); got != "    return \"é\" + 12" {
		t.Fatalf("rendered code content=%q", got)
	}
}
