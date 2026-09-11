package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func RenderStop(renderer *core.Writer, result appmodel.StopResult) error {
	return renderState(renderer, "app.stop", result.Application, result.State)
}
