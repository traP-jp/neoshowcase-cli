package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type StartCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

func (command *StartCommand) Validate(context ValidationContext) (model.Command, error) {
	if err := validation.Mutable(context.AllowMutable); err != nil {
		return nil, err
	}
	return appmodel.StartCommand{Application: command.Application}, nil
}
