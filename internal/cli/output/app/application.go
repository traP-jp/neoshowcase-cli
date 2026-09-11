package app

import (
	"fmt"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
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

func renderState(renderer *core.Writer, operation string, application appmodel.Application, state string) error {
	value := map[string]any{"operation": operation, "application_id": application.ID, "application_name": application.Name, "commit": application.Commit, "state": state}
	return renderer.WriteValue(value, fmt.Sprintf("%s: %s (%s): %s", operation, application.Name, application.ID, state))
}
