package build

import (
	"time"

	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type Command struct {
	List   ListCommand   `cmd:"" help:"List one page of builds"`
	Get    GetCommand    `cmd:"" help:"Get one build"`
	Logs   LogsCommand   `cmd:"" help:"Print build logs"`
	Wait   WaitCommand   `cmd:"" help:"Wait for a build to finish"`
	Watch  WatchCommand  `cmd:"" help:"Watch for a commit's build and wait for it"`
	Retry  RetryCommand  `cmd:"" help:"Retry a retriable build"`
	Cancel CancelCommand `cmd:"" help:"Cancel an in-progress build"`
}

func (Command) Help() string {
	return `Waiting and streaming operations default to a whole-command timeout of 10 minutes. Build detection and state polling run immediately, then every 11 seconds. On rebuild and retry, --logs implies --wait.

A monitored build exits successfully only when its terminal status is SUCCEEDED. FAILED, CANCELLED, and SKIPPED builds return a non-zero status.`
}

type ValidationContext struct {
	AllowMutable bool
}

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
