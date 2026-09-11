package app

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func RenderStart(renderer *core.Writer, result appmodel.StartResult) error {
	return renderState(renderer, "app.start", result.Application, result.State)
}
