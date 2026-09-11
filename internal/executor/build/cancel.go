package build

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Cancel(ctx context.Context, options model.Connection, id string) error {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return err
	}
	if terminal(build.GetStatus()) {
		return e.emit(buildmodel.CancelResult{Build: ToModel(build, ""), State: "already terminal; no change"})
	}
	if _, err := apiClient.CancelBuild(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()})); err != nil {
		return client.RPCError("cancel build", err)
	}
	return e.emit(buildmodel.CancelResult{Build: ToModel(build, ""), State: "cancel requested"})
}
