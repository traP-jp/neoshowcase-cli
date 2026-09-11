package build

import (
	"fmt"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderCompletion(renderer *core.Writer, result buildmodel.CompletionResult) error {
	view := buildViewOf(result.Build)
	if result.Streaming && renderer.Format() == "text" {
		return renderer.Writef("build %s: %s\n", result.Build.ID, result.Build.Status)
	}
	if result.Streaming {
		return renderer.WriteJSONLine(view)
	}
	return renderer.WriteValue(view, fmt.Sprintf("build %s: %s", result.Build.ID, result.Build.Status))
}

func RenderWait(writer *core.Writer, result buildmodel.CompletionResult) error {
	return RenderCompletion(writer, result)
}
