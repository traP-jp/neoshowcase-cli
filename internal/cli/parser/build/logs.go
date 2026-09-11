package build

import (
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type LogsCommand struct {
	BuildID string        `arg:"" name:"build-id" help:"Build ID"`
	Follow  bool          `short:"f" help:"Wait for and stream an in-progress build"`
	Timeout time.Duration `default:"10m" help:"Whole-command timeout"`
}

func (command *LogsCommand) Validate(ValidationContext) (model.Command, error) {
	if err := validation.Timeout(command.Timeout); err != nil {
		return nil, err
	}
	return buildmodel.LogsCommand{BuildID: command.BuildID, Follow: command.Follow, Timeout: command.Timeout}, nil
}
