package build

import (
	"fmt"
	"time"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
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

func RenderCompletion(renderer *clioutput.Renderer, result buildmodel.CompletionResult) error {
	view := buildViewOf(result.Build)
	if result.Streaming && renderer.Format() == "text" {
		return renderer.Writef("build %s: %s\n", result.Build.ID, result.Build.Status)
	}
	if result.Streaming {
		return renderer.WriteJSONLine(view)
	}
	return renderer.WriteValue(view, fmt.Sprintf("build %s: %s", result.Build.ID, result.Build.Status))
}
