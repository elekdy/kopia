---
title: "Scheduled Backups"
linkTitle: "Scheduled Backups"
weight: 35
---

Kopia provides multiple ways to schedule automatic backups:

## Option 1: OS-Native Task Scheduling (Windows)

On Windows, Kopia can integrate with Windows Task Scheduler to run backups at specific times without requiring a continuously running server or UI application.

### Creating a Scheduled Backup

To create a scheduled backup that runs daily at 8:00 AM:

```shell
$ kopia schedule set C:\Users\MyUser --time 08:00 --description "Daily backup at 8 AM"
Successfully created scheduled task 'kopia-backup-c-users-myuser-a1b2c3d4'
  Source: C:\Users\MyUser
  Schedule: Daily at 08:00
  Command: C:\Program Files\kopia\kopia.exe snapshot create C:\Users\MyUser

Use 'kopia schedule list' to view all scheduled tasks.
```

### Scheduling Options

You can schedule backups using different methods:

#### Specific Time of Day

Run backup every day at a specific time:

```shell
$ kopia schedule set /path/to/backup --time 08:00
```

Run backup multiple times per day:

```shell
$ kopia schedule set /path/to/backup --time 08:00 --time 20:00
```

#### Interval-Based

Run backup at regular intervals:

```shell
$ kopia schedule set /path/to/backup --interval 24h  # Every 24 hours
$ kopia schedule set /path/to/backup --interval 12h  # Every 12 hours
$ kopia schedule set /path/to/backup --interval 1h   # Every hour
```

### Managing Scheduled Tasks

#### List All Scheduled Backups

```shell
$ kopia schedule list
TASK NAME                      SCHEDULE           LAST RUN            NEXT RUN            STATUS
kopia-backup-c-users           Daily at 08:00     2024-01-15 08:00    2024-01-16 08:00    Ready
kopia-backup-d-documents       Every 12h          2024-01-15 14:00    2024-01-16 02:00    Ready
```

#### Check Task Status

Get detailed status of a specific task:

```shell
$ kopia schedule status kopia-backup-c-users
Task: kopia-backup-c-users
  Description: Daily backup at 8 AM
  Command: C:\Program Files\kopia\kopia.exe snapshot create C:\Users
  Schedule: Daily at 08:00
  Status: Enabled
  Last Run: 2024-01-15 08:00:00
  Last Result: 0
  Next Run: 2024-01-16 08:00:00
```

#### Remove a Scheduled Task

```shell
$ kopia schedule remove kopia-backup-c-users
Successfully removed scheduled task 'kopia-backup-c-users'
```

### Important Notes

1. **Repository Connection Required**: You must first connect to a Kopia repository before creating scheduled tasks:
   ```shell
   $ kopia repository connect filesystem --path /path/to/repo
   ```

2. **Credentials**: Scheduled tasks use Kopia's existing credential management system. Make sure you have persisted credentials using the `--persist-credentials` flag when connecting to the repository.

3. **Task Naming**: Tasks are automatically named based on the source path. You can provide a custom name with `--task-name`:
   ```shell
   $ kopia schedule set C:\Users --task-name my-backup --time 08:00
   ```

4. **Windows Only**: OS-native task scheduling is currently only implemented for Windows. For other platforms, use Server Mode (see below).

## Option 2: Server Mode with Scheduling Policies

Kopia server mode provides scheduling through policies and requires a continuously running server process.

### Starting the Server

```shell
$ kopia server start
```

### Setting Scheduling Policies

Configure when snapshots should be taken:

```shell
$ kopia policy set --global --snapshot-time 08:00
$ kopia policy set --global --snapshot-interval 24h
$ kopia policy set --global --snapshot-time-crontab "0 8 * * *"
```

For more information about server mode and scheduling policies, see:
- [Repository Server](../../repository-server/)
- [Snapshot Policies](../../reference/command-line/common/policy-set/)

## Option 3: KopiaUI Auto-Launch

The KopiaUI desktop application can be configured to launch at system startup and run scheduled backups in the background.

1. Open KopiaUI
2. Go to Settings
3. Enable "Launch at startup"
4. Configure snapshot frequency in your policy settings

For more information, see [Getting Started](../../getting-started/).

## Comparison of Scheduling Methods

| Feature | OS-Native (Windows) | Server Mode | KopiaUI |
|---------|-------------------|-------------|---------|
| Platform | Windows only | All platforms | All platforms |
| Requires running process | No | Yes | Yes |
| Setup complexity | Simple | Moderate | Simple |
| Multiple schedules per source | No (yet) | Yes | Yes |
| Best for | Windows CLI users | Multi-user setups | Desktop users |

## Troubleshooting

### Task Not Running

1. Check task status: `kopia schedule status <task-name>`
2. Verify repository connection: `kopia repository status`
3. Check Windows Task Scheduler logs (Event Viewer → Task Scheduler)

### Permission Issues

Scheduled tasks run as the current user. Ensure:
- The user has access to the source paths
- The user has access to the repository storage
- Credentials are properly stored

### Task Running but Failing

Check the last result code:
```shell
$ kopia schedule status kopia-backup-c-users
  Last Result: 1 (Exit code: 1)
```

Common exit codes:
- 0: Success
- 1: General error (check Kopia logs)
- Other: See Windows Task Scheduler documentation
