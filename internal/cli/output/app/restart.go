package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func RenderRestart(renderer *core.Writer, result appmodel.RestartResult) error {
	return renderState(renderer, "app.restart", result.Application, result.State)
}
