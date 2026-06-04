// Package fonts embeds Inter (OFL) for identical typography across SVG, HTML, PNG, and PDF.
package fonts

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed data/Inter-Regular.ttf
var regularBytes []byte

//go:embed data/Inter-Medium.ttf
var mediumBytes []byte

//go:embed data/Inter-SemiBold.ttf
var semiBoldBytes []byte

//go:embed data/Inter-Bold.ttf
var boldBytes []byte

// Weight selects an Inter face.
type Weight int

const (
	WeightRegular Weight = iota
	WeightMedium
	WeightSemiBold
	WeightBold
)

var (
	loadOnce sync.Once
	loadErr  error
	regular  *truetype.Font
	medium   *truetype.Font
	semiBold *truetype.Font
	bold     *truetype.Font
)

func load() error {
	loadOnce.Do(func() {
		regular, loadErr = truetype.Parse(regularBytes)
		if loadErr != nil {
			return
		}
		medium, loadErr = truetype.Parse(mediumBytes)
		if loadErr != nil {
			return
		}
		semiBold, loadErr = truetype.Parse(semiBoldBytes)
		if loadErr != nil {
			return
		}
		bold, loadErr = truetype.Parse(boldBytes)
	})
	return loadErr
}

// Family is the CSS/SVG font-family name.
func Family() string { return "Inter" }

// Face returns a font.Face at the given pixel size (72 DPI).
func Face(w Weight, size float64) (font.Face, error) {
	if err := load(); err != nil {
		return nil, err
	}
	f := regular
	switch w {
	case WeightMedium:
		f = medium
	case WeightSemiBold:
		f = semiBold
	case WeightBold:
		f = bold
	}
	return truetype.NewFace(f, &truetype.Options{Size: size, DPI: 72}), nil
}

// WeightFromCSS maps SVG/CSS font-weight strings to a face weight.
func WeightFromCSS(w string) Weight {
	switch w {
	case "500", "medium":
		return WeightMedium
	case "600", "semibold":
		return WeightSemiBold
	case "700", "bold":
		return WeightBold
	default:
		return WeightRegular
	}
}

// RegularBytes for PDF registration.
func RegularBytes() []byte { return regularBytes }

// MediumBytes for PDF registration.
func MediumBytes() []byte { return mediumBytes }

// SemiBoldBytes for PDF registration.
func SemiBoldBytes() []byte { return semiBoldBytes }

// BoldBytes for PDF registration.
func BoldBytes() []byte { return boldBytes }

// SVGStyle returns embedded @font-face rules for SVG/HTML.
func SVGStyle() string {
	if err := load(); err != nil {
		return ""
	}
	return fmt.Sprintf(`<style type="text/css"><![CDATA[
%s
text{font-family:%s,sans-serif}
]]></style>`, cssFontFaces(), Family())
}

func cssFontFaces() string {
	faces := []struct {
		weight int
		data   []byte
	}{
		{400, regularBytes},
		{500, mediumBytes},
		{600, semiBoldBytes},
		{700, boldBytes},
	}
	var b string
	for _, f := range faces {
		b64 := base64.StdEncoding.EncodeToString(f.data)
		b += fmt.Sprintf(`@font-face{font-family:%s;font-weight:%d;font-style:normal;src:url("data:font/ttf;base64,%s") format("truetype");}`, Family(), f.weight, b64)
	}
	return b
}

// HTMLStyle returns a <style> block with @font-face for standalone HTML.
func HTMLStyle() string {
	return `<style>` + cssFontFaces() + `</style>`
}
