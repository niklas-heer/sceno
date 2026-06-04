package export

import (
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/render"

	"github.com/jung-kurt/gofpdf"
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
