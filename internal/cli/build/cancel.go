package build

import (
	"fmt"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
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

func RenderCancel(renderer *clioutput.Renderer, result buildmodel.CancelResult) error {
	value := map[string]any{
		"operation":      "build.cancel",
		"build_id":       result.Build.ID,
		"application_id": result.Build.ApplicationID,
		"commit":         result.Build.Commit,
		"status":         result.Build.Status,
		"state":          result.State,
	}
	text := fmt.Sprintf("cancel requested: build %s (%s)", result.Build.ID, result.Build.Status)
	if result.State == "already terminal; no change" {
		text = fmt.Sprintf("build %s is already %s; no change", result.Build.ID, result.Build.Status)
	}
	return renderer.WriteValue(value, text)
}
