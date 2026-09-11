package app

import (
	"fmt"

	buildoutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output/build"
	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
)

func RenderRebuild(writer *core.Writer, result appmodel.RebuildResult) error {
	if result.Requested != nil {
		requested := result.Requested
		value := map[string]any{"operation": "app.rebuild", "application_id": requested.Application.ID, "application_name": requested.Application.Name, "commit": requested.Commit, "state": "requested"}
		return writer.WriteValue(value, fmt.Sprintf("rebuild requested: %s (%s) commit %s", requested.Application.Name, requested.Application.ID, requested.Commit))
	}
	if result.Completion != nil {
		return buildoutput.RenderCompletion(writer, *result.Completion)
	}
	return nil
}
