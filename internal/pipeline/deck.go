package pipeline

import "github.com/niklas-heer/sceno/internal/model"

// BuildDeck lays out a multi-slide spec or a single-slide diagram.
func BuildDeck(s model.Spec, opt Options) (model.Deck, []model.Collision, error) {
	if len(s.Slides) == 0 {
		d, colls, err := BuildFromSpec(s, opt)
		if err != nil {
			return model.Deck{}, colls, err
		}
		d.SlideAspect = s.SlideAspect
		return model.Deck{
			Title:       s.Title,
			Subtitle:    s.Subtitle,
			SlideAspect: s.SlideAspect,
			Theme:       s.Theme,
			Slides:      []model.Diagram{d},
		}, colls, nil
	}

	deck := model.Deck{
		Title:       s.Title,
		Subtitle:    s.Subtitle,
		SlideAspect: s.SlideAspect,
		Theme:       s.Theme,
	}
	var allColls []model.Collision
	for _, sl := range s.Slides {
		sub := model.Spec{
			Title:       pickTitle(sl.Title, s.Title),
			Subtitle:    s.Subtitle,
			Layout:      s.Layout,
			Style:       s.Style,
			Gap:         s.Gap,
			Padding:     s.Padding,
			SlideAspect: s.SlideAspect,
			Theme:       s.Theme,
			Nodes:       sl.Nodes,
			Edges:       sl.Edges,
		}
		d, colls, err := BuildFromSpec(sub, opt)
		if err != nil {
			return deck, allColls, err
		}
		d.Title = pickTitle(sl.Title, d.Title)
		d.SlideAspect = s.SlideAspect
		deck.Slides = append(deck.Slides, d)
		allColls = append(allColls, colls...)
	}
	return deck, allColls, nil
}

func pickTitle(slideTitle, fallback string) string {
	if slideTitle != "" {
		return slideTitle
	}
	return fallback
}
