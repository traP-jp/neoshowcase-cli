package build

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Wait(ctx context.Context, options model.Connection, id string, logs bool, emitLog func(buildmodel.LogResult) error) (buildmodel.CompletionResult, error) {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return buildmodel.CompletionResult{}, client.ContextError(ctx, err)
	}
	final, err := e.Monitor(ctx, apiClient, build, logs, emitLog)
	if err != nil {
		return buildmodel.CompletionResult{}, err
	}
	result := buildmodel.CompletionResult{Build: ToModel(final, ""), Streaming: logs}
	return result, Result(final)
}
