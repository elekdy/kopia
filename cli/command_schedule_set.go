package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/scheduler/task"
)

type commandScheduleSet struct {
	sourcePath       string
	timesOfDay       []string
	interval         time.Duration
	cronExpression   string
	taskName         string
	taskDescription  string
	workingDirectory string

	out textOutput
}

func (c *commandScheduleSet) setup(svc advancedAppServices, parent commandParent) {
	cmd := parent.Command("set", "Create or update a scheduled backup task")

	cmd.Arg("source", "Path to backup (e.g., C:\\Users or /home/user)").Required().StringVar(&c.sourcePath)
	cmd.Flag("time", "Time of day to run backup in HH:MM format (e.g., 08:00). Can be specified multiple times.").StringsVar(&c.timesOfDay)
	cmd.Flag("interval", "Interval between backups (e.g., 24h, 12h, 1h)").DurationVar(&c.interval)
	cmd.Flag("cron", "Cron expression for scheduling").StringVar(&c.cronExpression)
	cmd.Flag("task-name", "Custom task name (default: auto-generated from source path)").StringVar(&c.taskName)
	cmd.Flag("description", "Task description").StringVar(&c.taskDescription)
	cmd.Flag("working-dir", "Working directory for the task").StringVar(&c.workingDirectory)

	cmd.Action(svc.noRepositoryAction(c.run))

	c.out.setup(svc)
}

func (c *commandScheduleSet) run(ctx context.Context) error {
	// Validate inputs
	if len(c.timesOfDay) == 0 && c.interval == 0 && c.cronExpression == "" {
		return errors.New("must specify at least one of: --time, --interval, or --cron")
	}

	if len(c.timesOfDay) > 0 && c.interval > 0 {
		return errors.New("cannot specify both --time and --interval")
	}

	// Validate times of day format
	for _, tod := range c.timesOfDay {
		if !isValidTimeOfDay(tod) {
			return errors.Errorf("invalid time format %q, must be HH:MM (e.g., 08:00)", tod)
		}
	}

	// Generate task name if not specified
	if c.taskName == "" {
		c.taskName = generateTaskName(c.sourcePath)
	}

	// Validate that task name starts with "kopia-" prefix
	if !strings.HasPrefix(c.taskName, "kopia-") {
		c.taskName = "kopia-" + c.taskName
	}

	// Get the path to the current kopia executable
	kopiaExe, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "unable to determine kopia executable path")
	}

	// Resolve to absolute path
	kopiaExe, err = filepath.Abs(kopiaExe)
	if err != nil {
		return errors.Wrap(err, "unable to resolve kopia executable path")
	}

	// Build command arguments
	args := []string{"snapshot", "create", c.sourcePath}

	// Set default description if not provided
	if c.taskDescription == "" {
		c.taskDescription = fmt.Sprintf("Kopia backup of %s", c.sourcePath)
	}

	// Create task options
	opts := task.TaskOptions{
		TaskName:         c.taskName,
		Description:      c.taskDescription,
		Command:          kopiaExe,
		Arguments:        args,
		WorkingDirectory: c.workingDirectory,
		Schedule: task.ScheduleOptions{
			TimesOfDay:     c.timesOfDay,
			Interval:       c.interval,
			CronExpression: c.cronExpression,
		},
	}

	// Create the scheduled task
	scheduler := task.New()
	if err := scheduler.Create(ctx, opts); err != nil {
		return errors.Wrap(err, "failed to create scheduled task")
	}

	c.out.printStdout("Successfully created scheduled task '%s'\n", c.taskName)
	c.out.printStdout("  Source: %s\n", c.sourcePath)

	if len(c.timesOfDay) > 0 {
		c.out.printStdout("  Schedule: Daily at %s\n", strings.Join(c.timesOfDay, ", "))
	} else if c.interval > 0 {
		c.out.printStdout("  Schedule: Every %v\n", c.interval)
	} else if c.cronExpression != "" {
		c.out.printStdout("  Schedule: %s\n", c.cronExpression)
	}

	c.out.printStdout("  Command: %s %s\n", kopiaExe, strings.Join(args, " "))
	c.out.printStdout("\nUse 'kopia schedule list' to view all scheduled tasks.\n")

	return nil
}

// generateTaskName generates a task name from a source path.
func generateTaskName(sourcePath string) string {
	// Clean the path and remove special characters
	cleaned := strings.Map(func(r rune) rune {
		if r == ':' || r == '\\' || r == '/' {
			return '-'
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}

		return -1
	}, sourcePath)

	cleaned = strings.ToLower(cleaned)
	cleaned = strings.Trim(cleaned, "-")

	// Add a short hash to ensure uniqueness
	h := sha256.Sum256([]byte(sourcePath))
	hash := hex.EncodeToString(h[:4])

	return fmt.Sprintf("kopia-backup-%s-%s", cleaned, hash)
}

// isValidTimeOfDay checks if a string is in HH:MM format.
func isValidTimeOfDay(s string) bool {
	var hour, minute int

	_, err := fmt.Sscanf(s, "%d:%d", &hour, &minute)
	if err != nil {
		return false
	}

	return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}
