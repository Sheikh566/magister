package magister

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

//go:embed web/templates/*.html web/templates/pages/*.html web/templates/partials/*.html web/static/*
var webAssets embed.FS

type WebApp struct {
	registry  *Registry
	store     *Store
	templates *template.Template
	static    http.Handler
}

type homePageData struct {
	Title   string
	Courses []courseCardView
}

type courseCardView struct {
	ID          string
	Title       string
	Summary     string
	Description string
	Passed      int
	Failed      int
	Total       int
}

type coursePageData struct {
	Title              string
	Course             Course
	Runner             RunnerSettings
	Lessons            []lessonRowView
	Passed             int
	Failed             int
	Total              int
	Notice             string
	CommandHint        string
	SummaryStatusClass string
}

type lessonRowView struct {
	CourseID    string
	ID          string
	Chapter     string
	Title       string
	Summary     string
	Status      LessonStatus
	LastRunText string
}

type lessonPageData struct {
	Title           string
	Course          Course
	Lesson          Lesson
	PrevLesson      *Lesson
	NextLesson      *Lesson
	Runner          RunnerSettings
	Status          LessonStatus
	LastRunText     string
	LastOutput      string
	Notice          string
	MarkdownContent template.HTML
}

func NewWebApp(registry *Registry, store *Store) *WebApp {
	staticFS, err := fs.Sub(webAssets, "web/static")
	if err != nil {
		panic(err)
	}

	return &WebApp{
		registry:  registry,
		store:     store,
		templates: parseTemplates(),
		static:    http.FileServer(http.FS(staticFS)),
	}
}

func (a *WebApp) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", a.static))
	mux.HandleFunc("GET /", a.handleHome)
	mux.HandleFunc("GET /courses/{courseID}", a.handleCourse)
	mux.HandleFunc("POST /courses/{courseID}/config", a.handleSaveConfig)
	mux.HandleFunc("POST /courses/{courseID}/test/{lessonID}", a.handleRunLesson)
	mux.HandleFunc("GET /courses/{courseID}/lessons/{lessonID}", a.handleLesson)
	return mux
}

func (a *WebApp) handleHome(w http.ResponseWriter, r *http.Request) {
	state, err := a.store.Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cards := make([]courseCardView, 0, len(a.registry.All()))
	for _, course := range a.registry.All() {
		courseState := state.Courses[course.ID]
		passed, failed, total := Progress(course, courseState)
		cards = append(cards, courseCardView{
			ID:          course.ID,
			Title:       course.Title,
			Summary:     course.Summary,
			Description: course.Description,
			Passed:      passed,
			Failed:      failed,
			Total:       total,
		})
	}

	a.render(w, "home", homePageData{
		Title:   "Magister",
		Courses: cards,
	})
}

func (a *WebApp) handleCourse(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("courseID")
	course, ok := a.registry.ByID(courseID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	state, err := a.store.Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	courseState := state.Courses[course.ID]
	runner := courseState.Runner.RunnerConfig()
	if courseState.Runner == (RunnerSettings{}) {
		runner = DefaultRunnerConfig()
	}

	rows := make([]lessonRowView, 0, len(course.Lessons))
	for _, lesson := range course.Lessons {
		record, ok := courseState.Lessons[lesson.ID]
		status := StatusNotAttempted
		lastRunText := formatRunSummary("", "")
		if ok {
			status = record.Status
			lastRunText = formatRunSummary(record.LastRunAt, record.LastDuration)
		}
		rows = append(rows, lessonRowView{
			CourseID:    course.ID,
			ID:          lesson.ID,
			Chapter:     lesson.Chapter,
			Title:       lesson.Title,
			Summary:     lesson.Summary,
			Status:      status,
			LastRunText: lastRunText,
		})
	}

	passed, failed, total := Progress(course, courseState)
	a.render(w, "course", coursePageData{
		Title:              course.Title,
		Course:             course,
		Runner:             runnerSettingsFromConfig(runner),
		Lessons:            rows,
		Passed:             passed,
		Failed:             failed,
		Total:              total,
		Notice:             r.URL.Query().Get("notice"),
		CommandHint:        "Example: go run server.go",
		SummaryStatusClass: summaryStatusClass(passed, failed),
	})
}

func (a *WebApp) handleLesson(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("courseID")
	lessonID := r.PathValue("lessonID")

	course, ok := a.registry.ByID(courseID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	lesson, ok := course.LessonByID(lessonID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	state, err := a.store.Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	courseState := state.Courses[courseID]
	record, ok := courseState.Lessons[lessonID]
	status := StatusNotAttempted
	if ok {
		status = record.Status
	}
	runner := courseState.Runner.RunnerConfig()
	if courseState.Runner == (RunnerSettings{}) {
		runner = DefaultRunnerConfig()
	}

	var markdownHTML template.HTML
	mdPath := filepath.Join("content", courseID, lessonID+".md")
	if mdBytes, err := os.ReadFile(mdPath); err == nil {
		var buf bytes.Buffer
		if err := goldmark.Convert(mdBytes, &buf); err == nil {
			markdownHTML = template.HTML(buf.String())
		}
	}

	var prevLesson, nextLesson *Lesson
	for i, l := range course.Lessons {
		if l.ID == lessonID {
			if i > 0 {
				prevLesson = &course.Lessons[i-1]
			}
			if i < len(course.Lessons)-1 {
				nextLesson = &course.Lessons[i+1]
			}
			break
		}
	}

	a.render(w, "lesson", lessonPageData{
		Title:           lesson.Title,
		Course:          course,
		Lesson:          lesson,
		PrevLesson:      prevLesson,
		NextLesson:      nextLesson,
		Runner:          runnerSettingsFromConfig(runner),
		Status:          status,
		LastRunText:     formatRunSummary(record.LastRunAt, record.LastDuration),
		LastOutput:      record.LastOutput,
		Notice:          r.URL.Query().Get("notice"),
		MarkdownContent: markdownHTML,
	})
}

func (a *WebApp) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("courseID")
	if _, ok := a.registry.ByID(courseID); !ok {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := DefaultRunnerConfig()
	cfg.Command = strings.TrimSpace(r.FormValue("command"))
	cfg.Host = strings.TrimSpace(r.FormValue("host"))
	if v := strings.TrimSpace(r.FormValue("startup_timeout")); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.StartupTimeout = d
		}
	}
	if v := strings.TrimSpace(r.FormValue("request_timeout")); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RequestTimeout = d
		}
	}
	if _, err := fmt.Sscanf(strings.TrimSpace(r.FormValue("port")), "%d", &cfg.Port); err != nil || cfg.Port == 0 {
		cfg.Port = DefaultRunnerConfig().Port
	}

	if err := a.store.SaveRunner(courseID, cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%s?notice=%s", courseID, "Runner+settings+saved"), http.StatusSeeOther)
}

func (a *WebApp) handleRunLesson(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("courseID")
	lessonID := r.PathValue("lessonID")

	course, ok := a.registry.ByID(courseID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	state, err := a.store.Read()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cfg := state.Courses[courseID].Runner.RunnerConfig()
	if strings.TrimSpace(cfg.Command) == "" {
		location := fmt.Sprintf("/courses/%s?notice=%s", courseID, "Save+a+server+command+before+running+tests")
		if lessonID != "all" {
			location = fmt.Sprintf("/courses/%s/lessons/%s?notice=%s", courseID, lessonID, "Save+a+server+command+before+running+tests")
		}
		http.Redirect(w, r, location, http.StatusSeeOther)
		return
	}

	results, err := RunTarget(course, lessonID, cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, result := range results {
		if err := a.store.RecordResult(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if lessonID == "all" {
		http.Redirect(w, r, fmt.Sprintf("/courses/%s?notice=%s", courseID, "Ran+all+lesson+tests"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/courses/%s/lessons/%s?notice=%s", courseID, lessonID, "Lesson+test+finished"), http.StatusSeeOther)
}

func (a *WebApp) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func formatTime(raw string) string {
	if raw == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatRunSummary(lastRunAt, lastDuration string) string {
	if lastRunAt == "" {
		return "No attempts yet"
	}
	formatted := "Last run: " + formatTime(lastRunAt)
	if lastDuration != "" {
		formatted += " · " + lastDuration
	}
	return formatted
}

func parseTemplates() *template.Template {
	return template.Must(template.New("base").Funcs(templateFuncs()).ParseFS(
		webAssets,
		"web/templates/*.html",
		"web/templates/pages/*.html",
		"web/templates/partials/*.html",
	))
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"statusClass": func(status LessonStatus) string {
			switch status {
			case StatusPassed:
				return "passed"
			case StatusFailed:
				return "failed"
			default:
				return "pending"
			}
		},
		"summaryStatusClass": summaryStatusClass,
		"progressPercent": func(passed, total int) int {
			if total == 0 {
				return 0
			}
			return (passed * 100) / total
		},
	}
}

func summaryStatusClass(passed, failed int) string {
	if failed > 0 {
		return "failed"
	}
	if passed > 0 {
		return "passed"
	}
	return "pending"
}
