package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/starter"
)

func TestUnavailableFileDisablesGeometryAndRecovers(t *testing.T) {
	source, _ := starter.Source("")
	app := testApp(t, source)
	if err := os.Remove(app.path); err != nil {
		t.Fatal(err)
	}
	state := stateResponse(t, call(t, app, "GET", "/api/state", nil))
	if state.Valid || state.RenderReady || state.UndoAvailable || state.FileError == "" {
		t.Fatalf("missing file still presented as current: %+v", state)
	}
	if err := os.WriteFile(app.path, []byte(source), 0640); err != nil {
		t.Fatal(err)
	}
	state = stateResponse(t, call(t, app, "GET", "/api/state", nil))
	if !state.Valid || state.FileError != "" {
		t.Fatalf("restored file did not recover: %+v", state)
	}
}

func TestBuildStateRecoversInvalidLayout(t *testing.T) {
	bad := []byte("diagram {\n shape box a \"A\" at=0,-1\n}\n")
	state, _ := buildState(bad, "negative.kdl")
	if state.Valid || state.RenderReady || len(state.Issues) == 0 || len(state.Slides) != 0 {
		t.Fatalf("invalid layout not reported safely: %+v", state)
	}
	if state.Source != string(bad) || state.Revision != revision(bad) {
		t.Fatal("recovery lost editable source")
	}
	source, _ := starter.Source("")
	good, _ := buildState([]byte(source), "recovered.kdl")
	if !good.Valid || len(good.Slides) == 0 || good.Slides[0].SVG == "" {
		t.Fatal("layout failure poisoned later preview")
	}
}

func TestBuildStateRejectsUndisplayableSVG(t *testing.T) {
	source := []byte("diagram {\n shape box a \"A\" fill=\"red\\\"\" at=0,0\n}\n")
	state, _ := buildState(source, "bad-paint.kdl")
	if state.Valid || state.RenderReady || !strings.Contains(state.Summary, "safely") {
		t.Fatalf("undisplayable SVG marked ready: %+v", state)
	}
	for _, slide := range state.Slides {
		if slide.SVG != "" {
			t.Fatal("partial unsafe render retained")
		}
	}
}

func TestCanceledEditWaitingForLockDoesNotWrite(t *testing.T) {
	source, _ := starter.Source("")
	app := testApp(t, source)
	data, err := json.Marshal(map[string]string{"revision": app.state.Revision, "source": source + "\n// should not be saved\n"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := httptest.NewRequest("POST", "http://127.0.0.1/api/source", bytes.NewReader(data)).WithContext(ctx)
	r.Header.Set("X-Sceno-Token", app.token)
	w := httptest.NewRecorder()
	app.mu.Lock()
	done := make(chan struct{})
	go func() { app.Handler().ServeHTTP(w, r); close(done) }()
	cancel()
	app.mu.Unlock()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("canceled request stuck")
	}
	if w.Code != http.StatusRequestTimeout {
		t.Fatalf("canceled edit status = %d: %s", w.Code, w.Body.String())
	}
	actual, err := os.ReadFile(app.path)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != source || len(app.history) != 0 {
		t.Fatal("canceled edit changed file/history")
	}
}

func TestWatchSurvivesInvalidLayoutAndStops(t *testing.T) {
	source, _ := starter.Source("")
	app := testApp(t, source)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { app.Watch(ctx); close(done) }()
	if err := os.WriteFile(app.path, []byte("diagram {\n shape box a \"A\" at=0,-1\n}\n"), 0640); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	found := false
	for !found {
		select {
		case <-deadline:
			t.Fatal("watcher did not publish layout failure")
		case <-ticker.C:
			app.mu.Lock()
			found = !app.state.Valid
			app.mu.Unlock()
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop")
	}
}

func TestExportRejectsStaleVisibleRevision(t *testing.T) {
	source, _ := starter.Source("")
	app := testApp(t, source)
	original := app.state.Revision
	if err := os.WriteFile(app.path, []byte(source+"\n// changed externally\n"), 0640); err != nil {
		t.Fatal(err)
	}
	w := call(t, app, "GET", "/api/export?format=svg&revision="+original, nil)
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "stale_revision") {
		t.Fatalf("unseen source was exported: %d %s", w.Code, w.Body.String())
	}
	w = call(t, app, "GET", "/api/export?format=svg&revision="+app.state.Revision, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("current visible source export rejected: %d %s", w.Code, w.Body.String())
	}
}

func TestHugeGridCoordinatesRemainEditable(t *testing.T) {
	source := []byte("diagram {\n shape box a \"A\" at=2147483647,0\n}\n")
	state, _ := buildState(source, "huge.kdl")
	if state.Valid || state.RenderReady || len(state.Issues) == 0 || !strings.Contains(state.Issues[0].Message, "grid coordinates") {
		t.Fatalf("large grid coordinate was not rejected before layout: %+v", state)
	}
}

func TestHugeFiniteGeometryReturnsUsefulJSON(t *testing.T) {
	source := []byte("diagram gap=1e308 {\n shape box a \"A\" x=1e308 y=0 w=1e308 h=1e308\n}\n")
	state, result := buildState(source, "huge-finite.kdl")
	if state.Valid || state.RenderReady || len(state.Issues) == 0 || len(result.Slides) != 0 {
		t.Fatalf("unsafe finite geometry reached engine: %+v", state)
	}
	data, err := json.Marshal(state)
	if err != nil || !json.Valid(data) || state.Source != string(source) {
		t.Fatalf("failed to preserve editable JSON state: %v", err)
	}
}

func TestComputedNonfiniteGeometryFallsBackToEditableState(t *testing.T) {
	state := State{Source: "editable source", Revision: "revision", Valid: true, RenderReady: true, Slides: []Slide{{Canvas: model.Rect{W: math.NaN()}}}}
	if ensureSerializableState(&state) {
		t.Fatal("computed NaN was accepted")
	}
	if state.Valid || state.RenderReady || len(state.Slides) != 0 || len(state.Issues) == 0 || state.Source != "editable source" {
		t.Fatalf("fallback discarded editable source or kept invalid geometry: %+v", state)
	}
	if _, err := json.Marshal(state); err != nil {
		t.Fatalf("fallback still cannot be encoded: %v", err)
	}
}
