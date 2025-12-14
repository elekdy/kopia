package task_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/kopia/kopia/internal/scheduler/task"
	"github.com/stretchr/testify/require"
)

func TestTaskOptions_Validation(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Task scheduling validation tests only run on Windows")
	}

	ctx := context.Background()
	scheduler := task.New()

	t.Run("EmptyTaskName", func(t *testing.T) {
		opts := task.TaskOptions{
			Command: "/usr/bin/kopia",
			Schedule: task.ScheduleOptions{
				TimesOfDay: []string{"08:00"},
			},
		}

		err := scheduler.Create(ctx, opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "task name")
	})

	t.Run("EmptyCommand", func(t *testing.T) {
		opts := task.TaskOptions{
			TaskName: "test-task",
			Schedule: task.ScheduleOptions{
				TimesOfDay: []string{"08:00"},
			},
		}

		err := scheduler.Create(ctx, opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "command")
	})

	t.Run("NoSchedule", func(t *testing.T) {
		opts := task.TaskOptions{
			TaskName: "test-task",
			Command:  "/usr/bin/kopia",
			Schedule: task.ScheduleOptions{},
		}

		err := scheduler.Create(ctx, opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "schedule")
	})
}

func TestUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unsupported platform test only runs on non-Windows")
	}

	ctx := context.Background()
	scheduler := task.New()

	opts := task.TaskOptions{
		TaskName: "test-task",
		Command:  "/usr/bin/kopia",
		Schedule: task.ScheduleOptions{
			TimesOfDay: []string{"08:00"},
		},
	}

	err := scheduler.Create(ctx, opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")
}

func TestScheduleOptions_TimesOfDay(t *testing.T) {
	schedule := task.ScheduleOptions{
		TimesOfDay: []string{"08:00", "20:00"},
	}

	require.Len(t, schedule.TimesOfDay, 2)
	require.Equal(t, "08:00", schedule.TimesOfDay[0])
	require.Equal(t, "20:00", schedule.TimesOfDay[1])
}

func TestScheduleOptions_Interval(t *testing.T) {
	schedule := task.ScheduleOptions{
		Interval: 24 * time.Hour,
	}

	require.Equal(t, 24*time.Hour, schedule.Interval)
}

func TestScheduleOptions_Cron(t *testing.T) {
	schedule := task.ScheduleOptions{
		CronExpression: "0 8 * * *",
	}

	require.Equal(t, "0 8 * * *", schedule.CronExpression)
}

func TestTaskInfo_Structure(t *testing.T) {
	now := time.Now()

	info := task.TaskInfo{
		TaskName:    "test-task",
		Description: "Test task",
		Command:     "/usr/bin/kopia",
		Arguments:   []string{"snapshot", "create", "/data"},
		Schedule:    "Daily at 08:00",
		Enabled:     true,
		LastRunTime: &now,
		NextRunTime: &now,
		LastResult:  0,
	}

	require.Equal(t, "test-task", info.TaskName)
	require.Equal(t, "Test task", info.Description)
	require.Equal(t, "/usr/bin/kopia", info.Command)
	require.Len(t, info.Arguments, 3)
	require.True(t, info.Enabled)
	require.NotNil(t, info.LastRunTime)
	require.NotNil(t, info.NextRunTime)
	require.Equal(t, 0, info.LastResult)
}

func TestTaskStatus_Structure(t *testing.T) {
	now := time.Now()

	status := task.TaskStatus{
		TaskName:          "test-task",
		State:             "Ready",
		LastRunTime:       &now,
		NextRunTime:       &now,
		LastResult:        0,
		LastResultMessage: "Success",
	}

	require.Equal(t, "test-task", status.TaskName)
	require.Equal(t, "Ready", status.State)
	require.NotNil(t, status.LastRunTime)
	require.NotNil(t, status.NextRunTime)
	require.Equal(t, 0, status.LastResult)
	require.Equal(t, "Success", status.LastResultMessage)
}
