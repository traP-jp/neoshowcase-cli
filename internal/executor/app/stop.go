package app

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Stop(ctx context.Context, options model.Connection, identifier string) (appmodel.StopResult, error) {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return appmodel.StopResult{}, err
	}
	if !application.GetRunning() {
		return appmodel.StopResult{Application: toModel(application), State: "already stopped; no change"}, nil
	}
	if _, err := apiClient.StopApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return appmodel.StopResult{}, client.RPCError("stop application", err)
	}
	return appmodel.StopResult{Application: toModel(application), State: "stopped"}, nil
}
