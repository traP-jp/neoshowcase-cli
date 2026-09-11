package build

import (
	"fmt"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderGet(renderer *core.Writer, result buildmodel.GetResult) error {
	view := buildViewOf(result.Build)
	text := fmt.Sprintf("ID: %s\nApplication ID: %s\nCommit: %s\nStatus: %s\nRetriable: %t", view.ID, view.ApplicationID, view.Commit, view.Status, view.Retriable)
	return renderer.WriteValue(view, text)
}
