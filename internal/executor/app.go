package executor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) AppList(ctx context.Context, options model.Connection) error {
	client := newAPIClient(options)
	response, err := client.GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
	if err != nil {
		return rpcError("list applications", err)
	}
	apps := response.Msg.GetApplications()
	sort.Slice(apps, func(i, j int) bool { return apps[i].GetName() < apps[j].GetName() })
	return e.output.Applications(apps)
}

func (e *Executor) AppGet(ctx context.Context, options model.Connection, identifier string) error {
	client := newAPIClient(options)
	app, err := resolveApplication(ctx, client, identifier)
	if err != nil {
		return err
	}
	return e.output.Application(app)
}

func resolveApplication(ctx context.Context, client apiClient, identifier string) (*api.Application, error) {
	response, err := client.GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
	if err != nil {
		return nil, rpcError("resolve application", err)
	}
	var matches []*api.Application
	for _, app := range response.Msg.GetApplications() {
		if app.GetId() == identifier {
			return app, nil
		}
		if app.GetName() == identifier {
			matches = append(matches, app)
		}
	}
	if len(matches) == 0 {
		return nil, cli.Fail(cli.ExitNotFound, "application %q not found", identifier)
	}
	if len(matches) > 1 {
		return nil, cli.Fail(cli.ExitNotFound, "application name %q is ambiguous (%d exact matches)", identifier, len(matches))
	}
	return matches[0], nil
}

func (e *Executor) AppLogs(ctx context.Context, options model.Connection, identifier string, follow bool, tail int32, since time.Time) error {
	client := newAPIClient(options)
	app, err := resolveApplication(ctx, client, identifier)
	if err != nil {
		return contextError(ctx, err)
	}
	historyBefore := time.Now().UTC()
	response, err := client.GetOutput(ctx, connect.NewRequest(&api.GetOutputRequest{ApplicationId: app.GetId(), Before: timestamppb.New(historyBefore), Limit: tail}))
	if err != nil {
		return contextError(ctx, rpcError("get application logs", err))
	}
	outputs := response.Msg.GetOutputs()
	sort.SliceStable(outputs, func(i, j int) bool { return outputs[i].GetTime().AsTime().Before(outputs[j].GetTime().AsTime()) })
	filtered := make([]*api.ApplicationOutput, 0, len(outputs))
	var cursor time.Time
	for _, output := range outputs {
		at := output.GetTime().AsTime()
		if at.After(cursor) {
			cursor = at
		}
		if since.IsZero() || !at.Before(since) {
			filtered = append(filtered, output)
		}
	}
	if err := e.output.ApplicationLogs(app.GetId(), filtered, follow); err != nil {
		return err
	}
	if !follow {
		return nil
	}
	if cursor.IsZero() {
		cursor = historyBefore
	} else {
		cursor = cursor.Add(time.Nanosecond)
	}
	if !since.IsZero() && since.After(cursor) {
		cursor = since
	}
	stream, err := client.GetOutputStream(ctx, connect.NewRequest(&api.GetOutputStreamRequest{ApplicationId: app.GetId(), Begin: timestamppb.New(cursor)}))
	if err != nil {
		return contextError(ctx, rpcError("follow application logs", err))
	}
	for stream.Receive() {
		if err := e.output.ApplicationLogs(app.GetId(), []*api.ApplicationOutput{stream.Msg()}, true); err != nil {
			return err
		}
	}
	if err := stream.Err(); err != nil {
		return contextError(ctx, rpcError("application log stream ended", err))
	}
	if ctx.Err() != nil {
		return contextError(ctx, ctx.Err())
	}
	return fmt.Errorf("application log stream ended unexpectedly")
}

func contextError(ctx context.Context, err error) error {
	if ctx.Err() == context.DeadlineExceeded {
		return cli.Fail(cli.ExitTimeout, "command timed out")
	}
	if ctx.Err() == context.Canceled {
		return cli.Fail(cli.ExitInterrupt, "command interrupted")
	}
	return err
}

func (e *Executor) AppStart(ctx context.Context, options model.Connection, identifier string) error {
	client := newAPIClient(options)
	app, err := resolveApplication(ctx, client, identifier)
	if err != nil {
		return err
	}
	if app.GetRunning() {
		return e.output.Mutation(appMutation("app.start", app, "already running; no change"))
	}
	if _, err := client.StartApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()})); err != nil {
		return rpcError("start application", err)
	}
	return e.output.Mutation(appMutation("app.start", app, "started"))
}

func (e *Executor) AppStop(ctx context.Context, options model.Connection, identifier string) error {
	client := newAPIClient(options)
	app, err := resolveApplication(ctx, client, identifier)
	if err != nil {
		return err
	}
	if !app.GetRunning() {
		return e.output.Mutation(appMutation("app.stop", app, "already stopped; no change"))
	}
	if _, err := client.StopApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()})); err != nil {
		return rpcError("stop application", err)
	}
	return e.output.Mutation(appMutation("app.stop", app, "stopped"))
}

func (e *Executor) AppRestart(ctx context.Context, options model.Connection, identifier string) error {
	client := newAPIClient(options)
	app, err := resolveApplication(ctx, client, identifier)
	if err != nil {
		return err
	}
	if !app.GetRunning() {
		return fmt.Errorf("application %q is stopped; refusing to start it (state check is non-atomic)", app.GetId())
	}
	if _, err := client.StartApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()})); err != nil {
		return rpcError("restart application (state check is non-atomic)", err)
	}
	return e.output.Mutation(appMutation("app.restart", app, "restart requested (state check was non-atomic)"))
}

func appMutation(operation string, app *api.Application, state string) cli.Mutation {
	return cli.Mutation{Operation: operation, ApplicationID: app.GetId(), ApplicationName: app.GetName(), Commit: app.GetCommit(), State: state}
}

func (e *Executor) AppRebuild(ctx context.Context, options model.Connection, identifier, commit string, wait, logs bool) error {
	client := newAPIClient(options)
	app, err := resolveApplication(ctx, client, identifier)
	if err != nil {
		return contextError(ctx, err)
	}
	selectedCommit := strings.TrimSpace(commit)
	if selectedCommit == "" {
		selectedCommit = strings.TrimSpace(app.GetCommit())
	}
	if selectedCommit == "" {
		return fmt.Errorf("application %s has no current commit; no build was started", app.GetId())
	}
	var excluded map[string]struct{}
	if wait {
		before, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
		if err != nil {
			return contextError(ctx, rpcError("snapshot builds before rebuild", err))
		}
		excluded = buildIDs(before.Msg.GetBuilds())
	}
	if _, err := client.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: app.GetId(), Commit: selectedCommit})); err != nil {
		return contextError(ctx, rpcError("rebuild application", err))
	}
	if !wait {
		return e.output.Mutation(cli.Mutation{Operation: "app.rebuild", ApplicationID: app.GetId(), ApplicationName: app.GetName(), Commit: selectedCommit, State: "requested"})
	}
	build, err := watchBuild(ctx, client, app.GetId(), selectedCommit, excluded)
	if err != nil {
		return err
	}
	final, err := e.monitorBuild(ctx, client, build, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, app.GetName(), logs); err != nil {
		return err
	}
	return buildResult(final)
}
