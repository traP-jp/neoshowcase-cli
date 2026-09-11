package app

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Rebuild(ctx context.Context, options model.Connection, identifier, commit string, wait, logs bool) error {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return client.ContextError(ctx, err)
	}
	selectedCommit := strings.TrimSpace(commit)
	if selectedCommit == "" {
		selectedCommit = strings.TrimSpace(application.GetCommit())
	}
	if selectedCommit == "" {
		return fmt.Errorf("application %s has no current commit; no build was started", application.GetId())
	}
	var excluded map[string]struct{}
	if wait {
		before, err := apiClient.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()}))
		if err != nil {
			return client.ContextError(ctx, client.RPCError("snapshot builds before rebuild", err))
		}
		excluded = buildexecutor.IDs(before.Msg.GetBuilds())
	}
	if _, err := apiClient.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: application.GetId(), Commit: selectedCommit})); err != nil {
		return client.ContextError(ctx, client.RPCError("rebuild application", err))
	}
	if !wait {
		return e.output.Mutation(cli.Mutation{Operation: "app.rebuild", ApplicationID: application.GetId(), ApplicationName: application.GetName(), Commit: selectedCommit, State: "requested"})
	}
	build, err := buildexecutor.Watch(ctx, apiClient, application.GetId(), selectedCommit, excluded)
	if err != nil {
		return err
	}
	final, err := e.builds.Monitor(ctx, apiClient, build, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, application.GetName(), logs); err != nil {
		return err
	}
	return buildexecutor.Result(final)
}
