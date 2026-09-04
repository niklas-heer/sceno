package spec

import (
	"strings"
	"testing"
)

func TestSourceLocationsAndEditsPreserveRepeatedSlideIDs(t *testing.T) {
	source := []byte("diagram {\r\n  slide \"One\" {\r\n    shape box same \"First\" x=10 y=30 // first\r\n  }\r\n  slide \"Two\" {\r\n    code same source=\"print(\\\"https://example.com\\\")\\n// body\" x=20 y=40 // second\r\n  }\r\n}\r\n")
	locs, err := SourceLocations(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(locs) != 2 || locs[0].SlideIndex != 1 || locs[1].SlideIndex != 2 || locs[1].Line != 6 {
		t.Fatalf("locations: %+v", locs)
	}
	got, err := SetNodeProperties(source, 2, "same", map[string]string{"x": "42.5", "dy": "8"})
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(string(source), "x=20 y=40 // second", "x=42.5 y=40 dy=8 // second", 1)
	if string(got) != want {
		t.Fatalf("source changed beyond selected properties:\ngot %q\nwant %q", got, want)
	}
}

func TestSourceEditsRejectAmbiguity(t *testing.T) {
	for _, body := range []string{
		"  shape box a \"A\" x=1 x=2\n",
		"  shape box a \"A\"\n  shape box a \"Again\"\n",
		"  shape box a \"A\" shape box b \"B\"\n",
	} {
		if _, err := SetNodeProperties([]byte("diagram {\n"+body+"}\n"), 1, "a", map[string]string{"x": "4"}); err == nil {
			t.Fatalf("ambiguous source accepted: %q", body)
		}
	}
}
