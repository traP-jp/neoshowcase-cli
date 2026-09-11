package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type StopCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

func (command *StopCommand) Validate(context ValidationContext) (model.Command, error) {
	if err := validation.Mutable(context.AllowMutable); err != nil {
		return nil, err
	}
	return appmodel.StopCommand{Application: command.Application}, nil
}
