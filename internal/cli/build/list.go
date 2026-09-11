package build

import (
	"fmt"

	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type ListCommand struct {
	Application string `arg:"" optional:"" help:"Application ID or exact name"`
	Page        int32  `default:"0" help:"Zero-indexed page"`
	Limit       int32  `default:"20" help:"Number of builds in the page"`
}

func (command *ListCommand) Validate(ValidationContext) (model.Command, error) {
	if command.Page < 0 || command.Limit <= 0 {
		return nil, fmt.Errorf("--page must be non-negative and --limit must be positive")
	}
	return buildmodel.ListCommand{Application: command.Application, Page: command.Page, Limit: command.Limit}, nil
}
