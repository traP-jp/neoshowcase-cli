package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) Logs(ctx context.Context, options model.Connection, identifier string, follow bool, tail int32, since time.Time, emit func(appmodel.LogsResult) error) error {
	apiClient := client.New(options)
	application, err := client.ResolveApplication(ctx, apiClient, identifier)
	if err != nil {
		return client.ContextError(ctx, err)
	}
	historyBefore := time.Now().UTC()
	response, err := apiClient.GetOutput(ctx, connect.NewRequest(&api.GetOutputRequest{ApplicationId: application.GetId(), Before: timestamppb.New(historyBefore), Limit: tail}))
	if err != nil {
		return client.ContextError(ctx, client.RPCError("get application logs", err))
	}
	outputs := response.Msg.GetOutputs()
	sort.SliceStable(outputs, func(i, j int) bool { return outputs[i].GetTime().AsTime().Before(outputs[j].GetTime().AsTime()) })
	filtered := make([]appmodel.Log, 0, len(outputs))
	var cursor time.Time
	for _, output := range outputs {
		at := output.GetTime().AsTime()
		if at.After(cursor) {
			cursor = at
		}
		if since.IsZero() || !at.Before(since) {
			filtered = append(filtered, appmodel.Log{ApplicationID: application.GetId(), Time: at, Text: output.GetLog()})
		}
	}
	if err := emit(appmodel.LogsResult{Logs: filtered, Streaming: follow}); err != nil {
		return err
	}
	if !follow {
		return nil
	}
	if cursor.IsZero() {
		cursor = historyBefore
	} else {
		cursor = cursor.Add(time.Nanosecond)
	}
	if !since.IsZero() && since.After(cursor) {
		cursor = since
	}
	stream, err := apiClient.GetOutputStream(ctx, connect.NewRequest(&api.GetOutputStreamRequest{ApplicationId: application.GetId(), Begin: timestamppb.New(cursor)}))
	if err != nil {
		return client.ContextError(ctx, client.RPCError("follow application logs", err))
	}
	for stream.Receive() {
		message := stream.Msg()
		entry := appmodel.Log{ApplicationID: application.GetId(), Time: message.GetTime().AsTime(), Text: message.GetLog()}
		if err := emit(appmodel.LogsResult{Logs: []appmodel.Log{entry}, Streaming: true}); err != nil {
			return err
		}
	}
	if err := stream.Err(); err != nil {
		return client.ContextError(ctx, client.RPCError("application log stream ended", err))
	}
	if ctx.Err() != nil {
		return client.ContextError(ctx, ctx.Err())
	}
	return fmt.Errorf("application log stream ended unexpectedly")
}
