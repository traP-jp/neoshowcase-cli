package build

import (
	"fmt"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type GetCommand struct {
	BuildID string `arg:"" name:"build-id" help:"Build ID"`
}

func (command *GetCommand) Validate(ValidationContext) (model.Command, error) {
	return buildmodel.GetCommand{BuildID: command.BuildID}, nil
}

func RenderGet(renderer *clioutput.Renderer, result buildmodel.GetResult) error {
	view := buildViewOf(result.Build)
	text := fmt.Sprintf("ID: %s\nApplication ID: %s\nCommit: %s\nStatus: %s\nRetriable: %t", view.ID, view.ApplicationID, view.Commit, view.Status, view.Retriable)
	return renderer.WriteValue(view, text)
}
