package build

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Retry(ctx context.Context, options model.Connection, id string, wait, logs bool) error {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return client.ContextError(ctx, err)
	}
	if !build.GetRetriable() {
		return fmt.Errorf("build %s is not retriable; no build was started", build.GetId())
	}
	var excluded map[string]struct{}
	if wait {
		before, err := apiClient.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: build.GetApplicationId()}))
		if err != nil {
			return client.ContextError(ctx, client.RPCError("snapshot builds before retry", err))
		}
		excluded = IDs(before.Msg.GetBuilds())
	}
	if _, err := apiClient.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: build.GetApplicationId(), Commit: build.GetCommit()})); err != nil {
		return client.ContextError(ctx, client.RPCError("retry build", err))
	}
	if !wait {
		return e.output.Mutation(cli.Mutation{Operation: "build.retry", SourceBuildID: build.GetId(), ApplicationID: build.GetApplicationId(), Commit: build.GetCommit(), State: "requested"})
	}
	newBuild, err := Watch(ctx, apiClient, build.GetApplicationId(), build.GetCommit(), excluded)
	if err != nil {
		return err
	}
	final, err := e.Monitor(ctx, apiClient, newBuild, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, "", logs); err != nil {
		return err
	}
	return Result(final)
}
