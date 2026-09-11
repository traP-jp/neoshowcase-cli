package build

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Wait(ctx context.Context, options model.Connection, id string, logs bool) error {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return client.ContextError(ctx, err)
	}
	final, err := e.Monitor(ctx, apiClient, build, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, "", logs); err != nil {
		return err
	}
	return Result(final)
}
