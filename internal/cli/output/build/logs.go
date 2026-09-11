package build

import (
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderLogs(renderer *core.Writer, result buildmodel.LogResult) error {
	if renderer.Format() == "text" {
		return renderer.WriteString(result.Text)
	}
	entry := core.Log{BuildID: result.BuildID, Text: result.Text}
	if !result.Streaming && renderer.Format() != "jsonl" {
		return renderer.WriteValue(entry, "")
	}
	return renderer.WriteLog(entry, result.Streaming)
}
