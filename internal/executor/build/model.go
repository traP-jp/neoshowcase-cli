package build

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func ToModel(build *api.Build, applicationName string) buildmodel.Build {
	return buildmodel.Build{
		ID:              build.GetId(),
		ApplicationID:   build.GetApplicationId(),
		ApplicationName: applicationName,
		Commit:          build.GetCommit(),
		Status:          build.GetStatus().String(),
		Retriable:       build.GetRetriable(),
		QueuedAt:        timestamp(build.GetQueuedAt()),
		StartedAt:       nullTime(build.GetStartedAt()),
		UpdatedAt:       nullTime(build.GetUpdatedAt()),
		FinishedAt:      nullTime(build.GetFinishedAt()),
	}
}

func timestamp(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime()
}

func nullTime(timestamp *api.NullTimestamp) *time.Time {
	if timestamp == nil || !timestamp.GetValid() {
		return nil
	}
	value := timestamp.GetTimestamp().AsTime()
	return &value
}
