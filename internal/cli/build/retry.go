package build

import (
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type RetryCommand struct {
	BuildID string        `arg:"" name:"build-id" help:"Build ID"`
	Wait    bool          `help:"Wait for the new build to finish"`
	Logs    bool          `help:"Print logs while waiting"`
	Timeout time.Duration `default:"10m" help:"Whole-command timeout when waiting"`
}

func (command *RetryCommand) Validate(context ValidationContext) (model.Command, error) {
	wait := command.Wait || command.Logs
	if wait {
		if err := validation.Timeout(command.Timeout); err != nil {
			return nil, err
		}
	}
	if err := validation.Mutable(context.AllowMutable); err != nil {
		return nil, err
	}
	timeout := time.Duration(0)
	if wait {
		timeout = command.Timeout
	}
	return buildmodel.RetryCommand{BuildID: command.BuildID, Wait: wait, Logs: command.Logs, Timeout: timeout}, nil
}
