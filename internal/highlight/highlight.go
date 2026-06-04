package highlight

import (
	"strings"
	"unicode"
)

// Kind is a syntax token class.
type Kind string

const (
	Plain    Kind = "plain"
	Keyword  Kind = "keyword"
	String   Kind = "string"
	Comment  Kind = "comment"
	Number   Kind = "number"
)

// Span is one highlighted segment.
type Span struct {
	Text string
	Kind Kind
}

// Tokenize returns colored spans for a language (go, json, yaml, bash, kdl, text).
func Tokenize(lang, code string) []Span {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = "text"
	}
	switch lang {
	case "go", "golang":
		return tokenizeGo(code)
	case "json":
		return tokenizeJSON(code)
	case "yaml", "yml":
		return tokenizeYAML(code)
	case "bash", "sh", "shell", "zsh":
		return tokenizeBash(code)
	case "kdl":
		return tokenizeKDL(code)
	default:
		return []Span{{Text: code, Kind: Plain}}
	}
}

func tokenizeGo(code string) []Span {
	keywords := map[string]bool{
		"package": true, "import": true, "func": true, "return": true, "if": true, "else": true,
		"for": true, "range": true, "switch": true, "case": true, "default": true, "type": true,
		"struct": true, "interface": true, "map": true, "chan": true, "go": true, "defer": true,
		"var": true, "const": true, "nil": true, "true": true, "false": true,
	}
	return scanCode(code, keywords, "//", "/*", "*/")
}

func tokenizeJSON(code string) []Span {
	keywords := map[string]bool{"true": true, "false": true, "null": true}
	return scanCode(code, keywords, "", "", "")
}

func tokenizeYAML(code string) []Span {
	var out []Span
	lines := strings.Split(code, "\n")
	for li, line := range lines {
		if li > 0 {
			out = append(out, Span{Text: "\n", Kind: Plain})
		}
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "#") {
			out = append(out, Span{Text: line, Kind: Comment})
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			out = append(out, Span{Text: line[:idx], Kind: Keyword})
			out = append(out, Span{Text: line[idx:], Kind: Plain})
			continue
		}
		out = append(out, Span{Text: line, Kind: Plain})
	}
	return out
}

func tokenizeBash(code string) []Span {
	keywords := map[string]bool{
		"if": true, "then": true, "else": true, "fi": true, "for": true, "do": true, "done": true,
		"echo": true, "export": true, "cd": true, "exit": true,
	}
	return scanCode(code, keywords, "#", "", "")
}

func tokenizeKDL(code string) []Span {
	keywords := map[string]bool{
		"diagram": true, "slide": true, "shape": true, "edge": true, "layout": true, "theme": true,
	}
	return scanCode(code, keywords, "//", "", "")
}

func scanCode(code string, keywords map[string]bool, lineComment, blockStart, blockEnd string) []Span {
	var out []Span
	i := 0
	inBlock := false
	for i < len(code) {
		if inBlock && blockEnd != "" {
			if j := strings.Index(code[i:], blockEnd); j >= 0 {
				out = append(out, Span{Text: code[i : i+j+len(blockEnd)], Kind: Comment})
				i += j + len(blockEnd)
				inBlock = false
				continue
			}
			out = append(out, Span{Text: code[i:], Kind: Comment})
			break
		}
		if lineComment != "" && strings.HasPrefix(code[i:], lineComment) {
			j := strings.IndexByte(code[i:], '\n')
			if j < 0 {
				out = append(out, Span{Text: code[i:], Kind: Comment})
				break
			}
			out = append(out, Span{Text: code[i:i+j], Kind: Comment})
			i += j
			continue
		}
		if blockStart != "" && strings.HasPrefix(code[i:], blockStart) {
			inBlock = true
			continue
		}
		ch := code[i]
		if ch == '"' || ch == '\'' || ch == '`' {
			end, seg := readString(code, i)
			out = append(out, Span{Text: seg, Kind: String})
			i = end
			continue
		}
		if unicode.IsDigit(rune(ch)) {
			j := i + 1
			for j < len(code) && (unicode.IsDigit(rune(code[j])) || code[j] == '.' || code[j] == 'x') {
				j++
			}
			out = append(out, Span{Text: code[i:j], Kind: Number})
			i = j
			continue
		}
		if unicode.IsLetter(rune(ch)) || ch == '_' {
			j := i + 1
			for j < len(code) && (unicode.IsLetter(rune(code[j])) || unicode.IsDigit(rune(code[j])) || code[j] == '_') {
				j++
			}
			word := code[i:j]
			k := Plain
			if keywords[word] {
				k = Keyword
			}
			out = append(out, Span{Text: word, Kind: k})
			i = j
			continue
		}
		out = append(out, Span{Text: string(ch), Kind: Plain})
		i++
	}
	return mergeAdjacent(out)
}

func readString(code string, start int) (end int, seg string) {
	q := code[start]
	i := start + 1
	for i < len(code) {
		if code[i] == '\\' && i+1 < len(code) {
			i += 2
			continue
		}
		if code[i] == q {
			return i + 1, code[start : i+1]
		}
		i++
	}
	return len(code), code[start:]
}

func mergeAdjacent(spans []Span) []Span {
	if len(spans) == 0 {
		return spans
	}
	out := []Span{spans[0]}
	for i := 1; i < len(spans); i++ {
		last := &out[len(out)-1]
		if last.Kind == spans[i].Kind {
			last.Text += spans[i].Text
		} else {
			out = append(out, spans[i])
		}
	}
	return out
}
