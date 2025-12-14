//go:build !windows

package task

import (
	"context"
	"runtime"

	"github.com/pkg/errors"
)

type unsupportedTaskScheduler struct{}

func newPlatformScheduler() TaskScheduler {
	return &unsupportedTaskScheduler{}
}

func (u *unsupportedTaskScheduler) Create(ctx context.Context, opts TaskOptions) error {
	return errors.Errorf("native task scheduling is not yet implemented for %s", runtime.GOOS)
}

func (u *unsupportedTaskScheduler) Delete(ctx context.Context, taskName string) error {
	return errors.Errorf("native task scheduling is not yet implemented for %s", runtime.GOOS)
}

func (u *unsupportedTaskScheduler) List(ctx context.Context) ([]TaskInfo, error) {
	return nil, errors.Errorf("native task scheduling is not yet implemented for %s", runtime.GOOS)
}

func (u *unsupportedTaskScheduler) Status(ctx context.Context, taskName string) (TaskStatus, error) {
	return TaskStatus{}, errors.Errorf("native task scheduling is not yet implemented for %s", runtime.GOOS)
}
