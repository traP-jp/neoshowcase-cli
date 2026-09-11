package app

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Stop(ctx context.Context, options model.Connection, identifier string) error {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return err
	}
	if !application.GetRunning() {
		return e.emit(appmodel.StopResult{Application: toModel(application), State: "already stopped; no change"})
	}
	if _, err := apiClient.StopApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return client.RPCError("stop application", err)
	}
	return e.emit(appmodel.StopResult{Application: toModel(application), State: "stopped"})
}
