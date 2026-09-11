package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	api "github.com/traP-jp/neoshowcase-cli/internal/api"
)

type applicationView struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Commit            string `json:"commit,omitempty"`
	Running           bool   `json:"running"`
	ContainerState    string `json:"container_state"`
	LatestBuildStatus string `json:"latest_build_status,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
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

type logView struct {
	ApplicationID string `json:"application_id,omitempty"`
	BuildID       string `json:"build_id,omitempty"`
	Time          string `json:"time,omitempty"`
	Log           string `json:"log"`
}

func utc(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

func nullUTC(ts *api.NullTimestamp) string {
	if ts == nil || !ts.GetValid() {
		return ""
	}
	return utc(ts.GetTimestamp())
}

func viewApplication(app *api.Application) applicationView {
	latest := ""
	if app.LatestBuildStatus != nil {
		latest = app.GetLatestBuildStatus().String()
	}
	return applicationView{
		ID: app.GetId(), Name: app.GetName(), Commit: app.GetCommit(), Running: app.GetRunning(),
		ContainerState: app.GetContainer().String(), LatestBuildStatus: latest,
		CreatedAt: utc(app.GetCreatedAt()), UpdatedAt: utc(app.GetUpdatedAt()),
	}
}

func viewBuild(build *api.Build, appName string) buildView {
	return buildView{
		ID: build.GetId(), ApplicationID: build.GetApplicationId(), ApplicationName: appName,
		Commit: build.GetCommit(), Status: build.GetStatus().String(), Retriable: build.GetRetriable(),
		QueuedAt: utc(build.GetQueuedAt()), StartedAt: nullUTC(build.GetStartedAt()),
		UpdatedAt: nullUTC(build.GetUpdatedAt()), FinishedAt: nullUTC(build.GetFinishedAt()),
	}
}

func writeValue(out io.Writer, format string, value any, text string) error {
	switch format {
	case "text":
		_, err := fmt.Fprintln(out, text)
		return err
	case "json":
		encoder := json.NewEncoder(out)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case "jsonl":
		return writeJSONLine(out, value)
	default:
		return fail(ExitUsage, "unsupported output format %q", format)
	}
}

func writeJSONLine(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func writeLog(out io.Writer, format string, entry logView, streaming bool) error {
	if format == "text" {
		_, err := io.WriteString(out, entry.Log)
		if err == nil && entry.Log != "" && !strings.HasSuffix(entry.Log, "\n") {
			_, err = io.WriteString(out, "\n")
		}
		return err
	}
	if streaming || format == "jsonl" {
		return writeJSONLine(out, entry)
	}
	return nil
}
