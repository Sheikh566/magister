package magister

import (
	"context"
	"fmt"
	"time"
)

func RunLesson(course Course, lesson Lesson, cfg RunnerConfig) LessonResult {
	cfg = NormalizeRunnerConfig(cfg)
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.StartupTimeout+15*time.Second)
	defer cancel()

	err := lesson.Run(ctx, cfg)
	finishedAt := time.Now()
	result := LessonResult{
		CourseID:   course.ID,
		LessonID:   lesson.ID,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Duration:   finishedAt.Sub(startedAt),
	}
	if err != nil {
		result.Status = StatusFailed
		result.Output = fmt.Sprintf("[%s] %s\nFAILED: %s", lesson.ID, lesson.Title, err)
		return result
	}

	result.Status = StatusPassed
	result.Output = fmt.Sprintf("[%s] %s\nPASSED", lesson.ID, lesson.Title)
	return result
}

func RunTarget(course Course, target string, cfg RunnerConfig) ([]LessonResult, error) {
	if target == "all" {
		results := make([]LessonResult, 0, len(course.Lessons))
		for _, lesson := range course.Lessons {
			results = append(results, RunLesson(course, lesson, cfg))
		}
		return results, nil
	}

	lesson, ok := course.LessonByID(target)
	if !ok {
		return nil, fmt.Errorf("unknown lesson %q", target)
	}
	return []LessonResult{RunLesson(course, lesson, cfg)}, nil
}
