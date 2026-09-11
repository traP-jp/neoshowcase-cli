package app

import (
	"fmt"
	"time"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type Command struct {
	List    ListCommand    `cmd:"" help:"List applications"`
	Get     GetCommand     `cmd:"" help:"Get one application by ID or exact name"`
	Logs    LogsCommand    `cmd:"" help:"Print application logs"`
	Start   StartCommand   `cmd:"" help:"Start a stopped application"`
	Stop    StopCommand    `cmd:"" help:"Stop a running application"`
	Restart RestartCommand `cmd:"" help:"Restart a running application"`
	Rebuild RebuildCommand `cmd:"" help:"Rebuild an application commit"`
}

func (Command) Help() string {
	return `An <application> argument accepts either an application ID or a unique exact application name. A missing or ambiguous name fails without making changes.

Restart performs a non-atomic state check followed by StartApplication, so a concurrent server-side state change can race the check.`
}

type ValidationContext struct {
	AllowMutable bool
	Now          time.Time
}

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

func utc(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func applicationViewOf(application appmodel.Application) applicationView {
	return applicationView{
		ID: application.ID, Name: application.Name, Commit: application.Commit, Running: application.Running,
		ContainerState: application.ContainerState, LatestBuildStatus: application.LatestBuildStatus,
		CreatedAt: utc(application.CreatedAt), UpdatedAt: utc(application.UpdatedAt),
	}
}

func renderState(renderer *clioutput.Renderer, operation string, application appmodel.Application, state string) error {
	value := map[string]any{
		"operation":        operation,
		"application_id":   application.ID,
		"application_name": application.Name,
		"commit":           application.Commit,
		"state":            state,
	}
	text := fmt.Sprintf("%s: %s (%s): %s", operation, application.Name, application.ID, state)
	return renderer.WriteValue(value, text)
}
