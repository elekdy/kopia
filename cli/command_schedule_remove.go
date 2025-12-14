package cli

import (
	"context"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/scheduler/task"
)

type commandScheduleRemove struct {
	taskName string
	out      textOutput
}

func (c *commandScheduleRemove) setup(svc advancedAppServices, parent commandParent) {
	cmd := parent.Command("remove", "Remove a scheduled backup task").Alias("rm")

	cmd.Arg("task-name", "Name of the task to remove").Required().StringVar(&c.taskName)

	cmd.Action(svc.noRepositoryAction(c.run))

	c.out.setup(svc)
}

func (c *commandScheduleRemove) run(ctx context.Context) error {
	scheduler := task.New()

	// Check if task exists first
	_, err := scheduler.Status(ctx, c.taskName)
	if err != nil {
		return errors.Wrapf(err, "task '%s' not found", c.taskName)
	}

	// Delete the task
	if err := scheduler.Delete(ctx, c.taskName); err != nil {
		return errors.Wrap(err, "failed to remove scheduled task")
	}

	c.out.printStdout("Successfully removed scheduled task '%s'\n", c.taskName)

	return nil
}
