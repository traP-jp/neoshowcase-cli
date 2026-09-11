package build

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Cancel(ctx context.Context, options model.Connection, id string) (buildmodel.CancelResult, error) {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return buildmodel.CancelResult{}, err
	}
	if terminal(build.GetStatus()) {
		return buildmodel.CancelResult{Build: ToModel(build, ""), State: "already terminal; no change"}, nil
	}
	if _, err := apiClient.CancelBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()})); err != nil {
		return buildmodel.CancelResult{}, client.RPCError("cancel build", err)
	}
	return buildmodel.CancelResult{Build: ToModel(build, ""), State: "cancel requested"}, nil
}
