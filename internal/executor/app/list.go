package app

import (
	"context"
	"sort"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func (e *Executor) List(ctx context.Context, options model.Connection) error {
	response, err := client.New(options).GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
	if err != nil {
		return client.RPCError("list applications", err)
	}
	applications := response.Msg.GetApplications()
	sort.Slice(applications, func(i, j int) bool { return applications[i].GetName() < applications[j].GetName() })
	result := appmodel.ListResult{Applications: make([]appmodel.Application, 0, len(applications))}
	for _, application := range applications {
		result.Applications = append(result.Applications, toModel(application))
	}
	return e.emit(result)
}
