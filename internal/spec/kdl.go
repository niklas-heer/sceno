package spec

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/niklas-heer/sceno/internal/model"
)

// LoadKDL parses KDL bytes into a Spec.
func LoadKDL(data []byte) (model.Spec, error) {
	s, err := parseKDL(string(data))
	if err != nil {
		return model.Spec{}, err
	}
	defaults(&s)
	return s, nil
}

type kdlBlock struct {
	lines      []string
	isSlide    bool
	slideTitle string
}

func parseKDL(src string) (model.Spec, error) {
	var s model.Spec
	var stack []kdlBlock
	lines := splitKDLLines(src)

	applyPopped := func(b kdlBlock) error {
		if b.isSlide {
			sl := model.SlideSpec{Title: b.slideTitle}
			if err := applyKDLBlockSlide(&sl, b.lines); err != nil {
				return err
			}
			s.Slides = append(s.Slides, sl)
			return nil
		}
		return applyKDLBlock(&s, b.lines)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "{" {
			continue
		}
		if line == "}" {
			if len(stack) == 0 {
				continue
			}
			b := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if err := applyPopped(b); err != nil {
				return s, err
			}
			continue
		}
		if strings.HasSuffix(line, "{") {
			hdr := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			if strings.HasPrefix(hdr, "slide") {
				title, err := parseSlideHeader(hdr)
				if err != nil {
					return s, err
				}
				stack = append(stack, kdlBlock{isSlide: true, slideTitle: title})
				continue
			}
			if strings.HasPrefix(hdr, "diagram") {
				if err := applyKDLLine(&s, hdr); err != nil {
					return s, err
				}
			}
			stack = append(stack, kdlBlock{})
			continue
		}
		if len(stack) == 0 {
			stack = append(stack, kdlBlock{})
		}
		stack[len(stack)-1].lines = append(stack[len(stack)-1].lines, line)
	}
	for len(stack) > 0 {
		b := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if err := applyPopped(b); err != nil {
			return s, err
		}
	}
	return s, nil
}

func parseSlideHeader(hdr string) (string, error) {
	toks, err := tokenizeKDLine(hdr)
	if err != nil || len(toks) < 1 || toks[0].str != "slide" {
		return "", fmt.Errorf("invalid slide header %q", hdr)
	}
	for i := 1; i < len(toks); i++ {
		if toks[i].typ == kdString {
			return toks[i].str, nil
		}
		if toks[i].typ == kdProp && toks[i].key == "title" {
			return toks[i].str, nil
		}
	}
	return "Slide", nil
}

func applyKDLBlockSlide(sl *model.SlideSpec, lines []string) error {
	for _, line := range lines {
		var frag model.Spec
		if err := applyKDLLine(&frag, line); err != nil {
			return err
		}
		sl.Nodes = append(sl.Nodes, frag.Nodes...)
		sl.Edges = append(sl.Edges, frag.Edges...)
	}
	return nil
}

func splitKDLLines(src string) []string {
	var lines []string
	for _, raw := range strings.Split(src, "\n") {
		if i := strings.Index(raw, "//"); i >= 0 {
			raw = raw[:i]
		}
		lines = append(lines, strings.TrimSpace(raw))
	}
	return lines
}

func applyKDLBlock(s *model.Spec, lines []string) error {
	for _, line := range lines {
		if err := applyKDLLine(s, line); err != nil {
			return err
		}
	}
	return nil
}

func applyKDLLine(s *model.Spec, line string) error {
	toks, err := tokenizeKDLine(line)
	if err != nil || len(toks) == 0 {
		return nil
	}
	// diagram title="..." layout=auto { ... }  — header line on open block
	if len(toks) >= 1 && toks[0].str == "diagram" {
		i := 1
		for i < len(toks) {
			if toks[i].typ == kdProp {
				applyKDLProp(s, toks[i])
				i++
				continue
			}
			break
		}
		return nil
	}

	for i := 0; i < len(toks); {
		if toks[i].typ == kdProp {
			applyKDLProp(s, toks[i])
			i++
			continue
		}
		if toks[i].typ != kdWord {
			i++
			continue
		}
		kw := toks[i].str
		i++
		switch kw {
		case "title":
			if i < len(toks) {
				s.Title = kdlTokStr(toks[i])
				i++
			}
		case "subtitle":
			if i < len(toks) {
				s.Subtitle = kdlTokStr(toks[i])
				i++
			}
		case "layout", "style", "gap", "padding":
			i = applyKDLKeyword(s, kw, toks, i)
		case "shape", "node", "code":
			ns, ni, err := parseKDLShapeAt(toks, i)
			if kw == "code" && ns.Kind == model.ShapeBox {
				ns.Kind = model.ShapeCode
			}
			if err != nil {
				return err
			}
			s.Nodes = append(s.Nodes, ns)
			i = ni
		case "edge":
			es, ni, err := parseKDLEdgeAt(toks, i)
			if err != nil {
				return err
			}
			s.Edges = append(s.Edges, es)
			i = ni
		default:
			i++
		}
	}
	return nil
}

func applyKDLKeyword(s *model.Spec, kw string, toks []kdlTok, i int) int {
	if i >= len(toks) {
		return i
	}
	switch kw {
	case "layout":
		s.Layout = model.LayoutMode(kdlTokStr(toks[i]))
	case "style":
		s.Style = model.RenderStyle(kdlTokStr(toks[i]))
	case "gap":
		s.Gap = kdlTokNum(toks[i])
	case "padding":
		s.Padding = kdlTokNum(toks[i])
	}
	return i + 1
}

// shape box api "Label" icon=server layer=1
// shape api "Label"   (kind omitted → box)
func parseKDLShapeAt(toks []kdlTok, start int) (model.NodeSpec, int, error) {
	ns := model.NodeSpec{Kind: model.ShapeBox}
	props := map[string]kdlTok{}
	i := start

	// Optional kind as first word
	rawKind := ""
	if i < len(toks) && toks[i].typ == kdWord && !isKDLStmt(toks[i].str) && isShapeName(toks[i].str) {
		rawKind = toks[i].str
		ns.Kind = model.NormalizeShape(model.ShapeKind(toks[i].str))
		i++
	}
	// id
	if i < len(toks) && toks[i].typ == kdWord && !isKDLStmt(toks[i].str) {
		ns.ID = toks[i].str
		i++
	}
	if i < len(toks) && toks[i].typ == kdString {
		ns.Label = toks[i].str
		i++
	}
	for ; i < len(toks); i++ {
		if toks[i].typ == kdArrow {
			continue
		}
		if toks[i].typ == kdProp {
			props[toks[i].key] = toks[i]
		} else if toks[i].typ == kdWord && isKDLStmt(toks[i].str) {
			break
		}
	}
	applyNodeProps(&ns, props)
	applyShapeVariantDefaults(&ns, rawKind)
	if ns.ID == "" {
		return ns, i, fmt.Errorf("shape missing id")
	}
	if ns.Label == "" {
		ns.Label = ns.ID
	}
	return ns, i, nil
}

func applyNodeProps(ns *model.NodeSpec, props map[string]kdlTok) {
	for k, v := range props {
		switch k {
		case "id":
			ns.ID = v.str
		case "label":
			ns.Label = unescapeLabel(v.str)
		case "subtitle":
			ns.Subtitle = unescapeLabel(v.str)
		case "kind", "shape":
			ns.Kind = model.NormalizeShape(model.ShapeKind(v.str))
		case "icon":
			ns.Icon = v.str
		case "iconPos", "icon-pos", "iconpos":
			ns.IconPos = model.ParseIconPosition(v.str)
		case "fill":
			ns.Fill = v.str
		case "stroke":
			ns.Stroke = v.str
		case "accent":
			ns.Accent = v.str
		case "parent":
			ns.Parent = v.str
		case "layer", "col", "column":
			ns.Layer = int(v.num)
		case "row":
			ns.Row = int(v.num)
		case "at":
			// at=layer,row — explicit grid slot (layer 0 is valid)
			ns.AtSet = true
			parts := strings.Split(v.str, ",")
			if len(parts) >= 1 {
				if n, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
					ns.Layer = n
				}
			}
			if len(parts) >= 2 {
				if n, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					ns.Row = n
				}
			}
		case "w":
			ns.W = v.num
		case "h":
			ns.H = v.num
		case "fontSize":
			ns.FontSize = v.num
		case "x":
			x := v.num
			ns.X = &x
		case "y":
			y := v.num
			ns.Y = &y
		case "lang", "language":
			ns.CodeLang = v.str
		case "source", "body":
			ns.Code = unescapeLabel(v.str)
		}
	}
}

func applyShapeVariantDefaults(ns *model.NodeSpec, rawKind string) {
	if ns.Accent != "" || rawKind == "" {
		return
	}
	switch strings.ToLower(rawKind) {
	case "info":
		ns.Accent = "#3b82f6"
	case "warning", "warn":
		ns.Accent = "#f59e0b"
	case "tip", "hint":
		ns.Accent = "#10b981"
	}
}

func unescapeLabel(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}

func isShapeName(s string) bool {
	_, ok := shapeNames[strings.ToLower(s)]
	return ok
}

var shapeNames = func() map[string]bool {
	m := map[string]bool{}
	for _, n := range model.AllShapes() {
		m[n] = true
	}
	return m
}()

// edge api -> queue  |  edge e from=api to=queue
func parseKDLEdgeAt(toks []kdlTok, start int) (model.EdgeSpec, int, error) {
	es := model.EdgeSpec{}
	props := map[string]kdlTok{}
	i := start
	for ; i < len(toks); i++ {
		switch toks[i].typ {
		case kdProp:
			props[toks[i].key] = toks[i]
		case kdString:
			if es.Label == "" {
				es.Label = unescapeLabel(toks[i].str)
			}
		case kdArrow:
			// from -> to already captured
			continue
		case kdWord:
			if isKDLStmt(toks[i].str) {
				goto done
			}
			if es.From == "" {
				es.From = toks[i].str
			} else if es.To == "" {
				es.To = toks[i].str
			}
		}
	}
done:
	if v, ok := props["from"]; ok {
		es.From = v.str
	}
	if v, ok := props["to"]; ok {
		es.To = v.str
	}
	if v, ok := props["label"]; ok {
		es.Label = unescapeLabel(v.str)
	}
	if v, ok := props["fromSide"]; ok {
		es.FromSide = model.Side(v.str)
	}
	if v, ok := props["toSide"]; ok {
		es.ToSide = model.Side(v.str)
	}
	if v, ok := props["color"]; ok {
		es.Color = v.str
	}
	if v, ok := props["dashed"]; ok {
		es.Dashed = v.bool
	}
	if es.From == "" || es.To == "" {
		return es, i, fmt.Errorf("edge needs from and to (use: edge api -> queue)")
	}
	return es, i, nil
}

func isKDLStmt(w string) bool {
	switch w {
	case "diagram", "title", "subtitle", "layout", "style", "gap", "padding", "slide", "shape", "node", "edge", "code", "theme":
		return true
	}
	return false
}

func applyKDLProp(s *model.Spec, p kdlTok) {
	switch p.key {
	case "title":
		s.Title = unescapeLabel(p.str)
	case "subtitle":
		s.Subtitle = unescapeLabel(p.str)
	case "layout":
		s.Layout = model.LayoutMode(p.str)
	case "style":
		s.Style = model.RenderStyle(p.str)
	case "gap":
		s.Gap = p.num
	case "padding":
		s.Padding = p.num
	case "slide":
		s.SlideAspect = p.str
	case "theme":
		s.Theme.Mode = p.str
	case "background", "bg":
		if strings.EqualFold(p.str, "transparent") {
			s.Theme.Transparent = true
		} else {
			if s.Theme.Vars == nil {
				s.Theme.Vars = map[string]string{}
			}
			s.Theme.Vars["background"] = p.str
		}
	case "transparent":
		s.Theme.Transparent = p.bool || strings.EqualFold(p.str, "true")
	case "foreground", "fg":
		setThemeVar(s, "foreground", p.str)
	case "card":
		setThemeVar(s, "card", p.str)
	case "border":
		setThemeVar(s, "border", p.str)
	case "muted":
		setThemeVar(s, "muted", p.str)
	case "accent":
		setThemeVar(s, "accent", p.str)
	default:
		if strings.HasPrefix(p.key, "var.") {
			setThemeVar(s, strings.TrimPrefix(p.key, "var."), p.str)
		}
	}
}

func setThemeVar(s *model.Spec, key, val string) {
	if val == "" {
		return
	}
	if s.Theme.Vars == nil {
		s.Theme.Vars = map[string]string{}
	}
	s.Theme.Vars[key] = val
}

func kdlTokStr(t kdlTok) string {
	if t.typ == kdString || t.typ == kdWord {
		return unescapeLabel(t.str)
	}
	return ""
}

func kdlTokNum(t kdlTok) float64 {
	if t.typ == kdNumber {
		return t.num
	}
	return 0
}

type kdKind int

const (
	kdWord kdKind = iota
	kdString
	kdNumber
	kdBool
	kdProp
	kdArrow
)

type kdlTok struct {
	typ  kdKind
	str  string
	num  float64
	bool bool
	key  string
}

func tokenizeKDLine(line string) ([]kdlTok, error) {
	var out []kdlTok
	i := 0
	runes := []rune(line)
	skip := func() {
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			i++
		}
	}
	for {
		skip()
		if i >= len(runes) {
			break
		}
		if i+1 < len(runes) && runes[i] == '-' && runes[i+1] == '>' {
			out = append(out, kdlTok{typ: kdArrow, str: "->"})
			i += 2
			continue
		}
		if runes[i] == '"' {
			i++
			var b strings.Builder
			for i < len(runes) && runes[i] != '"' {
				if runes[i] == '\\' && i+1 < len(runes) {
					i++
					switch runes[i] {
					case 'n':
						b.WriteByte('\n')
					case 't':
						b.WriteByte('\t')
					case '"':
						b.WriteByte('"')
					case '\\':
						b.WriteByte('\\')
					default:
						b.WriteRune(runes[i])
					}
					i++
					continue
				}
				b.WriteRune(runes[i])
				i++
			}
			if i < len(runes) {
				i++
			}
			out = append(out, kdlTok{typ: kdString, str: b.String()})
			continue
		}
		j := i
		for j < len(runes) && !unicode.IsSpace(runes[j]) && !(runes[j] == '-' && j+1 < len(runes) && runes[j+1] == '>') {
			j++
		}
		word := string(runes[i:j])
		i = j
		if strings.Contains(word, "=") {
			parts := strings.SplitN(word, "=", 2)
			key := parts[0]
			valPart := parts[1]
			if strings.HasPrefix(valPart, `"`) {
				if strings.HasSuffix(valPart, `"`) && len(valPart) >= 2 {
					out = append(out, kdlTok{typ: kdProp, key: key, str: strings.Trim(valPart, `"`)})
					continue
				}
				var b strings.Builder
				b.WriteString(strings.TrimPrefix(valPart, `"`))
				skip()
				if b.Len() > 0 {
					b.WriteByte(' ')
				}
				for i < len(runes) {
					if runes[i] == '"' {
						i++
						break
					}
					b.WriteRune(runes[i])
					i++
				}
				out = append(out, kdlTok{typ: kdProp, key: key, str: b.String()})
				continue
			}
			tok, err := parseKDLValue(valPart)
			if err != nil {
				return nil, err
			}
			tok.key = key
			tok.typ = kdProp
			out = append(out, tok)
			continue
		}
		tok, err := parseKDLValue(word)
		if err != nil {
			return nil, err
		}
		out = append(out, tok)
	}
	return out, nil
}

func parseKDLValue(s string) (kdlTok, error) {
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return kdlTok{typ: kdString, str: strings.Trim(s, `"`)}, nil
	}
	if s == "true" {
		return kdlTok{typ: kdBool, bool: true}, nil
	}
	if s == "false" {
		return kdlTok{typ: kdBool, bool: false}, nil
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return kdlTok{typ: kdNumber, num: n, str: s}, nil
	}
	return kdlTok{typ: kdWord, str: s}, nil
}
