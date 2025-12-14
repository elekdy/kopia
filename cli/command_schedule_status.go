package cli

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/scheduler/task"
)

type commandScheduleStatus struct {
	taskName string
	out      textOutput
}

func (c *commandScheduleStatus) setup(svc advancedAppServices, parent commandParent) {
	cmd := parent.Command("status", "Show status of a scheduled backup task")

	cmd.Arg("task-name", "Name of the task (optional, shows all if not specified)").StringVar(&c.taskName)

	cmd.Action(svc.noRepositoryAction(c.run))

	c.out.setup(svc)
}

func (c *commandScheduleStatus) run(ctx context.Context) error {
	scheduler := task.New()

	if c.taskName == "" {
		// Show status for all tasks
		tasks, err := scheduler.List(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to list scheduled tasks")
		}

		if len(tasks) == 0 {
			c.out.printStdout("No scheduled backup tasks found.\n")

			return nil
		}

		for i, t := range tasks {
			if i > 0 {
				c.out.printStdout("\n")
			}

			c.printTaskInfo(t)
		}

		return nil
	}

	// Show status for specific task
	status, err := scheduler.Status(ctx, c.taskName)
	if err != nil {
		return errors.Wrapf(err, "failed to get status for task '%s'", c.taskName)
	}

	c.printTaskStatus(status)

	return nil
}

func (c *commandScheduleStatus) printTaskInfo(t task.TaskInfo) {
	c.out.printStdout("Task: %s\n", t.TaskName)

	if t.Description != "" {
		c.out.printStdout("  Description: %s\n", t.Description)
	}

	c.out.printStdout("  Command: %s", t.Command)

	if len(t.Arguments) > 0 {
		for _, arg := range t.Arguments {
			c.out.printStdout(" %s", arg)
		}
	}

	c.out.printStdout("\n")

	if t.Schedule != "" {
		c.out.printStdout("  Schedule: %s\n", t.Schedule)
	}

	status := "Disabled"
	if t.Enabled {
		if t.LastResult == 0 || t.LastRunTime == nil {
			status = "Enabled"
		} else {
			status = "Enabled (Last run failed)"
		}
	}

	c.out.printStdout("  Status: %s\n", status)

	if t.LastRunTime != nil {
		c.out.printStdout("  Last Run: %s\n", t.LastRunTime.Format("2006-01-02 15:04:05"))
		c.out.printStdout("  Last Result: %d\n", t.LastResult)
	} else {
		c.out.printStdout("  Last Run: Never\n")
	}

	if t.NextRunTime != nil {
		c.out.printStdout("  Next Run: %s\n", t.NextRunTime.Format("2006-01-02 15:04:05"))
	}
}

func (c *commandScheduleStatus) printTaskStatus(s task.TaskStatus) {
	c.out.printStdout("Task: %s\n", s.TaskName)
	c.out.printStdout("  State: %s\n", s.State)

	if s.LastRunTime != nil {
		c.out.printStdout("  Last Run: %s\n", s.LastRunTime.Format("2006-01-02 15:04:05"))
		c.out.printStdout("  Last Result: %d (%s)\n", s.LastResult, s.LastResultMessage)
	} else {
		c.out.printStdout("  Last Run: Never\n")
	}

	if s.NextRunTime != nil {
		c.out.printStdout("  Next Run: %s\n", s.NextRunTime.Format("2006-01-02 15:04:05"))
	}
}
