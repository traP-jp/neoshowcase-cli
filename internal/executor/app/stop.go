package app

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Stop(ctx context.Context, options model.Connection, identifier string) error {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return err
	}
	if !application.GetRunning() {
		return e.output.Mutation(mutation("app.stop", application, "already stopped; no change"))
	}
	if _, err := apiClient.StopApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return client.RPCError("stop application", err)
	}
	return e.output.Mutation(mutation("app.stop", application, "stopped"))
}
