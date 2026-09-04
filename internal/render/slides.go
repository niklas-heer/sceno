package render

import (
	"fmt"
	"strings"

	"github.com/niklas-heer/sceno/internal/fonts"
	"github.com/niklas-heer/sceno/internal/model"
)

// SlidesHTML renders a self-contained presentation (keyboard + scroll-snap).
func SlidesHTML(deck model.Deck) string {
	useDeckPalette(deck)
	aspect := deck.SlideAspect
	if aspect == "" {
		aspect = Aspect16x9
	}
	sw, _ := SlideSize(aspect)
	title := deck.Title
	if title == "" {
		title = "Diagram"
	}
	bodyClass := ""
	if paint.Mode == "dark" {
		bodyClass = ` class="dark"`
	}
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">")
	fmt.Fprintf(&b, "<title>%s</title>", xmlEsc(title))
	b.WriteString(fonts.HTMLStyle())
	b.WriteString("<style>")
	b.WriteString(paint.CSSVars())
	b.WriteString(paint.SlideCSS())
	b.WriteString(`
*{box-sizing:border-box}html,body{margin:0;height:100%;font-family:Inter,ui-sans-serif,system-ui,sans-serif;background:var(--background);color:var(--foreground)}
.deck{height:100vh;overflow-y:auto;scroll-snap-type:y mandatory;scroll-behavior:smooth}
.slide{scroll-snap-align:start;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:32px 24px 48px;gap:20px}
.slide-inner svg{display:block;width:100%;height:100%}
.deck-hdr{position:fixed;top:16px;left:24px;z-index:10;font-size:0.75rem;color:var(--muted-foreground);background:var(--card);border:1px solid var(--border);padding:6px 12px;border-radius:9999px;box-shadow:0 1px 2px var(--ring)}
.deck-hdr kbd{font-family:inherit;background:var(--muted);padding:2px 6px;border-radius:4px;margin:0 2px}
.nav{position:fixed;bottom:24px;right:24px;z-index:10;display:flex;gap:8px}
.nav button{font:inherit;border:1px solid var(--border);background:var(--card);color:var(--foreground);padding:8px 14px;border-radius:calc(var(--radius));cursor:pointer;box-shadow:0 1px 2px var(--ring)}
.nav button:hover{background:var(--muted)}
.slide-inner{width:100%;max-width:min(96vw,` + fmt.Sprintf("%.0fpx", sw) + `);aspect-ratio:` + aspectRatioCSS(aspect) + `;border:1px solid var(--border);border-radius:calc(var(--radius) * 2);box-shadow:0 1px 2px var(--ring),0 16px 48px rgb(9 9 11 / 6%);overflow:hidden;position:relative}
@media print{.deck-hdr,.nav{display:none}.slide{page-break-after:always;min-height:100vh}}
</style></head><body` + bodyClass + `>`)
	b.WriteString(`<div class="deck-hdr"><kbd>←</kbd><kbd>→</kbd> navigate · ` + fmt.Sprintf("%d", len(deck.Slides)) + ` slides</div>`)
	b.WriteString(`<div class="nav"><button type="button" id="prev">Prev</button><button type="button" id="next">Next</button></div>`)
	b.WriteString(`<div class="deck" id="deck">`)
	for i, d := range deck.Slides {
		if d.SlideAspect == "" {
			d.SlideAspect = aspect
		}
		useDiagramPalette(d)
		b.WriteString(`<section class="slide" data-index="` + fmt.Sprint(i) + `">`)
		// Keep code, routes, and chrome in the same computed scene as every
		// other export; separate HTML flow would invalidate describe geometry.
		b.WriteString(`<div class="slide-inner">`)
		b.WriteString(PolishedSVGSlide(d))
		b.WriteString(`</div>`)
		b.WriteString(`</section>`)
	}
	b.WriteString(`</div><script>
(function(){
  const deck=document.getElementById('deck');
  const slides=[...deck.querySelectorAll('.slide')];
  let i=0;
  function go(n){i=Math.max(0,Math.min(slides.length-1,n));slides[i].scrollIntoView({behavior:'smooth'});}
  document.getElementById('prev').onclick=()=>go(i-1);
  document.getElementById('next').onclick=()=>go(i+1);
  document.onkeydown=(e)=>{
    if(e.key==='ArrowRight'||e.key===' '||e.key==='PageDown'){e.preventDefault();go(i+1);}
    if(e.key==='ArrowLeft'||e.key==='PageUp'){e.preventDefault();go(i-1);}
  };
  deck.onscroll=()=>{
    const y=deck.scrollTop;
    let best=0,bestD=1e9;
    slides.forEach((s,j)=>{const d=Math.abs(s.offsetTop-y);if(d<bestD){bestD=d;best=j;}});
    i=best;
  };
})();
</script></body></html>`)
	return b.String()
}

func aspectRatioCSS(aspect string) string {
	switch aspect {
	case Aspect4x3:
		return "4 / 3"
	default:
		return "16 / 9"
	}
}
