package preview

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklas-heer/sceno/internal/repair"
	"github.com/niklas-heer/sceno/internal/starter"
)

func testApp(t *testing.T, source string) *App {
	t.Helper()
	path := filepath.Join(t.TempDir(), "diagram.kdl")
	if err := os.WriteFile(path, []byte(source), 0640); err != nil {
		t.Fatal(err)
	}
	app, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	return app
}
func call(t *testing.T, a *App, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	r := httptest.NewRequest(method, "http://127.0.0.1"+path, bytes.NewReader(data))
	r.Header.Set("X-Sceno-Token", a.token)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	return w
}
func stateResponse(t *testing.T, w *httptest.ResponseRecorder) State {
	t.Helper()
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var out State
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func TestEditUndoConflictAndInvalidSource(t *testing.T) {
	source, _ := starter.Source("")
	a := testApp(t, source)
	before := stateResponse(t, call(t, a, "GET", "/api/state", nil))
	if !before.Valid || !before.RenderReady || len(before.Slides) != 1 || before.Slides[0].SVG == "" {
		t.Fatalf("unexpected initial state: valid=%v ready=%v slides=%d", before.Valid, before.RenderReady, len(before.Slides))
	}
	changed := stateResponse(t, call(t, a, "POST", "/api/source", map[string]string{"revision": before.Revision, "source": "diagram { broken"}))
	if changed.Valid || changed.RenderReady || !changed.UndoAvailable {
		t.Fatalf("invalid source state: %+v", changed)
	}
	for _, slide := range changed.Slides {
		if slide.SVG != "" {
			t.Fatal("invalid source rendered")
		}
	}
	if w := call(t, a, "GET", "/api/export?format=svg", nil); w.Code != 409 {
		t.Fatal("invalid export allowed")
	}
	restored := stateResponse(t, call(t, a, "POST", "/api/undo", map[string]string{"revision": changed.Revision}))
	if restored.Source != source || restored.UndoAvailable {
		t.Fatal("undo failed")
	}
	info, _ := os.Stat(a.path)
	if info.Mode().Perm() != 0640 {
		t.Fatal("file mode changed")
	}
	external := source + "\n// External editor\n"
	if err := os.WriteFile(a.path, []byte(external), 0640); err != nil {
		t.Fatal(err)
	}
	if w := call(t, a, "POST", "/api/source", map[string]string{"revision": before.Revision, "source": source}); w.Code != 409 {
		t.Fatalf("stale edit accepted: %s", w.Body.String())
	}
	actual, _ := os.ReadFile(a.path)
	if string(actual) != external {
		t.Fatal("external change overwritten")
	}
}
func TestVerifiedRepairApplyUndo(t *testing.T) {
	source := "diagram layout=free gap=20 {\n shape box a \"A\" x=0 y=150 w=100 h=80\n shape box b \"B\" x=60 y=150 w=100 h=80\n}\n"
	a := testApp(t, source)
	state := stateResponse(t, call(t, a, "GET", "/api/state", nil))
	var request map[string]any
	for _, issue := range state.Issues {
		if string(issue.Code) == "collision" && len(issue.Repairs) > 0 {
			option := issue.Repairs[0]
			request = map[string]any{"revision": state.Revision, "slide_index": issue.SlideIndex, "issue_id": issue.ID, "target": option.Target, "properties": option.Properties}
			break
		}
	}
	if request == nil {
		t.Fatal("no collision repair")
	}
	w := call(t, a, "POST", "/api/repair/preview", request)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var proposed struct {
		ID string `json:"id"`
		repair.PreviewResult
	}
	if err := json.Unmarshal(w.Body.Bytes(), &proposed); err != nil {
		t.Fatal(err)
	}
	if !proposed.Applicable {
		t.Fatalf("repair rejected: %v", proposed.Reasons)
	}
	disk, _ := os.ReadFile(a.path)
	if string(disk) != source {
		t.Fatal("preview wrote file")
	}
	applied := stateResponse(t, call(t, a, "POST", "/api/repair/apply", map[string]string{"revision": state.Revision, "proposal_id": proposed.ID}))
	if !applied.Valid || applied.Source != proposed.AfterSource || !applied.UndoAvailable {
		t.Fatal("repair not applied")
	}
	if w = call(t, a, "POST", "/api/repair/apply", map[string]string{"revision": applied.Revision, "proposal_id": proposed.ID}); w.Code != 409 {
		t.Fatal("proposal reused")
	}
	undone := stateResponse(t, call(t, a, "POST", "/api/undo", map[string]string{"revision": applied.Revision}))
	if undone.Source != source {
		t.Fatal("repair undo failed")
	}
}
func TestLoopbackAuthenticationAndOrigin(t *testing.T) {
	source, _ := starter.Source("")
	a := testApp(t, source)
	for _, tc := range []struct {
		host, token, origin string
		want                int
	}{{"127.0.0.1", "", "", 401}, {"evil.example", a.token, "", 403}, {"127.0.0.1", a.token, "https://evil.example", 403}} {
		r := httptest.NewRequest("POST", "http://"+tc.host+"/api/source", strings.NewReader(`{}`))
		r.Header.Set("X-Sceno-Token", tc.token)
		if tc.origin != "" {
			r.Header.Set("Origin", tc.origin)
		}
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("host %s origin %s: got %d", tc.host, tc.origin, w.Code)
		}
	}
}
func TestExportAndSourceLocations(t *testing.T) {
	source, _ := starter.Source("")
	a := testApp(t, source)
	state := stateResponse(t, call(t, a, "GET", "/api/state", nil))
	if state.Slides[0].Nodes[0].Line <= 0 {
		t.Fatal("missing source location")
	}
	w := call(t, a, "GET", "/api/export?format=svg", nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "<svg") || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("export failed: %d %s", w.Code, w.Body.String())
	}
	if w := call(t, a, "GET", "/api/export?format=svg&slide=999", nil); w.Code != 400 {
		t.Fatal("invalid slide accepted")
	}
}
func TestMissingFileAndSymlinkReplacement(t *testing.T) {
	source, _ := starter.Source("")
	a := testApp(t, source)
	rev := a.state.Revision
	if err := os.Remove(a.path); err != nil {
		t.Fatal(err)
	}
	state := stateResponse(t, call(t, a, "GET", "/api/state", nil))
	if state.FileError == "" || state.RenderReady {
		t.Fatal("missing file not reported")
	}
	other := filepath.Join(t.TempDir(), "other.kdl")
	os.WriteFile(other, []byte(source), 0600)
	if err := os.Symlink(other, a.path); err != nil {
		t.Fatal(err)
	}
	w := call(t, a, "POST", "/api/source", map[string]string{"revision": rev, "source": "oops"})
	if w.Code != 409 {
		t.Fatal("symlink replacement accepted")
	}
	contents, _ := os.ReadFile(other)
	if string(contents) != source {
		t.Fatal("symlink target changed")
	}
}
func TestEventsStream(t *testing.T) {
	source, _ := starter.Source("")
	a := testApp(t, source)
	server := httptest.NewServer(a.Handler())
	defer server.Close()
	response, err := http.Get(server.URL + "/api/events?token=" + a.token)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	buf := make([]byte, len("event: change\n"))
	if _, err := io.ReadFull(response.Body, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "event: change\n" {
		t.Fatalf("unexpected event: %s", buf)
	}
}
