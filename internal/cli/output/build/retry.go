package build

import (
	"fmt"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderRetry(writer *core.Writer, result buildmodel.RetryResult) error {
	if result.Requested != nil {
		requested := result.Requested
		value := map[string]any{"operation": "build.retry", "source_build_id": requested.Build.ID, "application_id": requested.Build.ApplicationID, "commit": requested.Build.Commit, "state": "requested"}
		return writer.WriteValue(value, fmt.Sprintf("retry requested: build %s commit %s", requested.Build.ID, requested.Build.Commit))
	}
	if result.Completion != nil {
		return RenderCompletion(writer, *result.Completion)
	}
	return nil
}
