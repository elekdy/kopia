//go:build windows

package task

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type windowsTaskScheduler struct{}

func newPlatformScheduler() TaskScheduler {
	return &windowsTaskScheduler{}
}

// Create creates a new scheduled task using Windows Task Scheduler.
func (w *windowsTaskScheduler) Create(ctx context.Context, opts TaskOptions) error {
	if opts.TaskName == "" {
		return errors.New("task name is required")
	}

	if opts.Command == "" {
		return errors.New("command is required")
	}

	// Build the task action (command to execute)
	taskAction := escapeWindowsArg(opts.Command)
	if len(opts.Arguments) > 0 {
		// Escape each argument according to Windows cmd.exe rules
		for _, arg := range opts.Arguments {
			taskAction += " " + escapeWindowsArg(arg)
		}
	}

	// Determine the schedule type and build schtasks command
	var scheduleArgs []string

	if opts.Schedule.CronExpression != "" {
		return errors.New("cron expressions not yet supported for Windows Task Scheduler")
	}

	if len(opts.Schedule.TimesOfDay) > 0 {
		// For multiple times of day, we need to create multiple triggers
		// For now, just use the first one
		timeOfDay := opts.Schedule.TimesOfDay[0]

		scheduleArgs = []string{
			"/sc", "daily",
			"/st", timeOfDay,
		}
	} else if opts.Schedule.Interval > 0 {
		// Convert interval to minutes
		minutes := int(opts.Schedule.Interval.Minutes())
		if minutes < 1 {
			return errors.Errorf("interval must be at least 1 minute, got %v", opts.Schedule.Interval)
		}

		scheduleArgs = []string{
			"/sc", "minute",
			"/mo", strconv.Itoa(minutes),
		}
	} else {
		return errors.New("schedule must specify either TimesOfDay, Interval, or CronExpression")
	}

	// Build the full schtasks command
	args := []string{
		"/create",
		"/tn", opts.TaskName,
		"/tr", taskAction,
		"/f", // Force creation (overwrite if exists)
	}

	args = append(args, scheduleArgs...)

	if opts.Description != "" {
		args = append(args, "/desc", opts.Description)
	}

	// Add working directory if specified
	if opts.WorkingDirectory != "" {
		args = append(args, "/sd", opts.WorkingDirectory)
	}

	// Run as current user by default
	// Note: Leaving /ru empty makes it run as the current user
	// We could also explicitly specify the current user with "/ru", username
	if opts.RunAsUser != "" {
		args = append(args, "/ru", opts.RunAsUser)
	}

	log(ctx).Debugf("Creating scheduled task with schtasks: %v", args)

	//nolint:gosec // Command is constructed from validated inputs
	cmd := exec.CommandContext(ctx, "schtasks", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to create scheduled task: %s", string(output))
	}

	log(ctx).Infof("Successfully created scheduled task '%s'", opts.TaskName)

	return nil
}

// Delete removes a scheduled task.
func (w *windowsTaskScheduler) Delete(ctx context.Context, taskName string) error {
	if taskName == "" {
		return errors.New("task name is required")
	}

	args := []string{
		"/delete",
		"/tn", taskName,
		"/f", // Force deletion without confirmation
	}

	log(ctx).Debugf("Deleting scheduled task with schtasks: %v", args)

	//nolint:gosec // Command is constructed from validated inputs
	cmd := exec.CommandContext(ctx, "schtasks", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to delete scheduled task: %s", string(output))
	}

	log(ctx).Infof("Successfully deleted scheduled task '%s'", taskName)

	return nil
}

// List returns all Kopia-created scheduled tasks.
func (w *windowsTaskScheduler) List(ctx context.Context) ([]TaskInfo, error) {
	// Query all tasks with kopia in the name
	args := []string{
		"/query",
		"/fo", "list",
		"/v",
		"/tn", "kopia-*",
	}

	log(ctx).Debugf("Querying scheduled tasks with schtasks: %v", args)

	//nolint:gosec // Command is constructed from validated inputs
	cmd := exec.CommandContext(ctx, "schtasks", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If no tasks found, schtasks returns an error
		if strings.Contains(string(output), "ERROR: The system cannot find") {
			return []TaskInfo{}, nil
		}

		return nil, errors.Wrapf(err, "failed to query scheduled tasks: %s", string(output))
	}

	return parseTaskList(string(output)), nil
}

// Status returns the status of a scheduled task.
func (w *windowsTaskScheduler) Status(ctx context.Context, taskName string) (TaskStatus, error) {
	if taskName == "" {
		return TaskStatus{}, errors.New("task name is required")
	}

	args := []string{
		"/query",
		"/fo", "list",
		"/v",
		"/tn", taskName,
	}

	log(ctx).Debugf("Querying task status with schtasks: %v", args)

	//nolint:gosec // Command is constructed from validated inputs
	cmd := exec.CommandContext(ctx, "schtasks", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return TaskStatus{}, errors.Wrapf(err, "failed to query task status: %s", string(output))
	}

	return parseTaskStatus(taskName, string(output)), nil
}

// parseTaskList parses the output of "schtasks /query /fo list /v".
func parseTaskList(output string) []TaskInfo {
	var tasks []TaskInfo

	// Split by task (tasks are separated by blank lines)
	taskBlocks := strings.Split(output, "\r\n\r\n")

	for _, block := range taskBlocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		task := parseTaskBlock(block)
		if task.TaskName != "" {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// parseTaskBlock parses a single task block from schtasks output.
func parseTaskBlock(block string) TaskInfo {
	task := TaskInfo{}

	lines := strings.Split(block, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "TaskName":
			// Remove leading path separator if present
			task.TaskName = strings.TrimPrefix(value, "\\")
		case "Task To Run":
			// Parse command and arguments
			if cmdParts := parseCommand(value); len(cmdParts) > 0 {
				task.Command = cmdParts[0]
				if len(cmdParts) > 1 {
					task.Arguments = cmdParts[1:]
				}
			}
		case "Status":
			task.Enabled = value == "Ready" || value == "Running"
		case "Last Run Time":
			if t := parseTime(value); t != nil {
				task.LastRunTime = t
			}
		case "Next Run Time":
			if t := parseTime(value); t != nil {
				task.NextRunTime = t
			}
		case "Last Result":
			if code, err := strconv.Atoi(value); err == nil {
				task.LastResult = code
			}
		case "Comment":
			task.Description = value
		case "Schedule Type":
			// Build schedule description
			task.Schedule = value
		case "Start Time":
			if task.Schedule != "" && value != "N/A" {
				task.Schedule += " at " + value
			}
		case "Repeat: Every":
			if value != "N/A" && value != "Disabled" {
				task.Schedule = "Every " + value
			}
		}
	}

	return task
}

// parseTaskStatus parses task status from schtasks output.
func parseTaskStatus(taskName, output string) TaskStatus {
	task := parseTaskBlock(output)

	status := TaskStatus{
		TaskName:    taskName,
		LastRunTime: task.LastRunTime,
		NextRunTime: task.NextRunTime,
		LastResult:  task.LastResult,
	}

	if task.Enabled {
		status.State = "Ready"
	} else {
		status.State = "Disabled"
	}

	if status.LastResult == 0 {
		status.LastResultMessage = "Success"
	} else {
		status.LastResultMessage = fmt.Sprintf("Exit code: %d", status.LastResult)
	}

	return status
}

// parseCommand parses a command string into command and arguments.
func parseCommand(cmd string) []string {
	// Simple parsing - handles quoted paths
	re := regexp.MustCompile(`"[^"]*"|[^\s]+`)
	matches := re.FindAllString(cmd, -1)

	for i, match := range matches {
		matches[i] = strings.Trim(match, "\"")
	}

	return matches
}

// parseTime parses a time string from schtasks output.
func parseTime(timeStr string) *time.Time {
	if timeStr == "N/A" || timeStr == "Disabled" {
		return nil
	}

	// Try various time formats used by schtasks
	formats := []string{
		"1/2/2006 3:04:05 PM",
		"1/2/2006 15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return &t
		}
	}

	return nil
}

// escapeWindowsArg escapes a command line argument for Windows.
// It follows the rules documented by Microsoft for CommandLineToArgvW.
func escapeWindowsArg(arg string) string {
	// If the argument is empty, return empty quotes
	if arg == "" {
		return "\"\""
	}

	// Check if the argument needs quoting
	needsQuote := false
	for _, r := range arg {
		if r == ' ' || r == '\t' || r == '\n' || r == '\v' || r == '"' {
			needsQuote = true
			break
		}
	}

	// If no special characters and no quotes, return as-is
	if !needsQuote && !strings.Contains(arg, "\"") {
		return arg
	}

	// Build the escaped argument
	var result strings.Builder
	result.WriteRune('"')

	backslashCount := 0

	for _, r := range arg {
		switch r {
		case '\\':
			// Count consecutive backslashes
			backslashCount++
		case '"':
			// Escape all preceding backslashes and the quote
			for i := 0; i <= backslashCount; i++ {
				result.WriteRune('\\')
			}

			result.WriteRune(r)
			backslashCount = 0
		default:
			// Write any accumulated backslashes
			for i := 0; i < backslashCount; i++ {
				result.WriteRune('\\')
			}

			result.WriteRune(r)
			backslashCount = 0
		}
	}

	// Escape trailing backslashes before closing quote
	for i := 0; i < backslashCount; i++ {
		result.WriteRune('\\')
		result.WriteRune('\\')
	}

	result.WriteRune('"')

	return result.String()
}
