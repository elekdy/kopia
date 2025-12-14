# Windows Task Scheduler Integration Design

## Overview

This document outlines the design for integrating Kopia with Windows Task Scheduler to enable scheduled backups without requiring a continuously running server or UI application.

## Current State

Kopia currently has two ways to schedule backups:

1. **Server Mode**: `kopia server start` with scheduling policies - requires continuous operation
2. **KopiaUI**: Desktop app with auto-launch - requires app to be running

## Problem Statement

Users who want scheduled backups (e.g., daily at 8 AM) on Windows currently need to:
- Keep `kopia server start` running continuously, OR
- Keep KopiaUI running in the background, OR
- Manually run `kopia snapshot create` on schedule

This is not ideal for Windows users who expect native Task Scheduler integration.

## Proposed Solution

Add Windows Task Scheduler integration via new CLI commands that:
1. Create/update/delete Windows scheduled tasks
2. Configure tasks to run `kopia snapshot create` at specified times
3. Support the same scheduling expressions as policies (time-of-day, interval, cron)
4. Work with existing Kopia authentication mechanisms

## Design

### New CLI Commands

```
kopia schedule set <source> [flags]
  --time HH:MM            Time of day to run backup (can be specified multiple times)
  --interval DURATION     Interval between backups (e.g., 24h, 12h)
  --cron EXPRESSION       Cron expression (Windows Task Scheduler format)
  --task-name NAME        Custom task name (default: kopia-backup-<hash>)
  --description TEXT      Task description
  
kopia schedule list
  List all Kopia-created scheduled tasks
  
kopia schedule remove <task-name>
  Remove a scheduled task
  
kopia schedule status [<task-name>]
  Show status of scheduled task(s)
```

### Implementation Components

1. **Task Manager Interface** (`internal/scheduler/task/task.go`)
   ```go
   type TaskScheduler interface {
       Create(ctx context.Context, opts TaskOptions) error
       Delete(ctx context.Context, taskName string) error
       List(ctx context.Context) ([]TaskInfo, error)
       Status(ctx context.Context, taskName string) (TaskStatus, error)
   }
   ```

2. **Windows Implementation** (`internal/scheduler/task/task_windows.go`)
   - Use Windows Task Scheduler API via PowerShell or `schtasks.exe`
   - Create tasks that execute `kopia snapshot create <source>`
   - Handle credential management (run as current user)
   - Support multiple schedule types (daily, weekly, on-startup, etc.)

3. **Other OS Implementations**
   - **macOS**: Use `launchd` (`.plist` files in `~/Library/LaunchAgents/`)
   - **Linux**: Use systemd timers or cron
   - **Generic**: Stub implementation with helpful error messages

### Security Considerations

1. **Credential Storage**
   - Task runs as current user
   - Kopia repository password stored using existing credential manager
   - No passwords in task definitions

2. **Task Permissions**
   - Tasks created only for current user (no SYSTEM tasks)
   - Limited to executing Kopia binary from known locations

3. **Validation**
   - Verify Kopia binary path and permissions
   - Validate source paths exist and are accessible
   - Check repository connectivity before creating task

### Windows Task Scheduler Details

#### Task Creation via schtasks.exe
```powershell
schtasks /create /tn "Kopia-Backup-C-Users" /tr "\"C:\Path\To\kopia.exe\" snapshot create C:\Users" /sc daily /st 08:00
```

#### Task Creation via PowerShell
```powershell
$action = New-ScheduledTaskAction -Execute "kopia.exe" -Argument "snapshot create C:\Users"
$trigger = New-ScheduledTaskTrigger -Daily -At 8:00AM
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "Kopia-Backup"
```

### Integration with Existing Policies

The scheduled task feature should:
- Be independent of server mode scheduling policies
- Work with repository connections (direct or server)
- Support both local and remote repositories
- Honor existing snapshot policies (retention, compression, etc.)

### User Experience

#### First-time setup
```bash
# Connect to repository
kopia repository connect filesystem --path /path/to/repo

# Create scheduled backup
kopia schedule set C:\Users --time 08:00 --description "Daily backup at 8 AM"
```

#### List schedules
```bash
$ kopia schedule list
TASK NAME                  SOURCE      SCHEDULE           LAST RUN            NEXT RUN            STATUS
kopia-backup-c-users       C:\Users    Daily at 08:00     2024-01-15 08:00    2024-01-16 08:00    Success
```

#### Remove schedule
```bash
kopia schedule remove kopia-backup-c-users
```

## Alternatives Considered

1. **Keep server-only approach**: Rejected - requires continuous operation
2. **Use external scheduler only**: Rejected - inconsistent cross-platform experience
3. **Extend KopiaUI only**: Rejected - doesn't help CLI-only users

## Future Enhancements

1. Multiple schedules per source (e.g., hourly during work hours, daily backup overnight)
2. Conditional scheduling (e.g., only when on AC power, only when on specific network)
3. Task chaining (e.g., backup followed by maintenance)
4. Email/notification on task completion or failure
5. Integration with Windows Event Log

## Testing Strategy

1. Unit tests for task creation/deletion/listing
2. Integration tests with actual Task Scheduler (Windows only)
3. Mock tests for other platforms
4. Manual testing on different Windows versions (10, 11, Server 2019/2022)

## Documentation Updates

1. New CLI command reference pages
2. Getting Started guide update with scheduling example
3. Platform-specific guides (Windows, macOS, Linux)
4. Migration guide from server mode to task scheduler
