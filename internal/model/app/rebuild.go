package app

import (
	"time"

	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type RebuildCommand struct {
	Application string
	Commit      string
	Wait        bool
	Logs        bool
	Timeout     time.Duration
}

func (RebuildCommand) IsCommand() {}

type RebuildRequested struct {
	Application Application
	Commit      string
}

type RebuildResult struct {
	Requested  *RebuildRequested
	Completion *buildmodel.CompletionResult
}
