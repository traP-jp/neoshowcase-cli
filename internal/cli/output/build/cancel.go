package build

import (
	"fmt"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderCancel(renderer *core.Writer, result buildmodel.CancelResult) error {
	value := map[string]any{"operation": "build.cancel", "build_id": result.Build.ID, "application_id": result.Build.ApplicationID, "commit": result.Build.Commit, "status": result.Build.Status, "state": result.State}
	text := fmt.Sprintf("cancel requested: build %s (%s)", result.Build.ID, result.Build.Status)
	if result.State == "already terminal; no change" {
		text = fmt.Sprintf("build %s is already %s; no change", result.Build.ID, result.Build.Status)
	}
	return renderer.WriteValue(value, text)
}
