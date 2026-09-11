package app

import (
	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type RestartCommand struct {
	Application string `arg:"" help:"Application ID or exact name"`
}

func (command *RestartCommand) Validate(context ValidationContext) (model.Command, error) {
	if err := validation.Mutable(context.AllowMutable); err != nil {
		return nil, err
	}
	return appmodel.RestartCommand{Application: command.Application}, nil
}

func RenderRestart(renderer *clioutput.Renderer, result appmodel.RestartResult) error {
	return renderState(renderer, "app.restart", result.Application, result.State)
}
