package highlight

import "testing"

func TestTokenizeGoKeyword(t *testing.T) {
	spans := Tokenize("go", "package main")
	found := false
	for _, s := range spans {
		if s.Kind == Keyword && s.Text == "package" {
			found = true
		}
	}
	if !found {
		t.Fatalf("spans: %+v", spans)
	}
}
