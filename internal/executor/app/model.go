package app

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func toModel(application *api.Application) appmodel.Application {
	latestBuildStatus := ""
	if application.LatestBuildStatus != nil {
		latestBuildStatus = application.GetLatestBuildStatus().String()
	}
	return appmodel.Application{
		ID:                application.GetId(),
		Name:              application.GetName(),
		Commit:            application.GetCommit(),
		Running:           application.GetRunning(),
		ContainerState:    application.GetContainer().String(),
		LatestBuildStatus: latestBuildStatus,
		CreatedAt:         timestamp(application.GetCreatedAt()),
		UpdatedAt:         timestamp(application.GetUpdatedAt()),
	}
}

func timestamp(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime()
}
