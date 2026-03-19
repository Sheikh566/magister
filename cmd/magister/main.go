package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	httpcourse "magister/internal/course"
	"magister/internal/magister"
)

func main() {
	registry := magister.NewRegistry(httpcourse.HTTPServerCourse())
	store := magister.NewStore(".")
	os.Exit(run(registry, store, os.Args[1:], os.Stdout, os.Stderr))
}

func run(registry *magister.Registry, store *magister.Store, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "courses":
		return runCourses(registry, stdout)
	case "lessons":
		return runLessons(registry, args[1:], stdout, stderr)
	case "show":
		return runShow(registry, args[1:], stdout, stderr)
	case "test":
		return runTest(registry, store, args[1:], stdout, stderr)
	case "serve":
		return runServe(registry, store, args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runCourses(registry *magister.Registry, stdout io.Writer) int {
	for _, course := range registry.All() {
		fmt.Fprintf(stdout, "%s  %s\n", course.ID, course.Title)
		fmt.Fprintf(stdout, "  %s\n\n", course.Summary)
	}
	return 0
}

func runLessons(registry *magister.Registry, args []string, stdout io.Writer, stderr io.Writer) int {
	course, ok := resolveCourse(registry, args)
	if !ok {
		fmt.Fprintln(stderr, "usage: magister lessons <course-id>")
		return 2
	}

	for _, lesson := range course.Lessons {
		fmt.Fprintf(stdout, "%-8s %-28s %s\n", lesson.ID, lesson.Chapter, lesson.Title)
	}
	return 0
}

func runShow(registry *magister.Registry, args []string, stdout io.Writer, stderr io.Writer) int {
	course, lesson, ok := resolveLesson(registry, args)
	if !ok {
		fmt.Fprintln(stderr, "usage: magister show <course-id> <lesson-id>")
		return 2
	}

	fmt.Fprintf(stdout, "%s  %s\n", lesson.ID, lesson.Title)
	fmt.Fprintf(stdout, "Course: %s\n", course.Title)
	fmt.Fprintf(stdout, "Chapter: %s\n\n", lesson.Chapter)
	fmt.Fprintf(stdout, "%s\n\n", lesson.Summary)
	fmt.Fprintln(stdout, "Observable contract:")
	for _, item := range lesson.Objectives {
		fmt.Fprintf(stdout, "- %s\n", item)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Black-box test focus:")
	for _, item := range lesson.TestFocus {
		fmt.Fprintf(stdout, "- %s\n", item)
	}
	if lesson.WireText != "" {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Raw wire text:")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, lesson.WireText)
	}
	return 0
}

func runTest(registry *magister.Registry, store *magister.Store, args []string, stdout io.Writer, stderr io.Writer) int {
	courseID, target, flagArgs, ok := splitTestArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "usage: magister test <course-id> <lesson-id|all> --cmd \"go run ./your-server\"")
		return 2
	}

	course, found := registry.ByID(courseID)
	if !found {
		fmt.Fprintf(stderr, "unknown course %q\n", courseID)
		return 2
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(stderr)
	cmd := fs.String("cmd", "", "command used to start your server")
	host := fs.String("host", "", "host the tester should connect to")
	port := fs.Int("port", 0, "port your server should listen on")
	startupTimeout := fs.Duration("startup-timeout", 0, "max time to wait for the server to start")
	requestTimeout := fs.Duration("request-timeout", 0, "per-request timeout")
	if err := fs.Parse(flagArgs); err != nil {
		return 2
	}

	state, err := store.Read()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	cfg := state.Courses[courseID].Runner.RunnerConfig()
	if (state.Courses[courseID].Runner == magister.RunnerSettings{}) {
		cfg = magister.DefaultRunnerConfig()
	}

	if strings.TrimSpace(*cmd) != "" {
		cfg.Command = strings.TrimSpace(*cmd)
	}
	if strings.TrimSpace(*host) != "" {
		cfg.Host = strings.TrimSpace(*host)
	}
	if *port != 0 {
		cfg.Port = *port
	}
	if *startupTimeout != 0 {
		cfg.StartupTimeout = *startupTimeout
	}
	if *requestTimeout != 0 {
		cfg.RequestTimeout = *requestTimeout
	}
	cfg = magister.NormalizeRunnerConfig(cfg)

	if strings.TrimSpace(cfg.Command) == "" {
		fmt.Fprintln(stderr, "--cmd is required the first time you run a course")
		return 2
	}

	if err := store.SaveRunner(courseID, cfg); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	results, err := magister.RunTarget(course, target, cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	exitCode := 0
	for _, result := range results {
		if err := store.RecordResult(result); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, result.Output)
		fmt.Fprintln(stdout)
		if result.Status == magister.StatusFailed {
			exitCode = 1
		}
	}
	return exitCode
}

func runServe(registry *magister.Registry, store *magister.Store, args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	addr := fs.String("addr", "127.0.0.1:8787", "local address for the Magister web UI")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	app := magister.NewWebApp(registry, store)
	server := &http.Server{
		Addr:              *addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 2 * time.Second,
	}

	fmt.Fprintf(stdout, "Magister running at http://%s\n", *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func resolveCourse(registry *magister.Registry, args []string) (magister.Course, bool) {
	if len(args) == 1 {
		return registry.ByID(args[0])
	}
	if len(args) == 0 {
		all := registry.All()
		if len(all) == 1 {
			return all[0], true
		}
	}
	return magister.Course{}, false
}

func resolveLesson(registry *magister.Registry, args []string) (magister.Course, magister.Lesson, bool) {
	if len(args) == 2 {
		course, ok := registry.ByID(args[0])
		if !ok {
			return magister.Course{}, magister.Lesson{}, false
		}
		lesson, ok := course.LessonByID(args[1])
		return course, lesson, ok
	}
	if len(args) == 1 {
		all := registry.All()
		if len(all) != 1 {
			return magister.Course{}, magister.Lesson{}, false
		}
		lesson, ok := all[0].LessonByID(args[0])
		return all[0], lesson, ok
	}
	return magister.Course{}, magister.Lesson{}, false
}

func splitTestArgs(args []string) (string, string, []string, bool) {
	if len(args) >= 2 && !strings.HasPrefix(args[0], "-") && !strings.HasPrefix(args[1], "-") {
		return args[0], args[1], args[2:], true
	}
	return "", "", nil, false
}

func printUsage(w io.Writer) {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	fmt.Fprintln(bw, "Magister")
	fmt.Fprintln(bw)
	fmt.Fprintln(bw, "Commands:")
	fmt.Fprintln(bw, "  courses                              list available courses")
	fmt.Fprintln(bw, "  lessons <course-id>                  list lessons for one course")
	fmt.Fprintln(bw, "  show <course-id> <lesson-id>         print one lesson")
	fmt.Fprintln(bw, "  test <course-id> <lesson-id|all>     run black-box tests")
	fmt.Fprintln(bw, "  serve [--addr 127.0.0.1:8787]        start the local web UI")
	fmt.Fprintln(bw)
	fmt.Fprintln(bw, "Examples:")
	fmt.Fprintln(bw, "  go run ./cmd/magister courses")
	fmt.Fprintln(bw, "  go run ./cmd/magister lessons http-server")
	fmt.Fprintln(bw, "  go run ./cmd/magister show http-server http-01")
	fmt.Fprintln(bw, "  go run ./cmd/magister test http-server tcp-01 --cmd \"go run server.go\"")
	fmt.Fprintln(bw, "  go run ./cmd/magister serve")
}
