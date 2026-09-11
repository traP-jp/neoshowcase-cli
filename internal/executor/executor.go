package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli"
)

type Executor struct {
	output *cli.Renderer
}

func New(output *cli.Renderer) *Executor {
	return &Executor{output: output}
}

func (e *Executor) Execute(ctx context.Context, invocation *cli.Invocation) error {
	switch command := invocation.Selected.(type) {
	case *cli.VersionCommand:
		return e.output.Version(invocation.Version)
	case *cli.AppListCommand:
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.AppList(ctx, options)
	case *cli.AppGetCommand:
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.AppGet(ctx, options, command.Application)
	case *cli.AppLogsCommand:
		return e.executeAppLogs(ctx, invocation, command)
	case *cli.AppStartCommand:
		if err := requireMutable(invocation.Command.Global); err != nil {
			return err
		}
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.AppStart(ctx, options, command.Application)
	case *cli.AppStopCommand:
		if err := requireMutable(invocation.Command.Global); err != nil {
			return err
		}
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.AppStop(ctx, options, command.Application)
	case *cli.AppRestartCommand:
		if err := requireMutable(invocation.Command.Global); err != nil {
			return err
		}
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.AppRestart(ctx, options, command.Application)
	case *cli.AppRebuildCommand:
		return e.executeAppRebuild(ctx, invocation, command)
	case *cli.BuildListCommand:
		if command.Page < 0 || command.Limit <= 0 {
			return cli.Fail(cli.ExitUsage, "--page must be non-negative and --limit must be positive")
		}
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.BuildList(ctx, options, command.Application, command.Page, command.Limit)
	case *cli.BuildGetCommand:
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.BuildGet(ctx, options, command.BuildID)
	case *cli.BuildLogsCommand:
		return e.executeBuildLogs(ctx, invocation, command)
	case *cli.BuildWaitCommand:
		return e.executeBuildWait(ctx, invocation, command)
	case *cli.BuildWatchCommand:
		return e.executeBuildWatch(ctx, invocation, command)
	case *cli.BuildRetryCommand:
		return e.executeBuildRetry(ctx, invocation, command)
	case *cli.BuildCancelCommand:
		if err := requireMutable(invocation.Command.Global); err != nil {
			return err
		}
		options, err := e.connectionOptions(invocation)
		if err != nil {
			return err
		}
		return e.BuildCancel(ctx, options, command.BuildID)
	default:
		return fmt.Errorf("unsupported command %q", invocation.CommandName)
	}
}

func requireMutable(options cli.GlobalOptions) error {
	if !options.AllowMutableOperation {
		return cli.Fail(cli.ExitUsage, "this command changes NeoShowcase state; pass --allow-mutable-operation for this invocation")
	}
	return nil
}

func validateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return cli.Fail(cli.ExitUsage, "--timeout must be greater than zero")
	}
	return nil
}

func parseSince(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if duration, err := time.ParseDuration(value); err == nil {
		if duration < 0 {
			return time.Time{}, cli.Fail(cli.ExitUsage, "--since duration must not be negative")
		}
		return now.Add(-duration), nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, cli.Fail(cli.ExitUsage, "--since must be RFC 3339 or a relative duration: %v", err)
	}
	return parsed, nil
}

func (e *Executor) executeAppLogs(ctx context.Context, invocation *cli.Invocation, command *cli.AppLogsCommand) error {
	if command.Tail < 0 {
		return cli.Fail(cli.ExitUsage, "--tail must not be negative")
	}
	if command.Follow {
		if err := validateTimeout(command.Timeout); err != nil {
			return err
		}
	}
	since, err := parseSince(command.Since, time.Now())
	if err != nil {
		return err
	}
	options, err := e.connectionOptions(invocation)
	if err != nil {
		return err
	}
	cancel := func() {}
	if command.Follow {
		ctx, cancel = context.WithTimeout(ctx, command.Timeout)
	}
	defer cancel()
	return e.AppLogs(ctx, options, command.Application, command.Follow, command.Tail, since)
}

func (e *Executor) executeAppRebuild(ctx context.Context, invocation *cli.Invocation, command *cli.AppRebuildCommand) error {
	if err := requireMutable(invocation.Command.Global); err != nil {
		return err
	}
	wait := command.Wait || command.Logs
	if wait {
		if err := validateTimeout(command.Timeout); err != nil {
			return err
		}
	}
	options, err := e.connectionOptions(invocation)
	if err != nil {
		return err
	}
	cancel := func() {}
	if wait {
		ctx, cancel = context.WithTimeout(ctx, command.Timeout)
	}
	defer cancel()
	return e.AppRebuild(ctx, options, command.Application, command.Commit, wait, command.Logs)
}

func (e *Executor) executeBuildLogs(ctx context.Context, invocation *cli.Invocation, command *cli.BuildLogsCommand) error {
	if err := validateTimeout(command.Timeout); err != nil {
		return err
	}
	options, err := e.connectionOptions(invocation)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, command.Timeout)
	defer cancel()
	return e.BuildLogs(ctx, options, command.BuildID, command.Follow)
}

func (e *Executor) executeBuildWait(ctx context.Context, invocation *cli.Invocation, command *cli.BuildWaitCommand) error {
	if err := validateTimeout(command.Timeout); err != nil {
		return err
	}
	options, err := e.connectionOptions(invocation)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, command.Timeout)
	defer cancel()
	return e.BuildWait(ctx, options, command.BuildID, command.Logs)
}

func (e *Executor) executeBuildWatch(ctx context.Context, invocation *cli.Invocation, command *cli.BuildWatchCommand) error {
	if strings.TrimSpace(command.Commit) == "" {
		return cli.Fail(cli.ExitUsage, "--commit is required")
	}
	if err := validateTimeout(command.Timeout); err != nil {
		return err
	}
	options, err := e.connectionOptions(invocation)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, command.Timeout)
	defer cancel()
	return e.BuildWatch(ctx, options, command.Application, command.Commit, command.Logs)
}

func (e *Executor) executeBuildRetry(ctx context.Context, invocation *cli.Invocation, command *cli.BuildRetryCommand) error {
	if err := requireMutable(invocation.Command.Global); err != nil {
		return err
	}
	wait := command.Wait || command.Logs
	if wait {
		if err := validateTimeout(command.Timeout); err != nil {
			return err
		}
	}
	options, err := e.connectionOptions(invocation)
	if err != nil {
		return err
	}
	cancel := func() {}
	if wait {
		ctx, cancel = context.WithTimeout(ctx, command.Timeout)
	}
	defer cancel()
	return e.BuildRetry(ctx, options, command.BuildID, wait, command.Logs)
}
