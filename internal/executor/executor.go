package executor

import (
	"context"
	"fmt"
	"time"

	appexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/app"
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type Executor struct {
	apps   *appexecutor.Executor
	builds *buildexecutor.Executor
	emit   model.Emit
}

func New(emit model.Emit) *Executor {
	builds := buildexecutor.New(emit)
	return &Executor{apps: appexecutor.New(emit, builds), builds: builds, emit: emit}
}

func (e *Executor) Execute(ctx context.Context, invocation *model.Invocation) error {
	connection := invocation.Connection

	switch command := invocation.Command.(type) {
	case model.VersionCommand:
		return e.emit(model.VersionResult{Version: command.Version})
	case appmodel.ListCommand:
		return e.apps.List(ctx, connection)
	case appmodel.GetCommand:
		return e.apps.Get(ctx, connection, command.Application)
	case appmodel.LogsCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.apps.Logs(ctx, connection, command.Application, command.Follow, command.Tail, command.Since)
	case appmodel.StartCommand:
		return e.apps.Start(ctx, connection, command.Application)
	case appmodel.StopCommand:
		return e.apps.Stop(ctx, connection, command.Application)
	case appmodel.RestartCommand:
		return e.apps.Restart(ctx, connection, command.Application)
	case appmodel.RebuildCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.apps.Rebuild(ctx, connection, command.Application, command.Commit, command.Wait, command.Logs)
	case buildmodel.ListCommand:
		return e.builds.List(ctx, connection, command.Application, command.Page, command.Limit)
	case buildmodel.GetCommand:
		return e.builds.Get(ctx, connection, command.BuildID)
	case buildmodel.LogsCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.builds.Logs(ctx, connection, command.BuildID, command.Follow)
	case buildmodel.WaitCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.builds.Wait(ctx, connection, command.BuildID, command.Logs)
	case buildmodel.WatchCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.builds.Watch(ctx, connection, command.Application, command.Commit, command.Logs)
	case buildmodel.RetryCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.builds.Retry(ctx, connection, command.BuildID, command.Wait, command.Logs)
	case buildmodel.CancelCommand:
		return e.builds.Cancel(ctx, connection, command.BuildID)
	default:
		return fmt.Errorf("unsupported command type %T", invocation.Command)
	}
}

func commandContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}
