package app

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Start(ctx context.Context, options model.Connection, identifier string) error {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return err
	}
	if application.GetRunning() {
		return e.output.Mutation(mutation("app.start", application, "already running; no change"))
	}
	if _, err := apiClient.StartApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return client.RPCError("start application", err)
	}
	return e.output.Mutation(mutation("app.start", application, "started"))
}
