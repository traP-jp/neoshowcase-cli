package app

import (
	"fmt"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type GetCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

func (command *GetCommand) Validate(ValidationContext) (model.Command, error) {
	return appmodel.GetCommand{Application: command.Application}, nil
}

func RenderGet(renderer *clioutput.Renderer, result appmodel.GetResult) error {
	view := applicationViewOf(result.Application)
	text := fmt.Sprintf("ID: %s\nName: %s\nCommit: %s\nRunning: %t\nContainer state: %s\nLatest build: %s", view.ID, view.Name, view.Commit, view.Running, view.ContainerState, view.LatestBuildStatus)
	return renderer.WriteValue(view, text)
}
