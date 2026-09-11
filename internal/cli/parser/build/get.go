package build

import (
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type GetCommand struct {
	BuildID string `arg:"" name:"build-id" help:"Build ID"`
}

func (command *GetCommand) Validate(ValidationContext) (model.Command, error) {
	return buildmodel.GetCommand{BuildID: command.BuildID}, nil
}
