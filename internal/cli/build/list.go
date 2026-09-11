package build

import (
	"fmt"
	"strings"

	clioutput "github.com/traP-jp/neoshowcase-cli/internal/cli/output"
	"github.com/traP-jp/neoshowcase-cli/internal/model"
	buildmodel "github.com/traP-jp/neoshowcase-cli/internal/model/build"
)

type ListCommand struct {
	Application string `arg:"" optional:"" help:"Application ID or exact name"`
	Page        int32  `default:"0" help:"Zero-indexed page"`
	Limit       int32  `default:"20" help:"Number of builds in the page"`
}

func (command *ListCommand) Validate(ValidationContext) (model.Command, error) {
	if command.Page < 0 || command.Limit <= 0 {
		return nil, fmt.Errorf("--page must be non-negative and --limit must be positive")
	}
	return buildmodel.ListCommand{Application: command.Application, Page: command.Page, Limit: command.Limit}, nil
}

func RenderList(renderer *clioutput.Renderer, result buildmodel.ListResult) error {
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
