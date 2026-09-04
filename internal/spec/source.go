package spec

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SourceLocation identifies one source statement. SlideIndex and Line are 1-based.
type SourceLocation struct {
	SlideIndex int    `json:"slide_index"`
	Target     string `json:"target"`
	Line       int    `json:"line"`
	Source     string `json:"source"`
}

// SourceLocations maps shape/code statements without normalizing their source.
// Ambiguous multi-statement lines and nested unsupported blocks are rejected.
func SourceLocations(source []byte) ([]SourceLocation, error) {
	if _, err := LoadKDL(source); err != nil {
		return nil, err
	}
	var locations []SourceLocation
	slide, nextSlide := 1, 0
	var blocks []string
	for i, raw := range strings.Split(string(source), "\n") {
		line := splitKDLLines(raw)[0]
		if line == "" {
			continue
		}
		if line == "}" {
			if len(blocks) == 0 {
				return nil, fmt.Errorf("line %d: unmatched block close", i+1)
			}
			blocks = blocks[:len(blocks)-1]
			slide = 1
			continue
		}
		if strings.HasSuffix(line, "{") {
			head := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			toks, err := tokenizeKDLine(head)
			if err != nil || len(toks) == 0 {
				return nil, fmt.Errorf("line %d: invalid block header", i+1)
			}
			switch toks[0].str {
			case "diagram":
				if len(blocks) != 0 {
					return nil, fmt.Errorf("line %d: nested diagram cannot be mapped", i+1)
				}
			case "slide":
				if len(blocks) != 1 || blocks[0] != "diagram" {
					return nil, fmt.Errorf("line %d: ambiguous slide scope", i+1)
				}
				nextSlide++
				slide = nextSlide
			default:
				return nil, fmt.Errorf("line %d: unsupported source block", i+1)
			}
			blocks = append(blocks, toks[0].str)
			continue
		}
		toks, err := tokenizeKDLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		if len(toks) == 0 {
			continue
		}
		if toks[0].str != "shape" && toks[0].str != "node" && toks[0].str != "code" {
			continue
		}
		n, end, err := parseKDLShapeAt(toks, 1)
		if err != nil || end != len(toks) {
			return nil, fmt.Errorf("line %d: ambiguous shape statement", i+1)
		}
		quoted := false
		for j := 0; j < len(line); j++ {
			if quoted && line[j] == '\\' {
				j++
				continue
			}
			if line[j] == '"' {
				quoted = !quoted
			} else if !quoted && strings.ContainsRune(";{}", rune(line[j])) {
				return nil, fmt.Errorf("line %d: multiple statements cannot be edited", i+1)
			}
		}
		locations = append(locations, SourceLocation{SlideIndex: slide, Target: n.ID, Line: i + 1, Source: strings.TrimSuffix(raw, "\r")})
	}
	if len(blocks) != 0 {
		return nil, fmt.Errorf("unclosed source block")
	}
	return locations, nil
}

type sourceToken struct {
	text       string
	start, end int
}

// rawSourceTokens retains byte ranges, treating escaped quoted strings as one token.
func rawSourceTokens(line string) []sourceToken {
	var out []sourceToken
	for i := 0; i < len(line); {
		r, n := utf8.DecodeRuneInString(line[i:])
		if unicode.IsSpace(r) {
			i += n
			continue
		}
		if strings.HasPrefix(line[i:], "//") {
			break
		}
		start := i
		quoted := false
		for i < len(line) {
			if quoted && line[i] == '\\' {
				i++
				if i < len(line) {
					_, n = utf8.DecodeRuneInString(line[i:])
					i += n
				}
				continue
			}
			if line[i] == '"' {
				quoted = !quoted
				i++
				continue
			}
			r, n = utf8.DecodeRuneInString(line[i:])
			if !quoted && (unicode.IsSpace(r) || strings.HasPrefix(line[i:], "//")) {
				break
			}
			i += n
		}
		out = append(out, sourceToken{text: line[start:i], start: start, end: i})
	}
	return out
}

// SetNodeProperties changes a uniquely mapped statement, preserving other bytes.
// Callers must validate property names/values and re-evaluate the returned source.
func SetNodeProperties(source []byte, slideIndex int, target string, properties map[string]string) ([]byte, error) {
	locations, err := SourceLocations(source)
	if err != nil {
		return nil, err
	}
	var found []SourceLocation
	for _, loc := range locations {
		if loc.SlideIndex == slideIndex && loc.Target == target {
			found = append(found, loc)
		}
	}
	if len(found) != 1 {
		return nil, fmt.Errorf("expected one source statement for slide %d node %q; found %d", slideIndex, target, len(found))
	}
	lines := strings.SplitAfter(string(source), "\n")
	index := found[0].Line - 1
	line := lines[index]
	tokens := rawSourceTokens(strings.TrimRight(line, "\r\n"))
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	type edit struct {
		start, end int
		value      string
	}
	var edits []edit
	var additions []string
	for _, key := range keys {
		value := properties[key]
		if key == "" || strings.ContainsAny(key, "= \t\r\n\"/") || value == "" || strings.ContainsAny(value, " \t\r\n\"/;") {
			return nil, fmt.Errorf("invalid source property %q", key)
		}
		var matches []sourceToken
		for _, tok := range tokens {
			if strings.HasPrefix(tok.text, key+"=") {
				matches = append(matches, tok)
			}
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("ambiguous duplicate property %q", key)
		}
		if len(matches) == 1 {
			tok := matches[0]
			edits = append(edits, edit{tok.start + len(key) + 1, tok.end, value})
		} else {
			additions = append(additions, key+"="+value)
		}
	}
	if len(additions) > 0 {
		pos := tokens[len(tokens)-1].end
		edits = append(edits, edit{pos, pos, " " + strings.Join(additions, " ")})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	for _, e := range edits {
		line = line[:e.start] + e.value + line[e.end:]
	}
	lines[index] = line
	return []byte(strings.Join(lines, "")), nil
}
