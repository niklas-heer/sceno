package export

import (
	"fmt"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"

	gofpdf "github.com/go-pdf/fpdf"
)

// WritePDF renders a polished PDF with embedded Inter fonts.
func WritePDF(d model.Diagram, path string, opt Options) error {
	vp := render.ViewportFrom(d)
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: vp.Width, Ht: vp.Height},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()
	render.DrawPolishedPDF(pdf, d, vp.MinX, vp.MinY)
	return pdf.OutputFileAndClose(path)
}

// WritePDFDeck renders every slide as a page while preserving each slide viewport.
func WritePDFDeck(deck model.Deck, path string, opt Options) error {
	if len(deck.Slides) == 0 {
		return fmt.Errorf("empty deck")
	}
	first := render.ViewportFrom(deck.Slides[0])
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: first.Width, Ht: first.Height},
	})
	pdf.SetMargins(0, 0, 0)
	for _, d := range deck.Slides {
		vp := render.ViewportFrom(d)
		pdf.AddPageFormat("", gofpdf.SizeType{Wd: vp.Width, Ht: vp.Height})
		render.DrawPolishedPDF(pdf, d, vp.MinX, vp.MinY)
	}
	return pdf.OutputFileAndClose(path)
}
