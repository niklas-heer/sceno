package icons

import (
	"testing"
)

func TestCatalogComplete(t *testing.T) {
	cat := Catalog()
	if len(cat) < 25 {
		t.Fatalf("expected rich catalog, got %d entries", len(cat))
	}
	for _, e := range cat {
		if e.ID == "" || e.Category == "" || e.Use == "" {
			t.Fatalf("incomplete entry: %+v", e)
		}
		if !Has(e.ID) {
			t.Fatalf("catalog entry %q missing SVG path", e.ID)
		}
	}
	for id := range svgPaths {
		if !Has(id) {
			t.Fatalf("path without Has: %s", id)
		}
	}
}

func TestByCategoryMatchesCatalog(t *testing.T) {
	by := ByCategory()
	var n int
	for _, entries := range by {
		n += len(entries)
	}
	if n != len(Catalog()) {
		t.Fatalf("by_category count %d != catalog %d", n, len(Catalog()))
	}
}

func TestDocTips(t *testing.T) {
	if len(DocTips()) < 3 {
		t.Fatal("expected icon authoring tips")
	}
}
