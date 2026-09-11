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

func (rt *runtime) Applications(apps []*api.Application) error {
	views := make([]applicationView, 0, len(apps))
	var text strings.Builder
	fmt.Fprintln(&text, "ID\tNAME\tCOMMIT\tSTATE")
	for _, app := range apps {
		views = append(views, viewApplication(app))
		fmt.Fprintf(&text, "%s\t%s\t%s\t%s\n", app.GetId(), app.GetName(), app.GetCommit(), app.GetContainer())
	}
	if rt.output == "jsonl" {
		for _, view := range views {
			if err := writeJSONLine(rt.out, view); err != nil {
				return err
			}
		}
		return nil
	}
	return writeValue(rt.out, rt.output, views, strings.TrimSuffix(text.String(), "\n"))
}

func (rt *runtime) Application(app *api.Application) error {
	view := viewApplication(app)
	text := fmt.Sprintf("ID: %s\nName: %s\nCommit: %s\nRunning: %t\nContainer state: %s\nLatest build: %s", view.ID, view.Name, view.Commit, view.Running, view.ContainerState, view.LatestBuildStatus)
	return writeValue(rt.out, rt.output, view, text)
}

func (rt *runtime) ApplicationLogs(applicationID string, outputs []*api.ApplicationOutput, streaming bool) error {
	entries := make([]logView, 0, len(outputs))
	for _, output := range outputs {
		entry := logView{ApplicationID: applicationID, Time: utc(output.GetTime()), Log: output.GetLog()}
		entries = append(entries, entry)
		if streaming {
			if err := writeLog(rt.out, rt.output, entry, true); err != nil {
				return err
			}
		}
	}
	if streaming {
		return nil
	}
	if rt.output == "text" {
		for _, entry := range entries {
			if err := writeLog(rt.out, rt.output, entry, false); err != nil {
				return err
			}
		}
		return nil
	}
	if rt.output == "jsonl" {
		for _, entry := range entries {
			if err := writeJSONLine(rt.out, entry); err != nil {
				return err
			}
		}
		return nil
	}
	return writeValue(rt.out, rt.output, entries, "")
}

func (rt *runtime) Builds(builds []*api.Build, names map[string]string) error {
	views := make([]buildView, 0, len(builds))
	var text strings.Builder
	fmt.Fprintln(&text, "ID\tAPPLICATION\tCOMMIT\tSTATUS")
	for _, build := range builds {
		applicationName := names[build.GetApplicationId()]
		views = append(views, viewBuild(build, applicationName))
		if applicationName == "" {
			applicationName = build.GetApplicationId()
		}
		fmt.Fprintf(&text, "%s\t%s\t%s\t%s\n", build.GetId(), applicationName, build.GetCommit(), build.GetStatus())
	}
	if rt.output == "jsonl" {
		for _, view := range views {
			if err := writeJSONLine(rt.out, view); err != nil {
				return err
			}
		}
		return nil
	}
	return writeValue(rt.out, rt.output, views, strings.TrimSuffix(text.String(), "\n"))
}

func (rt *runtime) BuildDetails(build *api.Build) error {
	view := viewBuild(build, "")
	text := fmt.Sprintf("ID: %s\nApplication ID: %s\nCommit: %s\nStatus: %s\nRetriable: %t", view.ID, view.ApplicationID, view.Commit, view.Status, view.Retriable)
	return writeValue(rt.out, rt.output, view, text)
}

func (rt *runtime) BuildResult(build *api.Build, applicationName string, streaming bool) error {
	view := viewBuild(build, applicationName)
	if streaming && rt.output == "text" {
		_, err := fmt.Fprintf(rt.errOut, "build %s: %s\n", build.GetId(), build.GetStatus())
		return err
	}
	if streaming {
		return writeJSONLine(rt.out, view)
	}
	return writeValue(rt.out, rt.output, view, fmt.Sprintf("build %s: %s", build.GetId(), build.GetStatus()))
}

func (rt *runtime) BuildLog(buildID string, data []byte, streaming bool) error {
	if rt.output == "text" {
		_, err := rt.out.Write(data)
		return err
	}
	entry := logView{BuildID: buildID, Log: string(data)}
	if !streaming && rt.output != "jsonl" {
		return writeValue(rt.out, rt.output, entry, "")
	}
	return writeLog(rt.out, rt.output, entry, streaming)
}

func (rt *runtime) Mutation(result MutationResult) error {
	var value map[string]any
	var text string
	switch result.Operation {
	case "app.rebuild":
		value = map[string]any{"operation": result.Operation, "application_id": result.ApplicationID, "application_name": result.ApplicationName, "commit": result.Commit, "state": result.State}
		text = fmt.Sprintf("rebuild requested: %s (%s) commit %s", result.ApplicationName, result.ApplicationID, result.Commit)
	case "build.retry":
		value = map[string]any{"operation": result.Operation, "source_build_id": result.SourceBuildID, "application_id": result.ApplicationID, "commit": result.Commit, "state": result.State}
		text = fmt.Sprintf("retry requested: build %s commit %s", result.SourceBuildID, result.Commit)
	case "build.cancel":
		value = map[string]any{"operation": result.Operation, "build_id": result.BuildID, "application_id": result.ApplicationID, "commit": result.Commit, "status": result.Status, "state": result.State}
		if result.State == "already terminal; no change" {
			text = fmt.Sprintf("build %s is already %s; no change", result.BuildID, result.Status)
		} else {
			text = fmt.Sprintf("cancel requested: build %s (%s)", result.BuildID, result.Status)
		}
	default:
		value = map[string]any{"operation": result.Operation, "application_id": result.ApplicationID, "application_name": result.ApplicationName, "commit": result.Commit, "state": result.State}
		text = fmt.Sprintf("%s: %s (%s): %s", result.Operation, result.ApplicationName, result.ApplicationID, result.State)
	}
	return writeValue(rt.out, rt.output, value, text)
}
