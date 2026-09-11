package build

import (
	"context"

	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Watch(ctx context.Context, options model.Connection, application, commit string, logs bool) error {
	apiClient := client.New(options)
	resolved, err := client.ResolveApplication(ctx, apiClient, application)
	if err != nil {
		return client.ContextError(ctx, err)
	}
	build, err := Watch(ctx, apiClient, resolved.GetId(), commit, nil)
	if err != nil {
		return err
	}
	final, err := e.Monitor(ctx, apiClient, build, logs)
	if err != nil {
		return err
	}
	if err := e.output.BuildResult(final, resolved.GetName(), logs); err != nil {
		return err
	}
	return Result(final)
}
