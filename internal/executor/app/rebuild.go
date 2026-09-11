package app

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	buildexecutor "github.com/traP-jp/neoshowcase-cli/internal/executor/build"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Rebuild(ctx context.Context, options model.Connection, identifier, commit string, wait, logs bool, emitLog func(buildmodel.LogResult) error) (appmodel.RebuildResult, error) {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return appmodel.RebuildResult{}, client.ContextError(ctx, err)
	}
	selectedCommit := strings.TrimSpace(commit)
	if selectedCommit == "" {
		selectedCommit = strings.TrimSpace(application.GetCommit())
	}
	if selectedCommit == "" {
		return appmodel.RebuildResult{}, fmt.Errorf("application %s has no current commit; no build was started", application.GetId())
	}
	var excluded map[string]struct{}
	if wait {
		before, err := apiClient.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()}))
		if err != nil {
			return appmodel.RebuildResult{}, client.ContextError(ctx, client.RPCError("snapshot builds before rebuild", err))
		}
		excluded = buildexecutor.IDs(before.Msg.GetBuilds())
	}
	if _, err := apiClient.RetryCommitBuild(ctx, connect.NewRequest(&api.RetryCommitBuildRequest{ApplicationId: application.GetId(), Commit: selectedCommit})); err != nil {
		return appmodel.RebuildResult{}, client.ContextError(ctx, client.RPCError("rebuild application", err))
	}
	if !wait {
		requested := appmodel.RebuildRequested{Application: toModel(application), Commit: selectedCommit}
		return appmodel.RebuildResult{Requested: &requested}, nil
	}
	build, err := buildexecutor.Watch(ctx, apiClient, application.GetId(), selectedCommit, excluded)
	if err != nil {
		return appmodel.RebuildResult{}, err
	}
	final, err := e.builds.Monitor(ctx, apiClient, build, logs, emitLog)
	if err != nil {
		return appmodel.RebuildResult{}, err
	}
	completion := buildmodel.CompletionResult{Build: buildexecutor.ToModel(final, application.GetName()), Streaming: logs}
	if err := buildexecutor.Result(final); err != nil {
		return appmodel.RebuildResult{Completion: &completion}, err
	}
	return appmodel.RebuildResult{Completion: &completion}, nil
}
