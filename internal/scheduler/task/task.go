// Package task provides OS-native task scheduling for Kopia snapshots.
package task

import (
	"context"
	"time"

	"github.com/kopia/kopia/repo/logging"
)

var log = logging.Module("scheduler/task")

// TaskScheduler provides an interface for managing scheduled tasks.
type TaskScheduler interface {
	// Create creates a new scheduled task.
	Create(ctx context.Context, opts TaskOptions) error

	// Delete removes a scheduled task by name.
	Delete(ctx context.Context, taskName string) error

	// List returns all Kopia-created scheduled tasks.
	List(ctx context.Context) ([]TaskInfo, error)

	// Status returns the status of a scheduled task.
	Status(ctx context.Context, taskName string) (TaskStatus, error)
}

// TaskOptions contains options for creating a scheduled task.
type TaskOptions struct {
	// TaskName is the name of the scheduled task.
	TaskName string

	// Description is a human-readable description of the task.
	Description string

	// Command is the full path to the kopia executable.
	Command string

	// Arguments are the arguments to pass to the command.
	Arguments []string

	// WorkingDirectory is the working directory for the task.
	WorkingDirectory string

	// Schedule defines when the task should run.
	Schedule ScheduleOptions

	// RunAsUser specifies the user account to run the task as.
	// If empty, runs as the current user.
	RunAsUser string
}

// ScheduleOptions defines when a task should run.
type ScheduleOptions struct {
	// TimesOfDay specifies specific times to run (e.g., "08:00", "20:30").
	TimesOfDay []string

	// Interval specifies the interval between runs (e.g., 24h, 12h).
	// Only used if TimesOfDay is empty.
	Interval time.Duration

	// CronExpression is a cron expression for scheduling.
	// Takes precedence over TimesOfDay and Interval if specified.
	CronExpression string
}

// TaskInfo contains information about a scheduled task.
type TaskInfo struct {
	// TaskName is the name of the scheduled task.
	TaskName string

	// Description is the task description.
	Description string

	// Command is the command that will be executed.
	Command string

	// Arguments are the command arguments.
	Arguments []string

	// Schedule describes when the task runs.
	Schedule string

	// Enabled indicates if the task is enabled.
	Enabled bool

	// LastRunTime is the last time the task ran.
	LastRunTime *time.Time

	// NextRunTime is the next scheduled run time.
	NextRunTime *time.Time

	// LastResult is the exit code of the last run.
	LastResult int
}

// TaskStatus contains the status of a scheduled task.
type TaskStatus struct {
	// TaskName is the name of the task.
	TaskName string

	// State is the current state of the task (e.g., "Ready", "Running", "Disabled").
	State string

	// LastRunTime is the last time the task ran.
	LastRunTime *time.Time

	// NextRunTime is the next scheduled run time.
	NextRunTime *time.Time

	// LastResult is the exit code of the last run.
	LastResult int

	// LastResultMessage is a human-readable message about the last run.
	LastResultMessage string
}

// New returns a platform-specific TaskScheduler implementation.
func New() TaskScheduler {
	return newPlatformScheduler()
}
