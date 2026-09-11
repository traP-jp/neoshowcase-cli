package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	appoutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output/app"
	buildoutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output/build"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser"
	appexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/app"
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

var version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr, version))
}

func run(ctx context.Context, args []string, out, errOut io.Writer, version string) int {
	commandParser := parser.New(out, errOut, version)
	invocation, err := commandParser.Parse(args)
	if err == nil && invocation != nil {
		err = execute(ctx, invocation, out)
	}
	if err == nil {
		return cli.ExitOK
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		commandParser.LogError(errors.New("command timed out"))
		return cli.ExitTimeout
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		commandParser.LogError(errors.New("command interrupted"))
		return cli.ExitInterrupt
	}
	commandParser.LogError(err)
	return cli.ExitCode(err)
}

func execute(ctx context.Context, invocation *model.Invocation, out io.Writer) error {
	writer := core.New(out, invocation.Output)
	builds := buildexecutor.New()
	apps := appexecutor.New(builds)
	connection := invocation.Connection

	switch command := invocation.Command.(type) {
	case model.VersionCommand:
		return output.RenderVersion(writer, model.VersionResult{Version: command.Version})
	case appmodel.ListCommand:
		result, err := apps.List(ctx, connection)
		if err != nil {
			return err
		}
		return appoutput.RenderList(writer, result)
	case appmodel.GetCommand:
		result, err := apps.Get(ctx, connection, command.Application)
		if err != nil {
			return err
		}
		return appoutput.RenderGet(writer, result)
	case appmodel.LogsCommand:
		commandCtx, cancel := commandContext(ctx, command.Timeout)
		err := apps.Logs(commandCtx, connection, command.Application, command.Follow, command.Tail, command.Since, func(result appmodel.LogsResult) error {
			return appoutput.RenderLogs(writer, result)
		})
		cancel()
		return err
	case appmodel.StartCommand:
		result, err := apps.Start(ctx, connection, command.Application)
		if err != nil {
			return err
		}
		return appoutput.RenderStart(writer, result)
	case appmodel.StopCommand:
		result, err := apps.Stop(ctx, connection, command.Application)
		if err != nil {
			return err
		}
		return appoutput.RenderStop(writer, result)
	case appmodel.RestartCommand:
		result, err := apps.Restart(ctx, connection, command.Application)
		if err != nil {
			return err
		}
		return appoutput.RenderRestart(writer, result)
	case appmodel.RebuildCommand:
		commandCtx, cancel := commandContext(ctx, command.Timeout)
		result, executeErr := apps.Rebuild(commandCtx, connection, command.Application, command.Commit, command.Wait, command.Logs, func(result buildmodel.LogResult) error {
			return buildoutput.RenderLogs(writer, result)
		})
		cancel()
		if err := appoutput.RenderRebuild(writer, result); err != nil {
			return err
		}
		return executeErr
	case buildmodel.ListCommand:
		result, err := builds.List(ctx, connection, command.Application, command.Page, command.Limit)
		if err != nil {
			return err
		}
		return buildoutput.RenderList(writer, result)
	case buildmodel.GetCommand:
		result, err := builds.Get(ctx, connection, command.BuildID)
		if err != nil {
			return err
		}
		return buildoutput.RenderGet(writer, result)
	case buildmodel.LogsCommand:
		commandCtx, cancel := commandContext(ctx, command.Timeout)
		err := builds.Logs(commandCtx, connection, command.BuildID, command.Follow, func(result buildmodel.LogResult) error {
			return buildoutput.RenderLogs(writer, result)
		})
		cancel()
		return err
	case buildmodel.WaitCommand:
		commandCtx, cancel := commandContext(ctx, command.Timeout)
		result, executeErr := builds.Wait(commandCtx, connection, command.BuildID, command.Logs, func(result buildmodel.LogResult) error {
			return buildoutput.RenderLogs(writer, result)
		})
		cancel()
		if result.Build.ID != "" {
			if err := buildoutput.RenderWait(writer, result); err != nil {
				return err
			}
		}
		return executeErr
	case buildmodel.WatchCommand:
		commandCtx, cancel := commandContext(ctx, command.Timeout)
		result, executeErr := builds.Watch(commandCtx, connection, command.Application, command.Commit, command.Logs, func(result buildmodel.LogResult) error {
			return buildoutput.RenderLogs(writer, result)
		})
		cancel()
		if result.Build.ID != "" {
			if err := buildoutput.RenderWatch(writer, result); err != nil {
				return err
			}
		}
		return executeErr
	case buildmodel.RetryCommand:
		commandCtx, cancel := commandContext(ctx, command.Timeout)
		result, executeErr := builds.Retry(commandCtx, connection, command.BuildID, command.Wait, command.Logs, func(result buildmodel.LogResult) error {
			return buildoutput.RenderLogs(writer, result)
		})
		cancel()
		if err := buildoutput.RenderRetry(writer, result); err != nil {
			return err
		}
		return executeErr
	case buildmodel.CancelCommand:
		result, err := builds.Cancel(ctx, connection, command.BuildID)
		if err != nil {
			return err
		}
		return buildoutput.RenderCancel(writer, result)
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
