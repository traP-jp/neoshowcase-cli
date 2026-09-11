package build

import (
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type WaitCommand struct {
	BuildID string        `arg:"" name:"build-id" help:"Build ID"`
	Logs    bool          `help:"Print build logs while waiting"`
	Timeout time.Duration `default:"10m" help:"Whole-command timeout"`
}

func (command *WaitCommand) Validate(ValidationContext) (model.Command, error) {
	if err := validation.Timeout(command.Timeout); err != nil {
		return nil, err
	}
	return buildmodel.WaitCommand{BuildID: command.BuildID, Logs: command.Logs, Timeout: command.Timeout}, nil
}
