package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

func TestRoutes_APICatalogAdvertisesPublicProductAPI(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{logger: logger, cfg: &config.Config{}}
	mux := api.routes(
		&workout.WorkoutHandler{},
		&exercise.ExerciseHandler{},
		&featureaccess.Handler{},
		health.NewHandler(logger, nil),
		aichat.NewHandler(logger, nil),
		nil, nil, nil, nil,
	)

	req := httptest.NewRequest(http.MethodGet, "https://fittrack.example/.well-known/api-catalog", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	const wantContentType = `application/linkset+json; profile="https://www.rfc-editor.org/info/rfc9727"`
	if got := rr.Header().Get("Content-Type"); got != wantContentType {
		t.Fatalf("Content-Type = %q, want %q", got, wantContentType)
	}

	var document map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode catalog JSON: %v", err)
	}
	if len(document) != 1 {
		t.Fatalf("catalog top-level keys = %v, want only linkset", document)
	}
	linksets, ok := document["linkset"].([]any)
	if !ok || len(linksets) != 1 {
		t.Fatalf("linkset = %#v, want one API entry", document["linkset"])
	}
	entry, ok := linksets[0].(map[string]any)
	if !ok || len(entry) != 4 {
		t.Fatalf("API entry = %#v, want anchor and three relations", linksets[0])
	}

	assertSameOriginCatalogURL(t, req.URL, entry["anchor"], "/api")
	assertCatalogRelation(t, req.URL, entry, "service-desc", "/swagger/doc.json", "application/json")
	assertCatalogRelation(t, req.URL, entry, "service-doc", "/swagger/", "text/html")
	assertCatalogRelation(t, req.URL, entry, "status", "/health", "application/json")
}

func TestHomepage_DiscoveryHeadersAndMarkdownNegotiation(t *testing.T) {
	withStaticTestDirectory(t)
	writeStaticTestFile(t, "index.html", "<!doctype html><title>FitTrack SPA</title>")
	writeStaticTestFile(t, filepath.Join("assets", "app.js"), "console.log('spa')")

	handler := (&api{}).handleStaticFiles()
	wantLinks := []string{
		`</.well-known/api-catalog>; rel="api-catalog"; type="application/linkset+json"`,
		`</swagger/doc.json>; rel="service-desc"; type="application/json"`,
		`</swagger/>; rel="service-doc"; type="text/html"`,
	}

	tests := []struct {
		name            string
		path            string
		accept          string
		wantContentType string
		wantBody        string
		wantLinks       bool
		wantVary        string
	}{
		{
			name:            "explicit Markdown",
			path:            "/",
			accept:          "text/markdown",
			wantContentType: "text/markdown; charset=utf-8",
			wantBody:        "# FitTrack",
			wantLinks:       true,
			wantVary:        "Accept",
		},
		{
			name:            "Markdown preferred by quality",
			path:            "/",
			accept:          "text/html;q=0.5, text/markdown;q=0.9",
			wantContentType: "text/markdown; charset=utf-8",
			wantBody:        "API description",
			wantLinks:       true,
			wantVary:        "Accept",
		},
		{
			name:            "HTML preferred by quality",
			path:            "/",
			accept:          "text/markdown;q=0.2, text/html;q=0.8",
			wantContentType: "text/html; charset=utf-8",
			wantBody:        "FitTrack SPA",
			wantLinks:       true,
			wantVary:        "Accept",
		},
		{
			name:            "Markdown explicitly unacceptable",
			path:            "/",
			accept:          "text/markdown;q=0, text/html",
			wantContentType: "text/html; charset=utf-8",
			wantBody:        "FitTrack SPA",
			wantLinks:       true,
			wantVary:        "Accept",
		},
		{
			name:            "normal browser HTML",
			path:            "/",
			accept:          "text/html,application/xhtml+xml,*/*;q=0.8",
			wantContentType: "text/html; charset=utf-8",
			wantBody:        "FitTrack SPA",
			wantLinks:       true,
			wantVary:        "Accept",
		},
		{
			name:            "SPA fallback unchanged",
			path:            "/workouts",
			accept:          "text/html",
			wantContentType: "text/html; charset=utf-8",
			wantBody:        "FitTrack SPA",
		},
		{
			name:            "static asset unchanged",
			path:            "/assets/app.js",
			accept:          "*/*",
			wantContentType: "text/javascript; charset=utf-8",
			wantBody:        "console.log('spa')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Accept", tt.accept)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", rr.Code, http.StatusOK, rr.Body.String())
			}
			if got := rr.Header().Get("Content-Type"); got != tt.wantContentType {
				t.Fatalf("Content-Type = %q, want %q", got, tt.wantContentType)
			}
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Fatalf("body %q does not contain %q", rr.Body.String(), tt.wantBody)
			}
			if got := rr.Header().Get("Vary"); got != tt.wantVary {
				t.Fatalf("Vary = %q, want %q", got, tt.wantVary)
			}
			if got := rr.Header().Values("Link"); !reflect.DeepEqual(got, conditionalLinks(tt.wantLinks, wantLinks)) {
				t.Fatalf("Link headers = %#v, want %#v", got, conditionalLinks(tt.wantLinks, wantLinks))
			}
		})
	}
}

func assertCatalogRelation(t *testing.T, base *url.URL, entry map[string]any, relation, wantPath, wantType string) {
	t.Helper()
	targets, ok := entry[relation].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("%s = %#v, want one target", relation, entry[relation])
	}
	target, ok := targets[0].(map[string]any)
	if !ok || len(target) != 2 {
		t.Fatalf("%s target = %#v, want href and type", relation, targets[0])
	}
	assertSameOriginCatalogURL(t, base, target["href"], wantPath)
	if got := target["type"]; got != wantType {
		t.Fatalf("%s type = %#v, want %q", relation, got, wantType)
	}
}

func assertSameOriginCatalogURL(t *testing.T, base *url.URL, raw any, wantPath string) {
	t.Helper()
	href, ok := raw.(string)
	if !ok {
		t.Fatalf("href = %#v, want string", raw)
	}
	parsed, err := url.Parse(href)
	if err != nil {
		t.Fatalf("parse href %q: %v", href, err)
	}
	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != base.Scheme || resolved.Host != base.Host || resolved.Path != wantPath {
		t.Fatalf("resolved href = %q, want same-origin path %q", resolved.String(), wantPath)
	}
}

func conditionalLinks(include bool, links []string) []string {
	if !include {
		return nil
	}
	return links
}

func withStaticTestDirectory(t *testing.T) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("switch to temporary static file directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
