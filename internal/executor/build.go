package executor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
)

const pollInterval = 11 * time.Second

func (e *Executor) BuildList(ctx context.Context, options cli.ConnectionOptions, application string, page, limit int32) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}

	var builds []*api.Build
	names := map[string]string{}
	if application != "" {
		app, err := resolveApplication(ctx, client, application)
		if err != nil {
			return err
		}
		names[app.GetId()] = app.GetName()
		response, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: app.GetId()}))
		if err != nil {
			return rpcError("list application builds", err)
		}
		builds = response.Msg.GetBuilds()
	} else {
		response, err := client.GetAllBuilds(ctx, connect.NewRequest(&api.GetAllBuildsRequest{Page: page, Limit: limit}))
		if err != nil {
			return rpcError("list builds", err)
		}
		builds = response.Msg.GetBuilds()
		apps, err := client.GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
		if err != nil {
			return rpcError("list applications for build names", err)
		}
		for _, app := range apps.Msg.GetApplications() {
			names[app.GetId()] = app.GetName()
		}
	}

	sort.SliceStable(builds, func(i, j int) bool {
		return builds[i].GetQueuedAt().AsTime().After(builds[j].GetQueuedAt().AsTime())
	})
	if application != "" {
		start := int64(page) * int64(limit)
		if start >= int64(len(builds)) {
			builds = nil
		} else {
			end := start + int64(limit)
			if end > int64(len(builds)) {
				end = int64(len(builds))
			}
			builds = builds[int(start):int(end)]
		}
	}
	return e.output.Builds(builds, names)
}

func (e *Executor) BuildGet(ctx context.Context, options cli.ConnectionOptions, id string) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}
	build, err := getBuild(ctx, client, id)
	if err != nil {
		return err
	}
	return e.output.BuildDetails(build)
}

func getBuild(ctx context.Context, client apiClient, id string) (*api.Build, error) {
	response, err := client.GetBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: id}))
	if err != nil {
		return nil, rpcError("get build", err)
	}
	return response.Msg, nil
}

func terminal(status api.BuildStatus) bool {
	switch status {
	case api.BuildStatus_SUCCEEDED, api.BuildStatus_FAILED, api.BuildStatus_CANCELLED, api.BuildStatus_SKIPPED:
		return true
	default:
		return false
	}
}

func buildResult(build *api.Build) error {
	if build.GetStatus() == api.BuildStatus_SUCCEEDED {
		return nil
	}
	if terminal(build.GetStatus()) {
		return cli.Fail(cli.ExitFailure, "build %s finished with status %s", build.GetId(), build.GetStatus())
	}
	return fmt.Errorf("build %s is not in a terminal state (%s)", build.GetId(), build.GetStatus())
}

func waitTick(ctx context.Context) error {
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return contextError(ctx, ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (e *Executor) streamBuildLog(ctx context.Context, client apiClient, buildID string) error {
	stream, err := client.GetBuildLogStream(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: buildID}))
	if err != nil {
		return contextError(ctx, rpcError("open build log stream", err))
	}
	for stream.Receive() {
		if err := e.output.BuildLog(buildID, stream.Msg().GetLog(), true); err != nil {
			return err
		}
	}
	if err := stream.Err(); err != nil {
		return contextError(ctx, rpcError("build log stream disconnected", err))
	}
	if ctx.Err() != nil {
		return contextError(ctx, ctx.Err())
	}
	return nil
}

func (e *Executor) monitorBuild(ctx context.Context, client apiClient, initial *api.Build, logs bool) (*api.Build, error) {
	build := initial
	for build.GetStatus() == api.BuildStatus_QUEUED {
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
		var err error
		build, err = getBuild(ctx, client, build.GetId())
		if err != nil {
			return nil, contextError(ctx, err)
		}
	}
	if terminal(build.GetStatus()) {
		if logs {
			response, err := client.GetBuildLog(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
			if err != nil {
				return nil, contextError(ctx, rpcError("get build log", err))
			}
			if err := e.output.BuildLog(build.GetId(), response.Msg.GetLog(), true); err != nil {
				return nil, err
			}
		}
		return build, nil
	}
	if logs {
		if err := e.streamBuildLog(ctx, client, build.GetId()); err != nil {
			return nil, err
		}
		final, err := getBuild(ctx, client, build.GetId())
		if err != nil {
			return nil, contextError(ctx, err)
		}
		if !terminal(final.GetStatus()) {
			return nil, fmt.Errorf("build log stream ended while build %s remained %s", final.GetId(), final.GetStatus())
		}
		return final, nil
	}
	for !terminal(build.GetStatus()) {
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
		var err error
		build, err = getBuild(ctx, client, build.GetId())
		if err != nil {
			return nil, contextError(ctx, err)
		}
	}
	return build, nil
}

func (e *Executor) BuildLogs(ctx context.Context, options cli.ConnectionOptions, id string, follow bool) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}
	build, err := getBuild(ctx, client, id)
	if err != nil {
		return contextError(ctx, err)
	}
	if terminal(build.GetStatus()) {
		response, err := client.GetBuildLog(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
		if err != nil {
			return contextError(ctx, rpcError("get build log", err))
		}
		return e.output.BuildLog(build.GetId(), response.Msg.GetLog(), false)
	}
	_ = follow // In-progress logs always follow the sole server stream.
	final, err := e.monitorBuild(ctx, client, build, true)
	if err != nil {
		return err
	}
	return buildResult(final)
}

func (e *Executor) BuildWait(ctx context.Context, options cli.ConnectionOptions, id string, logs bool) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}
	build, err := getBuild(ctx, client, id)
	if err != nil {
		return contextError(ctx, err)
	}
	final, err := e.monitorBuild(ctx, client, build, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, "", logs); err != nil {
		return err
	}
	return buildResult(final)
}

func newestMatching(builds []*api.Build, commit string, excluded map[string]struct{}) *api.Build {
	var found *api.Build
	for _, build := range builds {
		if build.GetCommit() != commit {
			continue
		}
		if _, skip := excluded[build.GetId()]; skip {
			continue
		}
		if found == nil || build.GetQueuedAt().AsTime().After(found.GetQueuedAt().AsTime()) {
			found = build
		}
	}
	return found
}

func watchBuild(ctx context.Context, client apiClient, appID, commit string, excluded map[string]struct{}) (*api.Build, error) {
	for {
		response, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: appID}))
		if err != nil {
			return nil, contextError(ctx, rpcError("watch for build", err))
		}
		if found := newestMatching(response.Msg.GetBuilds(), commit, excluded); found != nil {
			return found, nil
		}
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
	}
}

func (e *Executor) BuildWatch(ctx context.Context, options cli.ConnectionOptions, application, commit string, logs bool) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}
	app, err := resolveApplication(ctx, client, application)
	if err != nil {
		return contextError(ctx, err)
	}
	build, err := watchBuild(ctx, client, app.GetId(), commit, nil)
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

func buildIDs(builds []*api.Build) map[string]struct{} {
	ids := make(map[string]struct{}, len(builds))
	for _, build := range builds {
		ids[build.GetId()] = struct{}{}
	}
	return ids
}

func (e *Executor) BuildRetry(ctx context.Context, options cli.ConnectionOptions, id string, wait, logs bool) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}
	build, err := getBuild(ctx, client, id)
	if err != nil {
		return contextError(ctx, err)
	}
	if !build.GetRetriable() {
		return fmt.Errorf("build %s is not retriable; no build was started", build.GetId())
	}
	var excluded map[string]struct{}
	if wait {
		before, err := client.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: build.GetApplicationId()}))
		if err != nil {
			return contextError(ctx, rpcError("snapshot builds before retry", err))
		}
		excluded = buildIDs(before.Msg.GetBuilds())
	}
	if _, err := client.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: build.GetApplicationId(), Commit: build.GetCommit()})); err != nil {
		return contextError(ctx, rpcError("retry build", err))
	}
	if !wait {
		return e.output.Mutation(cli.MutationResult{Operation: "build.retry", SourceBuildID: build.GetId(), ApplicationID: build.GetApplicationId(), Commit: build.GetCommit(), State: "requested"})
	}
	newBuild, err := watchBuild(ctx, client, build.GetApplicationId(), build.GetCommit(), excluded)
	if err != nil {
		return err
	}
	final, err := e.monitorBuild(ctx, client, newBuild, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, "", logs); err != nil {
		return err
	}
	return buildResult(final)
}

func (e *Executor) BuildCancel(ctx context.Context, options cli.ConnectionOptions, id string) error {
	client, err := newAPIClient(options)
	if err != nil {
		return err
	}
	build, err := getBuild(ctx, client, id)
	if err != nil {
		return err
	}
	result := cli.MutationResult{Operation: "build.cancel", BuildID: build.GetId(), ApplicationID: build.GetApplicationId(), Commit: build.GetCommit(), Status: build.GetStatus().String()}
	if terminal(build.GetStatus()) {
		result.State = "already terminal; no change"
		return e.output.Mutation(result)
	}
	if _, err := client.CancelBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()})); err != nil {
		return rpcError("cancel build", err)
	}
	result.State = "cancel requested"
	return e.output.Mutation(result)
}
