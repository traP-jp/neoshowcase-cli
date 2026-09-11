package app

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Start(ctx context.Context, options model.Connection, identifier string) (appmodel.StartResult, error) {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return appmodel.StartResult{}, err
	}
	if application.GetRunning() {
		return appmodel.StartResult{Application: toModel(application), State: "already running; no change"}, nil
	}
	if _, err := apiClient.StartApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return appmodel.StartResult{}, client.RPCError("start application", err)
	}
	return appmodel.StartResult{Application: toModel(application), State: "started"}, nil
}
