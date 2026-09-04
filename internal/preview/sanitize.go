package preview

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/niklas-heer/sceno/internal/fonts"
)

const svgNamespace = "http://www.w3.org/2000/svg"
const xlinkNamespace = "http://www.w3.org/1999/xlink"

// Only passive SVG geometry is inserted into the application's DOM. Animation,
// HTML integration points, resource images, links, and executable elements are
// deliberately absent. Unknown elements are removed with their entire subtree.
var previewSVGElements = wordSet(`svg g defs symbol use path rect circle ellipse line polyline polygon text tspan textPath title desc marker filter feDropShadow feGaussianBlur feOffset feBlend feColorMatrix feComposite feFlood feMerge feMergeNode linearGradient radialGradient stop clipPath mask pattern`)
var previewSVGAttributes = wordSet(`id class viewBox preserveAspectRatio x y x1 y1 x2 y2 cx cy r rx ry width height d points transform opacity fill fill-opacity fill-rule stroke stroke-width stroke-opacity stroke-dasharray stroke-dashoffset stroke-linecap stroke-linejoin stroke-miterlimit clip-rule font-family font-size font-weight font-style text-anchor dominant-baseline alignment-baseline baseline-shift dx dy rotate textLength lengthAdjust startOffset method spacing markerWidth markerHeight markerUnits refX refY orient filter filterUnits primitiveUnits stdDeviation result in in2 mode operator k1 k2 k3 k4 type values flood-color flood-opacity color-interpolation-filters gradientUnits gradientTransform spreadMethod offset stop-color stop-opacity fx fy fr clip-path clipPathUnits mask maskUnits maskContentUnits patternUnits patternContentUnits patternTransform marker-start marker-mid marker-end vector-effect role aria-label aria-hidden`)
var previewSVGPaint = wordSet(`fill stroke color flood-color stop-color`)
var previewSVGReferences = wordSet(`filter clip-path mask marker-start marker-mid marker-end`)
var fragmentID = regexp.MustCompile(`^#[A-Za-z_][A-Za-z0-9_.:-]*$`)
var plainColor = regexp.MustCompile(`^(?:[A-Za-z]+|#[0-9A-Fa-f]{3}|#[0-9A-Fa-f]{4}|#[0-9A-Fa-f]{6}|#[0-9A-Fa-f]{8}|(?:rgb|rgba|hsl|hsla)\([0-9.,%+\- \t]+\))$`)

var previewFontsOnce sync.Once
var previewFontsStyle, previewFontsCSS string

func wordSet(words string) map[string]bool {
	out := map[string]bool{}
	for _, word := range strings.Fields(words) {
		out[word] = true
	}
	return out
}

// safeSVG fails closed on malformed XML. It reconstructs an inert SVG rather
// than attempting to remove dangerous strings from markup. The sole stylesheet
// accepted is Sceno's exact embedded-font stylesheet, supplied by trusted code.
func safeSVG(source string) string {
	decoder := xml.NewDecoder(strings.NewReader(source))
	var out bytes.Buffer
	encoder := xml.NewEncoder(&out)
	seenRoot := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ""
		}
		switch token := token.(type) {
		case xml.StartElement:
			if seenRoot || token.Name.Local != "svg" || !svgName(token.Name) {
				return ""
			}
			seenRoot = true
			if err := sanitizeSVGElement(decoder, encoder, &out, token, true, 0); err != nil {
				return ""
			}
		case xml.CharData:
			if strings.TrimSpace(string(token)) != "" {
				return ""
			}
		case xml.ProcInst:
			if seenRoot || token.Target != "xml" {
				return ""
			}
		case xml.Comment:
		default:
			return ""
		}
	}
	if !seenRoot || encoder.Flush() != nil {
		return ""
	}
	return out.String()
}

func svgName(name xml.Name) bool { return name.Space == "" || name.Space == svgNamespace }

func sanitizeSVGElement(decoder *xml.Decoder, encoder *xml.Encoder, out *bytes.Buffer, start xml.StartElement, root bool, depth int) error {
	if depth > 256 {
		return fmt.Errorf("SVG nesting exceeds preview limit")
	}
	if !svgName(start.Name) {
		return decoder.Skip()
	}
	seenAttrs := map[xml.Name]bool{}
	for _, a := range start.Attr {
		if seenAttrs[a.Name] {
			return fmt.Errorf("duplicate XML attribute")
		}
		seenAttrs[a.Name] = true
	}
	if start.Name.Local == "style" {
		var body struct {
			Text string `xml:",chardata"`
		}
		if err := decoder.DecodeElement(&body, &start); err != nil {
			return err
		}
		previewFontsOnce.Do(func() {
			previewFontsStyle = fonts.SVGStyle()
			var trusted struct {
				Text string `xml:",chardata"`
			}
			if xml.Unmarshal([]byte(previewFontsStyle), &trusted) == nil {
				previewFontsCSS = trusted.Text
			}
		})
		if previewFontsCSS != "" && body.Text == previewFontsCSS {
			if err := encoder.Flush(); err != nil {
				return err
			}
			out.WriteString(previewFontsStyle)
		}
		return nil
	}
	if !previewSVGElements[start.Name.Local] {
		return decoder.Skip()
	}
	clean := xml.StartElement{Name: xml.Name{Local: start.Name.Local}}
	if root {
		clean.Attr = append(clean.Attr, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: svgNamespace})
	}
	names := map[string]bool{}
	for _, attr := range start.Attr {
		name, value, ok := safeSVGAttribute(attr)
		if !ok {
			continue
		}
		if names[name] {
			return fmt.Errorf("ambiguous SVG attribute")
		}
		names[name] = true
		clean.Attr = append(clean.Attr, xml.Attr{Name: xml.Name{Local: name}, Value: value})
	}
	if err := encoder.EncodeToken(clean); err != nil {
		return err
	}
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			if err := sanitizeSVGElement(decoder, encoder, out, token, false, depth+1); err != nil {
				return err
			}
		case xml.EndElement:
			return encoder.EncodeToken(clean.End())
		case xml.CharData:
			if err := encoder.EncodeToken(token); err != nil {
				return err
			}
		case xml.Comment:
		default:
			return fmt.Errorf("unsupported XML directive in SVG")
		}
	}
}

func safeSVGAttribute(attr xml.Attr) (string, string, bool) {
	name, value := attr.Name.Local, strings.TrimSpace(attr.Value)
	if attr.Name.Space == "http://www.w3.org/XML/1998/namespace" && name == "space" {
		return "xml:space", value, value == "preserve" || value == "default"
	}
	if strings.HasPrefix(strings.ToLower(name), "on") {
		return "", "", false
	}
	if name == "href" && (attr.Name.Space == "" || attr.Name.Space == xlinkNamespace) {
		return "href", value, fragmentID.MatchString(value)
	}
	if attr.Name.Space != "" || !previewSVGAttributes[name] {
		return "", "", false
	}
	if previewSVGPaint[name] {
		return name, value, plainColor.MatchString(value) || safeSVGReference(value)
	}
	if previewSVGReferences[name] {
		return name, value, value == "none" || safeSVGReference(value)
	}
	// No other supported attribute needs CSS escapes, URL syntax, or a scheme.
	// Reject them instead of relying on browser-specific CSS error recovery.
	if strings.ContainsAny(value, "\\<>\x00") || strings.Contains(strings.ToLower(value), "url(") || strings.Contains(strings.ToLower(value), "javascript:") || strings.Contains(strings.ToLower(value), "data:") || strings.Contains(value, "://") {
		return "", "", false
	}
	return name, attr.Value, true
}

func safeSVGReference(value string) bool {
	if !strings.HasPrefix(value, "url(") || !strings.HasSuffix(value, ")") {
		return false
	}
	inner := strings.TrimSpace(value[4 : len(value)-1])
	if len(inner) >= 2 && ((inner[0] == '\'' && inner[len(inner)-1] == '\'') || (inner[0] == '"' && inner[len(inner)-1] == '"')) {
		inner = inner[1 : len(inner)-1]
	}
	return fragmentID.MatchString(inner)
}
