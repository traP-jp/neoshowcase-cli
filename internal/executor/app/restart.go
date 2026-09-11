package app

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Restart(ctx context.Context, options model.Connection, identifier string) (appmodel.RestartResult, error) {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return appmodel.RestartResult{}, err
	}
	if !application.GetRunning() {
		return appmodel.RestartResult{}, fmt.Errorf("application %q is stopped; refusing to start it (state check is non-atomic)", application.GetId())
	}
	if _, err := apiClient.StartApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return appmodel.RestartResult{}, client.RPCError("restart application (state check is non-atomic)", err)
	}
	return appmodel.RestartResult{Application: toModel(application), State: "restart requested (state check was non-atomic)"}, nil
}
