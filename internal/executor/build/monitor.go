package build

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
)

const pollInterval = 11 * time.Second

func get(ctx context.Context, apiClient client.Client, id string) (*api.Build, error) {
	response, err := apiClient.GetBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: id}))
	if err != nil {
		return nil, client.RPCError("get build", err)
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

func Result(build *api.Build) error {
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
		return client.ContextError(ctx, ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (e *Executor) streamLog(ctx context.Context, apiClient client.Client, buildID string) error {
	stream, err := apiClient.GetBuildLogStream(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: buildID}))
	if err != nil {
		return client.ContextError(ctx, client.RPCError("open build log stream", err))
	}
	for stream.Receive() {
		if err := e.output.BuildLog(buildID, stream.Msg().GetLog(), true); err != nil {
			return err
		}
	}
	if err := stream.Err(); err != nil {
		return client.ContextError(ctx, client.RPCError("build log stream disconnected", err))
	}
	if ctx.Err() != nil {
		return client.ContextError(ctx, ctx.Err())
	}
	return nil
}

func (e *Executor) Monitor(ctx context.Context, apiClient client.Client, initial *api.Build, logs bool) (*api.Build, error) {
	build := initial
	for build.GetStatus() == api.BuildStatus_QUEUED {
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
		var err error
		build, err = get(ctx, apiClient, build.GetId())
		if err != nil {
			return nil, client.ContextError(ctx, err)
		}
	}
	if terminal(build.GetStatus()) {
		if logs {
			response, err := apiClient.GetBuildLog(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
			if err != nil {
				return nil, client.ContextError(ctx, client.RPCError("get build log", err))
			}
			if err := e.output.BuildLog(build.GetId(), response.Msg.GetLog(), true); err != nil {
				return nil, err
			}
		}
		return build, nil
	}
	if logs {
		if err := e.streamLog(ctx, apiClient, build.GetId()); err != nil {
			return nil, err
		}
		final, err := get(ctx, apiClient, build.GetId())
		if err != nil {
			return nil, client.ContextError(ctx, err)
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
		build, err = get(ctx, apiClient, build.GetId())
		if err != nil {
			return nil, client.ContextError(ctx, err)
		}
	}
	return build, nil
}

func Watch(ctx context.Context, apiClient client.Client, appID, commit string, excluded map[string]struct{}) (*api.Build, error) {
	for {
		response, err := apiClient.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: appID}))
		if err != nil {
			return nil, client.ContextError(ctx, client.RPCError("watch for build", err))
		}
		if found := newestMatching(response.Msg.GetBuilds(), commit, excluded); found != nil {
			return found, nil
		}
		if err := waitTick(ctx); err != nil {
			return nil, err
		}
	}
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

func IDs(builds []*api.Build) map[string]struct{} {
	ids := make(map[string]struct{}, len(builds))
	for _, build := range builds {
		ids[build.GetId()] = struct{}{}
	}
	return ids
}
