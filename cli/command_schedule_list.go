package cli

import (
	"context"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/scheduler/task"
)

type commandScheduleList struct {
	out textOutput
}

func (c *commandScheduleList) setup(svc advancedAppServices, parent commandParent) {
	cmd := parent.Command("list", "List all scheduled backup tasks").Alias("ls")

	cmd.Action(svc.noRepositoryAction(c.run))

	c.out.setup(svc)
}

func (c *commandScheduleList) run(ctx context.Context) error {
	scheduler := task.New()

	tasks, err := scheduler.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list scheduled tasks")
	}

	if len(tasks) == 0 {
		c.out.printStdout("No scheduled backup tasks found.\n")
		c.out.printStdout("Use 'kopia schedule set <source> --time HH:MM' to create a scheduled backup.\n")

		return nil
	}

	// Print header
	c.out.printStdout("%-30s %-20s %-25s %-25s %-10s\n",
		"TASK NAME",
		"SCHEDULE",
		"LAST RUN",
		"NEXT RUN",
		"STATUS",
	)

	c.out.printStdout("%s\n", stringRepeat("-", 120))

	// Print each task
	for _, t := range tasks {
		lastRun := "Never"
		if t.LastRunTime != nil {
			lastRun = t.LastRunTime.Format("2006-01-02 15:04")
		}

		nextRun := "N/A"
		if t.NextRunTime != nil {
			nextRun = t.NextRunTime.Format("2006-01-02 15:04")
		}

		status := "Disabled"
		if t.Enabled {
			if t.LastResult == 0 || t.LastRunTime == nil {
				status = "Ready"
			} else {
				status = "Failed"
			}
		}

		c.out.printStdout("%-30s %-20s %-25s %-25s %-10s\n",
			truncateString(t.TaskName, 30),
			truncateString(t.Schedule, 20),
			lastRun,
			nextRun,
			status,
		)
	}

	c.out.printStdout("\nUse 'kopia schedule status <task-name>' for detailed information.\n")

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

func stringRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}

	return result
}
