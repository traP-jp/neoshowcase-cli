package build

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderWatch(writer *core.Writer, result buildmodel.CompletionResult) error {
	return RenderCompletion(writer, result)
}
