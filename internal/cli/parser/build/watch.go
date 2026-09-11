package build

import (
	"fmt"
	"strings"
	"time"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/parser/validation"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type WatchCommand struct {
	Application string        `arg:"" help:"Application ID or exact name"`
	Commit      string        `required:"" help:"Commit SHA to watch"`
	Logs        bool          `help:"Print build logs while waiting"`
	Timeout     time.Duration `default:"10m" help:"Whole-command timeout"`
}

func (command *WatchCommand) Validate(ValidationContext) (model.Command, error) {
	commit := strings.TrimSpace(command.Commit)
	if commit == "" {
		return nil, fmt.Errorf("--commit is required")
	}
	if err := validation.Timeout(command.Timeout); err != nil {
		return nil, err
	}
	return buildmodel.WatchCommand{Application: command.Application, Commit: commit, Logs: command.Logs, Timeout: command.Timeout}, nil
}
