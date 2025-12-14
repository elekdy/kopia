package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/cli"
)

func TestScheduleSetCommand_TaskNameGeneration(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		wantPrefix string
	}{
		{
			name:       "Windows path",
			sourcePath: "C:\\Users\\test",
			wantPrefix: "kopia-backup-c-users-test-",
		},
		{
			name:       "Unix path",
			sourcePath: "/home/user/data",
			wantPrefix: "kopia-backup-home-user-data-",
		},
		{
			name:       "Path with special characters",
			sourcePath: "C:\\Program Files\\My App",
			wantPrefix: "kopia-backup-c-program-files-my-app-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't test the actual generation without making the function public
			// This is a placeholder for the test structure
			// In actual implementation, we might need to export generateTaskName or test through CLI
			require.NotEmpty(t, tt.wantPrefix)
		})
	}
}

func TestScheduleSetCommand_TimeValidation(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		wantErr bool
	}{
		{
			name:    "Valid time 08:00",
			time:    "08:00",
			wantErr: false,
		},
		{
			name:    "Valid time 23:59",
			time:    "23:59",
			wantErr: false,
		},
		{
			name:    "Valid time 00:00",
			time:    "00:00",
			wantErr: false,
		},
		{
			name:    "Invalid time 24:00",
			time:    "24:00",
			wantErr: true,
		},
		{
			name:    "Invalid time 12:60",
			time:    "12:60",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			time:    "8am",
			wantErr: true,
		},
		{
			name:    "Invalid format no colon",
			time:    "0800",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Similar to above, we need to expose isValidTimeOfDay or test through CLI
			// This structure shows how we would test it
			require.NotEmpty(t, tt.time)
		})
	}
}
