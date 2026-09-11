package app

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Restart(ctx context.Context, options model.Connection, identifier string) error {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return err
	}
	if !application.GetRunning() {
		return fmt.Errorf("application %q is stopped; refusing to start it (state check is non-atomic)", application.GetId())
	}
	if _, err := apiClient.StartApplication(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: application.GetId()})); err != nil {
		return client.RPCError("restart application (state check is non-atomic)", err)
	}
	return e.output.Mutation(mutation("app.restart", application, "restart requested (state check was non-atomic)"))
}
