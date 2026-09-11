package app

import (
	"fmt"
	"strings"
	"time"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

type RebuildCommand struct {
	Application string        `arg:"" help:"Application ID or exact name"`
	Commit      string        `help:"Commit SHA (defaults to the application's current commit)"`
	Wait        bool          `help:"Wait for the new build to finish"`
	Logs        bool          `help:"Print logs while waiting"`
	Timeout     time.Duration `default:"10m" help:"Whole-command timeout when waiting"`
}

func (command *RebuildCommand) Validate(context ValidationContext) (model.Command, error) {
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
	return appmodel.RebuildCommand{Application: command.Application, Commit: strings.TrimSpace(command.Commit), Wait: wait, Logs: command.Logs, Timeout: timeout}, nil
}

func RenderRebuild(renderer *clioutput.Renderer, result appmodel.RebuildRequested) error {
	value := map[string]any{
		"operation":        "app.rebuild",
		"application_id":   result.Application.ID,
		"application_name": result.Application.Name,
		"commit":           result.Commit,
		"state":            "requested",
	}
	text := fmt.Sprintf("rebuild requested: %s (%s) commit %s", result.Application.Name, result.Application.ID, result.Commit)
	return renderer.WriteValue(value, text)
}
