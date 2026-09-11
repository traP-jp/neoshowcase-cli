package build

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Watch(ctx context.Context, options model.Connection, application, commit string, logs bool, emitLog func(buildmodel.LogResult) error) (buildmodel.CompletionResult, error) {
	apiClient := client.New(options)
	resolved, err := client.ResolveApplication(ctx, apiClient, application)
	if err != nil {
		return buildmodel.CompletionResult{}, client.ContextError(ctx, err)
	}
	build, err := Watch(ctx, apiClient, resolved.GetId(), commit, nil)
	if err != nil {
		return buildmodel.CompletionResult{}, err
	}
	final, err := e.Monitor(ctx, apiClient, build, logs, emitLog)
	if err != nil {
		return buildmodel.CompletionResult{}, err
	}
	result := buildmodel.CompletionResult{Build: ToModel(final, resolved.GetName()), Streaming: logs}
	return result, Result(final)
}
