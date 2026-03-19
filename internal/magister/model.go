package magister

import (
	"context"
	"time"
)

type LessonStatus string

const (
	StatusNotAttempted LessonStatus = "not attempted"
	StatusPassed       LessonStatus = "passed"
	StatusFailed       LessonStatus = "failed"
)

type RunnerConfig struct {
	Command        string
	Host           string
	Port           int
	StartupTimeout time.Duration
	RequestTimeout time.Duration
}

func DefaultRunnerConfig() RunnerConfig {
	return RunnerConfig{
		Host:           "127.0.0.1",
		Port:           8080,
		StartupTimeout: 5 * time.Second,
		RequestTimeout: 2 * time.Second,
	}
}

func NormalizeRunnerConfig(cfg RunnerConfig) RunnerConfig {
	defaults := DefaultRunnerConfig()
	if cfg.Host == "" {
		cfg.Host = defaults.Host
	}
	if cfg.Port == 0 {
		cfg.Port = defaults.Port
	}
	if cfg.StartupTimeout <= 0 {
		cfg.StartupTimeout = defaults.StartupTimeout
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = defaults.RequestTimeout
	}
	return cfg
}

type Lesson struct {
	ID         string
	Chapter    string
	Title      string
	Summary    string
	Objectives []string
	TestFocus  []string
	WireText   string
	Run        func(context.Context, RunnerConfig) error
}

type Course struct {
	ID          string
	Title       string
	Summary     string
	Description string
	Lessons     []Lesson
}

func (c Course) LessonByID(id string) (Lesson, bool) {
	for _, lesson := range c.Lessons {
		if lesson.ID == id {
			return lesson, true
		}
	}
	return Lesson{}, false
}

type Registry struct {
	order   []string
	courses map[string]Course
}

func NewRegistry(courses ...Course) *Registry {
	r := &Registry{
		order:   make([]string, 0, len(courses)),
		courses: make(map[string]Course, len(courses)),
	}
	for _, course := range courses {
		r.order = append(r.order, course.ID)
		r.courses[course.ID] = course
	}
	return r
}

func (r *Registry) All() []Course {
	out := make([]Course, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.courses[id])
	}
	return out
}

func (r *Registry) ByID(id string) (Course, bool) {
	course, ok := r.courses[id]
	return course, ok
}

type LessonResult struct {
	CourseID   string
	LessonID   string
	Status     LessonStatus
	Output     string
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
}
