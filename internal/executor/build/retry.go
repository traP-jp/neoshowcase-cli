package build

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Retry(ctx context.Context, options model.Connection, id string, wait, logs bool, emitLog func(buildmodel.LogResult) error) (buildmodel.RetryResult, error) {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return buildmodel.RetryResult{}, client.ContextError(ctx, err)
	}
	if !build.GetRetriable() {
		return buildmodel.RetryResult{}, fmt.Errorf("build %s is not retriable; no build was started", build.GetId())
	}
	var excluded map[string]struct{}
	if wait {
		before, err := apiClient.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: build.GetApplicationId()}))
		if err != nil {
			return buildmodel.RetryResult{}, client.ContextError(ctx, client.RPCError("snapshot builds before retry", err))
		}
		excluded = IDs(before.Msg.GetBuilds())
	}
	if _, err := apiClient.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: build.GetApplicationId(), Commit: build.GetCommit()})); err != nil {
		return buildmodel.RetryResult{}, client.ContextError(ctx, client.RPCError("retry build", err))
	}
	if !wait {
		requested := buildmodel.RetryRequested{Build: ToModel(build, "")}
		return buildmodel.RetryResult{Requested: &requested}, nil
	}
	newBuild, err := Watch(ctx, apiClient, build.GetApplicationId(), build.GetCommit(), excluded)
	if err != nil {
		return buildmodel.RetryResult{}, err
	}
	final, err := e.Monitor(ctx, apiClient, newBuild, logs, emitLog)
	if err != nil {
		return buildmodel.RetryResult{}, err
	}
	completion := buildmodel.CompletionResult{Build: ToModel(final, ""), Streaming: logs}
	if err := Result(final); err != nil {
		return buildmodel.RetryResult{Completion: &completion}, err
	}
	return buildmodel.RetryResult{Completion: &completion}, nil
}
