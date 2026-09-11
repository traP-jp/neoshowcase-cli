package build

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type CancelCommand struct {
	BuildID string `arg:"" name:"build-id" help:"Build ID"`
}

func (command *CancelCommand) Validate(context ValidationContext) (model.Command, error) {
	if err := validation.Mutable(context.AllowMutable); err != nil {
		return nil, err
	}
	return buildmodel.CancelCommand{BuildID: command.BuildID}, nil
}
