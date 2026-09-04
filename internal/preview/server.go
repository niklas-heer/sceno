// Package preview provides a local, authenticated edit and inspection session.
package preview

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/niklas-heer/sceno/internal/export"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/repair"
)

const maxSourceSize = 4 << 20

type proposal struct {
	Revision string
	Request  repair.Request
	Result   repair.PreviewResult
}
type change struct{ Before, After string }
type App struct {
	mu          sync.Mutex
	path, token string
	state       State
	result      pipeline.Result
	proposalID  string
	proposal    proposal
	history     []change
	subscribers map[chan string]struct{}
}

func New(path string) (*App, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	a := &App{path: absolute, token: rand.Text(), subscribers: make(map[chan string]struct{})}
	source, err := a.read()
	if err != nil {
		return nil, err
	}
	a.state, a.result = buildState(source, a.path)
	return a, nil
}
func (a *App) read() ([]byte, error) {
	info, err := os.Lstat(a.path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("source must remain a regular file")
	}
	if info.Size() > maxSourceSize {
		return nil, errors.New("source exceeds 4 MiB limit")
	}
	return os.ReadFile(a.path)
}
func (a *App) notifyLocked() {
	for ch := range a.subscribers {
		select {
		case ch <- a.state.Revision:
		default:
		}
	}
}
func (a *App) refreshLocked() error {
	source, err := a.read()
	if err != nil {
		if a.state.FileError != err.Error() {
			a.state.FileError = err.Error()
			a.state.Valid, a.state.RenderReady = false, false
			a.state.Summary = "The source file is unavailable. Restore it to resume the preview."
			a.state.UndoAvailable = false
			a.proposalID = ""
			a.history = nil
			a.notifyLocked()
		}
		return err
	}
	if revision(source) != a.state.Revision || a.state.FileError != "" {
		a.state, a.result = buildState(source, a.path)
		a.history = nil
		a.proposalID = ""
		a.notifyLocked()
	}
	return nil
}
func (a *App) Watch(ctx context.Context) {
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.Lock()
			_ = a.refreshLocked()
			a.mu.Unlock()
		}
	}
}
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", a.getState)
	mux.HandleFunc("GET /api/events", a.events)
	mux.HandleFunc("POST /api/source", a.source)
	mux.HandleFunc("POST /api/repair/preview", a.previewRepair)
	mux.HandleFunc("POST /api/repair/apply", a.applyRepair)
	mux.HandleFunc("POST /api/undo", a.undo)
	mux.HandleFunc("GET /api/export", a.export)
	mux.Handle("/", WebHandler())
	protected := http.NewCrossOriginProtection().Handler(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != "127.0.0.1" && host != "localhost" && host != "::1" {
			http.Error(w, "loopback host required", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self' data:; img-src 'self' data: blob:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			token := r.Header.Get("X-Sceno-Token")
			if r.URL.Path == "/api/events" {
				token = r.URL.Query().Get("token")
			}
			if subtle.ConstantTimeCompare([]byte(token), []byte(a.token)) != 1 {
				writeError(w, 401, "unauthorized", "Open the preview URL printed in your terminal.")
				return
			}
			w.Header().Set("Cache-Control", "no-store")
		}
		protected.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}
func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxSourceSize+65536)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeError(w, 400, "invalid_request", err.Error())
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "invalid_request", "Expected one JSON object.")
		return false
	}
	return true
}
func (a *App) currentLocked(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Context().Err() != nil {
		writeError(w, http.StatusRequestTimeout, "request_canceled", "The request was canceled. No file was changed.")
		return false
	}
	if err := a.refreshLocked(); err != nil {
		writeError(w, 409, "file_unavailable", err.Error())
		return false
	}
	if expected == "" || expected != a.state.Revision {
		writeError(w, 409, "stale_revision", "The file changed. Reload the current source before applying this edit.")
		return false
	}
	return true
}
func (a *App) getState(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = a.refreshLocked()
	writeJSON(w, 200, a.state)
}
func (a *App) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", 500)
		return
	}
	ch := make(chan string, 1)
	a.mu.Lock()
	a.subscribers[ch] = struct{}{}
	rev := a.state.Revision
	a.mu.Unlock()
	defer func() { a.mu.Lock(); delete(a.subscribers, ch); a.mu.Unlock() }()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	fmt.Fprintf(w, "event: change\ndata: %s\n\n", rev)
	flusher.Flush()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case rev := <-ch:
			fmt.Fprintf(w, "event: change\ndata: %s\n\n", rev)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// writeLocked preserves file permissions, checks external edits twice, and
// atomically replaces the source. No history entry is created on failure.
func (a *App) writeLocked(ctx context.Context, source string, remember bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	before := a.state.Source
	current, err := a.read()
	if err != nil {
		return err
	}
	if revision(current) != a.state.Revision {
		return errors.New("file changed outside preview")
	}
	info, err := os.Stat(a.path)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(a.path), ".sceno-preview-*.kdl")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if err = file.Chmod(info.Mode().Perm()); err == nil {
		_, err = file.WriteString(source)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	current, err = a.read()
	if err != nil {
		return err
	}
	if revision(current) != a.state.Revision {
		return errors.New("file changed outside preview")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err = os.Rename(name, a.path); err != nil {
		return err
	}
	if remember {
		a.history = append(a.history, change{Before: before, After: source})
		if len(a.history) > 20 {
			a.history = a.history[len(a.history)-20:]
		}
	}
	a.state, a.result = buildState([]byte(source), a.path)
	a.state.UndoAvailable = len(a.history) > 0
	a.proposalID = ""
	a.notifyLocked()
	return nil
}
func (a *App) source(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Revision string `json:"revision"`
		Source   string `json:"source"`
	}
	if !decode(w, r, &req) {
		return
	}
	if len(req.Source) > maxSourceSize {
		writeError(w, 400, "source_too_large", "Source exceeds 4 MiB limit.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.currentLocked(w, r, req.Revision) {
		return
	}
	if req.Source != a.state.Source {
		if err := a.writeLocked(r.Context(), req.Source, true); err != nil {
			writeError(w, 409, "write_failed", err.Error())
			return
		}
	}
	writeJSON(w, 200, a.state)
}
func (a *App) previewRepair(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Revision string `json:"revision"`
		repair.Request
	}
	if !decode(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.currentLocked(w, r, req.Revision) {
		return
	}
	result, err := repair.Preview([]byte(a.state.Source), req.Request)
	if err != nil {
		writeError(w, 400, "repair_failed", err.Error())
		return
	}
	if r.Context().Err() != nil {
		writeError(w, http.StatusRequestTimeout, "request_canceled", "The repair preview was canceled.")
		return
	}
	a.proposalID = rand.Text()
	a.proposal = proposal{Revision: req.Revision, Request: req.Request, Result: result}
	writeJSON(w, 200, struct {
		ID string `json:"id"`
		repair.PreviewResult
	}{a.proposalID, result})
}
func (a *App) applyRepair(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Revision   string `json:"revision"`
		ProposalID string `json:"proposal_id"`
	}
	if !decode(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.currentLocked(w, r, req.Revision) {
		return
	}
	if req.ProposalID == "" || req.ProposalID != a.proposalID || a.proposal.Revision != req.Revision {
		writeError(w, 409, "stale_proposal", "Preview the repair again against the current file.")
		return
	}
	verified, err := repair.Preview([]byte(a.state.Source), a.proposal.Request)
	if err != nil || !verified.Applicable || verified.AfterSource != a.proposal.Result.AfterSource {
		writeError(w, 409, "repair_rejected", "This repair did not pass verification. No file was changed.")
		return
	}
	if err = a.writeLocked(r.Context(), verified.AfterSource, true); err != nil {
		writeError(w, 409, "write_failed", err.Error())
		return
	}
	writeJSON(w, 200, a.state)
}
func (a *App) undo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Revision string `json:"revision"`
	}
	if !decode(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.currentLocked(w, r, req.Revision) {
		return
	}
	if len(a.history) == 0 {
		writeError(w, 409, "no_undo", "There is no change to undo in this session.")
		return
	}
	last := a.history[len(a.history)-1]
	if a.state.Source != last.After {
		writeError(w, 409, "stale_undo", "The source changed after the last edit.")
		return
	}
	if err := a.writeLocked(r.Context(), last.Before, false); err != nil {
		writeError(w, 409, "write_failed", err.Error())
		return
	}
	a.history = a.history[:len(a.history)-1]
	a.state.UndoAvailable = len(a.history) > 0
	writeJSON(w, 200, a.state)
}
func (a *App) export(w http.ResponseWriter, r *http.Request) {
	format := export.Format(r.URL.Query().Get("format"))
	if !format.Valid() {
		writeError(w, 400, "invalid_format", "Choose svg, png, pdf, html, or slides.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.refreshLocked(); err != nil {
		writeError(w, 409, "file_unavailable", err.Error())
		return
	}
	if expected := r.URL.Query().Get("revision"); expected != "" && expected != a.state.Revision {
		writeError(w, 409, "stale_revision", "The file changed. Reload the preview before exporting.")
		return
	}
	if !a.state.RenderReady {
		writeError(w, 409, "not_ready", "Resolve structural findings before exporting.")
		return
	}
	index := 1
	if raw := r.URL.Query().Get("slide"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, 400, "invalid_slide", "Invalid slide number.")
			return
		}
		index = parsed
	}
	if index < 1 || index > len(a.result.Slides) {
		writeError(w, 400, "invalid_slide", "Slide number is out of range.")
		return
	}
	dir, err := os.MkdirTemp("", "sceno-export-*")
	if err != nil {
		writeError(w, 500, "export_failed", err.Error())
		return
	}
	defer os.RemoveAll(dir)
	name := strings.TrimSuffix(a.state.Filename, filepath.Ext(a.state.Filename)) + format.Extension()
	path := filepath.Join(dir, "diagram"+format.Extension())
	func() {
		renderMu.Lock()
		defer renderMu.Unlock()
		if format == export.FormatSlides || (format == export.FormatPDF && len(a.result.Slides) > 1) {
			err = export.WriteDeck(a.result.Deck, path, format, export.Options{Style: export.StylePolished})
		} else {
			err = export.Write(a.result.Slides[index-1].Diagram, path, format, export.Options{Style: export.StylePolished})
		}
	}()
	if err != nil {
		writeError(w, 500, "export_failed", err.Error())
		return
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	http.ServeFile(w, r, path)
}

// Serve opens a loopback-only session and exits when ctx is cancelled.
func Serve(ctx context.Context, input string, port int, openBrowser bool) error {
	app, err := New(input)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	url := "http://" + listener.Addr().String() + "/?token=" + app.token
	fmt.Printf("Sceno preview · %s\n%s\nPress Ctrl-C to stop.\n", app.path, url)
	server := &http.Server{Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go app.Watch(watchCtx)
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	if openBrowser {
		if err := openURL(url); err != nil {
			fmt.Fprintf(os.Stderr, "Could not open browser: %v. Open the URL above.\n", err)
		}
	}
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		_ = server.Close()
		return nil
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
func openURL(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	return command.Run()
}
