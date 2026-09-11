package build

import (
	"time"

	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type buildView struct {
	ID              string `json:"id"`
	ApplicationID   string `json:"application_id"`
	ApplicationName string `json:"application_name,omitempty"`
	Commit          string `json:"commit"`
	Status          string `json:"status"`
	Retriable       bool   `json:"retriable"`
	QueuedAt        string `json:"queued_at,omitempty"`
	StartedAt       string `json:"started_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	FinishedAt      string `json:"finished_at,omitempty"`
}

func utc(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func nullUTC(value *time.Time) string {
	if value == nil {
		return ""
	}
	return utc(*value)
}

func buildViewOf(build buildmodel.Build) buildView {
	return buildView{
		ID: build.ID, ApplicationID: build.ApplicationID, ApplicationName: build.ApplicationName,
		Commit: build.Commit, Status: build.Status, Retriable: build.Retriable,
		QueuedAt: utc(build.QueuedAt), StartedAt: nullUTC(build.StartedAt),
		UpdatedAt: nullUTC(build.UpdatedAt), FinishedAt: nullUTC(build.FinishedAt),
	}
}
