package build

import (
	"fmt"
	"strings"

	"github.com/traP-jp/neoshowcase-cli/internal/cli/output/core"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

func RenderList(renderer *core.Writer, result buildmodel.ListResult) error {
	views := make([]buildView, 0, len(result.Builds))
	var text strings.Builder
	fmt.Fprintln(&text, "ID\tAPPLICATION\tCOMMIT\tSTATUS")
	for _, build := range result.Builds {
		views = append(views, buildViewOf(build))
		application := build.ApplicationName
		if application == "" {
			application = build.ApplicationID
		}
		fmt.Fprintf(&text, "%s\t%s\t%s\t%s\n", build.ID, application, build.Commit, build.Status)
	}
	if renderer.Format() == "jsonl" {
		for _, view := range views {
			if err := renderer.WriteJSONLine(view); err != nil {
				return err
			}
		}
		return nil
	}
	return renderer.WriteValue(views, strings.TrimSuffix(text.String(), "\n"))
}
