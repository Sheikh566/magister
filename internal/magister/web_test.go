package magister_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpcourse "magister/internal/course"
	"magister/internal/magister"
)

func TestWebRoutesRender(t *testing.T) {
	t.Parallel()

	registry := magister.NewRegistry(httpcourse.HTTPServerCourse())
	store := magister.NewStore(t.TempDir())
	if err := store.SaveRunner("http-server", magister.RunnerConfig{
		Command:        "go run server.go",
		Host:           "127.0.0.1",
		Port:           8080,
		StartupTimeout: 5 * time.Second,
		RequestTimeout: 2 * time.Second,
	}); err != nil {
		t.Fatalf("save runner: %v", err)
	}
	if err := store.RecordResult(magister.LessonResult{
		CourseID:   "http-server",
		LessonID:   "http-01",
		Status:     magister.StatusPassed,
		Output:     "[http-01] Return a Minimal HTTP Response\nPASSED",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Duration:   250 * time.Millisecond,
	}); err != nil {
		t.Fatalf("record result: %v", err)
	}

	app := magister.NewWebApp(registry, store)
	handler := app.Handler()

	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/", want: "Magister"},
		{path: "/courses/http-server", want: "Build Your Own HTTP Server"},
		{path: "/courses/http-server/lessons/http-01", want: "Return a Minimal HTTP Response"},
		{path: "/static/styles.css", want: ":root {"},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status for %s: %d", tc.path, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, tc.want) {
			t.Fatalf("expected %q in %s body", tc.want, tc.path)
		}
	}
}
