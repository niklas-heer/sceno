package advise

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/inspect"
)

func TestAdviseHowItWorks(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "how-it-works.kdl")
	report, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.ValidationOK {
		t.Fatalf("expected valid spec")
	}
	if report.VisualScore <= 0 {
		t.Fatalf("expected positive visual score")
	}
	if len(report.Engine.RulesRun) < 5 {
		t.Fatalf("expected engine rules")
	}
}

func TestAdvisePreservesEverySlidesMeasuredGeometry(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "slides-demo.kdl")
	report, err := Run(path, Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	description, err := inspect.Run(path, inspect.Options{FixCollisions: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Slides) != len(description.Slides) || len(report.Slides) != 3 {
		t.Fatalf("missing slide geometry: %d", len(report.Slides))
	}
	for i, slide := range report.Slides {
		if slide.Index != i+1 || !reflect.DeepEqual(slide.Engine, description.Slides[i].Engine) {
			t.Fatalf("advise and describe disagree on slide %d", i+1)
		}
	}
	if report.EngineSlideIndex < 1 || !reflect.DeepEqual(report.Engine.SceneStack, report.Slides[report.EngineSlideIndex-1].Engine.SceneStack) {
		t.Fatal("representative engine geometry has no matching slide")
	}
	prompt, err := buildAIPrompt(report)
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(prompt, "{\n")
	var payload struct {
		Slides []SlideReport `json:"slides"`
	}
	if start < 0 {
		t.Fatal("missing JSON in AI prompt")
	}
	if err := json.Unmarshal([]byte(prompt[start:]), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Slides) != len(report.Slides) {
		t.Fatal("AI review dropped measured geometry")
	}
	want, _ := json.Marshal(report.Slides)
	got, _ := json.Marshal(payload.Slides)
	if string(got) != string(want) {
		t.Fatal("AI review changed per-slide measurements")
	}
}
