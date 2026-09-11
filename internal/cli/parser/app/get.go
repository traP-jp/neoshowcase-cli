package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type GetCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

func (command *GetCommand) Validate(ValidationContext) (model.Command, error) {
	return appmodel.GetCommand{Application: command.Application}, nil
}
