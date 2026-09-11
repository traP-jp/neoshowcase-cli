package build

import (
	"context"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func (e *Executor) Logs(ctx context.Context, options model.Connection, id string, follow bool, emitLog func(buildmodel.LogResult) error) error {
	apiClient := client.New(options)
	build, err := get(ctx, apiClient, id)
	if err != nil {
		return client.ContextError(ctx, err)
	}
	if terminal(build.GetStatus()) {
		response, err := apiClient.GetBuildLog(ctx, connect.NewRequest(&api.BuildIdRequest{BuildId: build.GetId()}))
		if err != nil {
			return client.ContextError(ctx, client.RPCError("get build log", err))
		}
		return emitLog(buildmodel.LogResult{BuildID: build.GetId(), Text: string(response.Msg.GetLog())})
	}
	_ = follow // In-progress logs always follow the sole server stream.
	final, err := e.Monitor(ctx, apiClient, build, true, emitLog)
	if err != nil {
		return err
	}
	return Result(final)
}
