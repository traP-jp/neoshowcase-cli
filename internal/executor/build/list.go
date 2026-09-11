package build

import (
	"context"
	"sort"

	"connectrpc.com/connect"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	"github.com/traP-jp/neoshowcase-cli/internal/executor/client"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
)

func (e *Executor) List(ctx context.Context, options model.Connection, application string, page, limit int32) error {
	apiClient := client.New(options)
	var builds []*api.Build
	names := map[string]string{}
	if application != "" {
		resolved, err := client.ResolveApplication(ctx, apiClient, application)
		if err != nil {
			return err
		}
		names[resolved.GetId()] = resolved.GetName()
		response, err := apiClient.GetBuilds(ctx, connect.NewRequest(&api.ApplicationIdRequest{Id: resolved.GetId()}))
		if err != nil {
			return client.RPCError("list application builds", err)
		}
		builds = response.Msg.GetBuilds()
	} else {
		response, err := apiClient.GetAllBuilds(ctx, connect.NewRequest(&api.GetAllBuildsRequest{Page: page, Limit: limit}))
		if err != nil {
			return client.RPCError("list builds", err)
		}
		builds = response.Msg.GetBuilds()
		applications, err := apiClient.GetApplications(ctx, connect.NewRequest(&api.GetApplicationsRequest{Scope: api.GetApplicationsRequest_ALL}))
		if err != nil {
			return client.RPCError("list applications for build names", err)
		}
		for _, application := range applications.Msg.GetApplications() {
			names[application.GetId()] = application.GetName()
		}
	}
	sort.SliceStable(builds, func(i, j int) bool {
		return builds[i].GetQueuedAt().AsTime().After(builds[j].GetQueuedAt().AsTime())
	})
	if application != "" {
		start := int64(page) * int64(limit)
		if start >= int64(len(builds)) {
			builds = nil
		} else {
			end := start + int64(limit)
			if end > int64(len(builds)) {
				end = int64(len(builds))
			}
			builds = builds[int(start):int(end)]
		}
	}
	return e.output.Builds(builds, names)
}
