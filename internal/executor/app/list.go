package app

import (
	"context"
	"sort"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) List(ctx context.Context, options model.Connection) error {
	response, err := client.New(options).GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
	if err != nil {
		return client.RPCError("list applications", err)
	}
	applications := response.Msg.GetApplications()
	sort.Slice(applications, func(i, j int) bool { return applications[i].GetName() < applications[j].GetName() })
	return e.output.Applications(applications)
}
