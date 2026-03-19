package magister

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	path string
	mu   sync.Mutex
}

type State struct {
	Courses map[string]CourseState `json:"courses"`
}

type CourseState struct {
	Runner  RunnerSettings          `json:"runner"`
	Lessons map[string]LessonRecord `json:"lessons"`
}

type RunnerSettings struct {
	Command        string `json:"command"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	StartupTimeout string `json:"startup_timeout"`
	RequestTimeout string `json:"request_timeout"`
}

type LessonRecord struct {
	Status       LessonStatus `json:"status"`
	LastRunAt    string       `json:"last_run_at"`
	LastDuration string       `json:"last_duration"`
	LastOutput   string       `json:"last_output"`
}

func NewStore(root string) *Store {
	return &Store{
		path: filepath.Join(root, ".magister", "state.json"),
	}
}

func (s *Store) Read() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *Store) SaveRunner(courseID string, cfg RunnerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.readLocked()
	if err != nil {
		return err
	}
	courseState := ensureCourseState(state, courseID)
	courseState.Runner = runnerSettingsFromConfig(cfg)
	state.Courses[courseID] = courseState
	return s.writeLocked(state)
}

func (s *Store) RecordResult(result LessonResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.readLocked()
	if err != nil {
		return err
	}
	courseState := ensureCourseState(state, result.CourseID)
	if courseState.Lessons == nil {
		courseState.Lessons = map[string]LessonRecord{}
	}
	courseState.Lessons[result.LessonID] = LessonRecord{
		Status:       result.Status,
		LastRunAt:    result.FinishedAt.Format(time.RFC3339),
		LastDuration: result.Duration.String(),
		LastOutput:   result.Output,
	}
	state.Courses[result.CourseID] = courseState
	return s.writeLocked(state)
}

func (s *Store) readLocked() (State, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{Courses: map[string]CourseState{}}, nil
		}
		return State{}, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("parse state file %s: %w", s.path, err)
	}
	if state.Courses == nil {
		state.Courses = map[string]CourseState{}
	}
	return state, nil
}

func (s *Store) writeLocked(state State) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func ensureCourseState(state State, courseID string) CourseState {
	courseState, ok := state.Courses[courseID]
	if !ok {
		courseState = CourseState{
			Runner:  runnerSettingsFromConfig(DefaultRunnerConfig()),
			Lessons: map[string]LessonRecord{},
		}
	}
	if courseState.Lessons == nil {
		courseState.Lessons = map[string]LessonRecord{}
	}
	return courseState
}

func runnerSettingsFromConfig(cfg RunnerConfig) RunnerSettings {
	cfg = NormalizeRunnerConfig(cfg)
	return RunnerSettings{
		Command:        cfg.Command,
		Host:           cfg.Host,
		Port:           cfg.Port,
		StartupTimeout: cfg.StartupTimeout.String(),
		RequestTimeout: cfg.RequestTimeout.String(),
	}
}

func (s RunnerSettings) RunnerConfig() RunnerConfig {
	cfg := DefaultRunnerConfig()
	cfg.Command = s.Command
	if s.Host != "" {
		cfg.Host = s.Host
	}
	if s.Port != 0 {
		cfg.Port = s.Port
	}
	if d, err := time.ParseDuration(s.StartupTimeout); err == nil && d > 0 {
		cfg.StartupTimeout = d
	}
	if d, err := time.ParseDuration(s.RequestTimeout); err == nil && d > 0 {
		cfg.RequestTimeout = d
	}
	return NormalizeRunnerConfig(cfg)
}

func Progress(course Course, state CourseState) (passed, failed, total int) {
	total = len(course.Lessons)
	for _, lesson := range course.Lessons {
		record, ok := state.Lessons[lesson.ID]
		if !ok {
			continue
		}
		switch record.Status {
		case StatusPassed:
			passed++
		case StatusFailed:
			failed++
		}
	}
	return passed, failed, total
}
