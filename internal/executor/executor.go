package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type Executor struct {
	output *cli.Renderer
}

func New(output *cli.Renderer) *Executor {
	return &Executor{output: output}
}

func (e *Executor) Execute(ctx context.Context, invocation *model.Invocation) error {
	if invocation.Connection.Insecure {
		e.output.Warn("TLS certificate verification is disabled")
	}
	connection := invocation.Connection

	switch command := invocation.Command.(type) {
	case model.VersionCommand:
		return e.output.Version(command.Version)
	case appmodel.ListCommand:
		return e.AppList(ctx, connection)
	case appmodel.GetCommand:
		return e.AppGet(ctx, connection, command.Application)
	case appmodel.LogsCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.AppLogs(ctx, connection, command.Application, command.Follow, command.Tail, command.Since)
	case appmodel.StartCommand:
		return e.AppStart(ctx, connection, command.Application)
	case appmodel.StopCommand:
		return e.AppStop(ctx, connection, command.Application)
	case appmodel.RestartCommand:
		return e.AppRestart(ctx, connection, command.Application)
	case appmodel.RebuildCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.AppRebuild(ctx, connection, command.Application, command.Commit, command.Wait, command.Logs)
	case buildmodel.ListCommand:
		return e.BuildList(ctx, connection, command.Application, command.Page, command.Limit)
	case buildmodel.GetCommand:
		return e.BuildGet(ctx, connection, command.BuildID)
	case buildmodel.LogsCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.BuildLogs(ctx, connection, command.BuildID, command.Follow)
	case buildmodel.WaitCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.BuildWait(ctx, connection, command.BuildID, command.Logs)
	case buildmodel.WatchCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.BuildWatch(ctx, connection, command.Application, command.Commit, command.Logs)
	case buildmodel.RetryCommand:
		ctx, cancel := commandContext(ctx, command.Timeout)
		defer cancel()
		return e.BuildRetry(ctx, connection, command.BuildID, command.Wait, command.Logs)
	case buildmodel.CancelCommand:
		return e.BuildCancel(ctx, connection, command.BuildID)
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
