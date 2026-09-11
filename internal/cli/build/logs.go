package build

import (
	"time"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
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

func RenderLogs(renderer *clioutput.Renderer, result buildmodel.LogResult) error {
	if renderer.Format() == "text" {
		return renderer.WriteString(result.Text)
	}
	entry := clioutput.Log{BuildID: result.BuildID, Text: result.Text}
	if !result.Streaming && renderer.Format() != "jsonl" {
		return renderer.WriteValue(entry, "")
	}
	return renderer.WriteLog(entry, result.Streaming)
}
