package cli

import (
	"fmt"

	appcli "github.com/traP-jp/neoshowcase-cli/internal/cli/app"
	buildcli "github.com/traP-jp/neoshowcase-cli/internal/cli/build"
	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	appmodel "github.com/traP-jp/neoshowcase-cli/internal/model/app"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type Renderer struct {
	output *clioutput.Renderer
}

func (r *Renderer) Render(event model.Event) error {
	switch value := event.(type) {
	case model.VersionResult:
		return renderVersion(r.output, value)
	case appmodel.ListResult:
		return appcli.RenderList(r.output, value)
	case appmodel.GetResult:
		return appcli.RenderGet(r.output, value)
	case appmodel.LogsResult:
		return appcli.RenderLogs(r.output, value)
	case appmodel.StartResult:
		return appcli.RenderStart(r.output, value)
	case appmodel.StopResult:
		return appcli.RenderStop(r.output, value)
	case appmodel.RestartResult:
		return appcli.RenderRestart(r.output, value)
	case appmodel.RebuildRequested:
		return appcli.RenderRebuild(r.output, value)
	case buildmodel.ListResult:
		return buildcli.RenderList(r.output, value)
	case buildmodel.GetResult:
		return buildcli.RenderGet(r.output, value)
	case buildmodel.LogResult:
		return buildcli.RenderLogs(r.output, value)
	case buildmodel.CompletionResult:
		return buildcli.RenderCompletion(r.output, value)
	case buildmodel.RetryRequested:
		return buildcli.RenderRetry(r.output, value)
	case buildmodel.CancelResult:
		return buildcli.RenderCancel(r.output, value)
	default:
		return fmt.Errorf("unsupported output event type %T", event)
	}
}
