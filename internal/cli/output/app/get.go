package app

import (
	"fmt"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func RenderGet(renderer *core.Writer, result appmodel.GetResult) error {
	view := applicationViewOf(result.Application)
	text := fmt.Sprintf("ID: %s\nName: %s\nCommit: %s\nRunning: %t\nContainer state: %s\nLatest build: %s", view.ID, view.Name, view.Commit, view.Running, view.ContainerState, view.LatestBuildStatus)
	return renderer.WriteValue(view, text)
}
