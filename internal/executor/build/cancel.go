package build

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/cli"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) Cancel(ctx context.Context, options model.Connection, id string) error {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return err
	}
	result := cli.Mutation{Operation: "build.cancel", BuildID: build.GetId(), ApplicationID: build.GetApplicationId(), Commit: build.GetCommit(), Status: build.GetStatus().String()}
	if terminal(build.GetStatus()) {
		result.State = "already terminal; no change"
		return e.output.Mutation(result)
	}
	if _, err := apiClient.CancelBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()})); err != nil {
		return client.RPCError("cancel build", err)
	}
	result.State = "cancel requested"
	return e.output.Mutation(result)
}
